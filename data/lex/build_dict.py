#!/usr/bin/env python3
"""Build app/internal/dict/dict.bin from the sources in this directory.

Sources (all fetched into data/lex/):
  freq-en.txt          FrequencyWords en_50k — decides what "basic vocabulary" is
  package/dict/*       WordNet 3.0 database (wordnet-db 3.1.14, Princeton licence)
  cmudict.dict         CMU Pronouncing Dictionary — IPA
  cc-cedict*/          CC-CEDICT — Chinese glosses (inverted)
  open-dict.jsonl.gz   Open Dictionary v2 — Chinese learner glosses and word forms
  acai-names.json      ACAI preferred labels — Chinese biblical proper names
  ../db/bible-en-study.db   BSB + WEBP + KJV + CUV — word forms, collocations, counts

Output format, same shape as the Bible store:
  "LEXI2\n" | uint32 compressed index length | zlib(index JSON) | payload
  index  = {"words": [...], "forms": {form: lemmaId}, "entries": {word: [off,len]}}
  entry  = zlib(JSON {"w","pos","uk","s":[...],"zh":[...],"c":[...],"n","r","lemma"})

Usage:
  python3 build_dict.py [-vocab all|bible|top10k|top33k]
"""

import argparse
import gzip
import json
import math
import os
import re
import sqlite3
import struct
import sys
import zlib
from collections import Counter, defaultdict

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(os.path.dirname(HERE))
OUT = os.path.join(ROOT, "app", "internal", "dict", "dict.bin")

WORD = re.compile(r"^[a-z][a-z'-]*$")
OPEN_WORD = re.compile(r"^[a-z][a-z']*$")

# Open Dictionary uses Wiktionary's full part-of-speech names. Only one badge
# fits on the card, so merge the groups and pick the one a reader is most
# likely to want.
OPEN_POS = {
    "noun": "n.", "verb": "v.", "adj": "adj.", "adv": "adv.",
    "prep": "prep.", "pron": "pron.", "det": "det.", "num": "num.",
    "intj": "intj.", "name": "专名",
}
OPEN_POS_WEIGHT = {
    "noun": 3, "verb": 3, "adj": 2.5, "adv": 2.2, "prep": 2,
    "pron": 2, "det": 2, "num": 2, "intj": 1, "name": 0.6,
}
OPEN_FORM_TAGS = {
    "plural", "singular", "past", "participle", "present",
    "third-person", "comparative", "superlative", "archaic",
}

# These forms are homographs of ordinary nouns in English ("saw" the tool,
# "rose" the flower). In this reader the inflected reading is the useful one,
# so the pointer entry wins over the headword entry.
INFLECTION_FIRST = {
    "saw", "rose", "left", "found", "spake", "dwelt", "smote", "arose",
    # Forms of be/have/do and the common irregular verbs. Open Dictionary also
    # lists them as headwords, but their useful reading is the verb they belong
    # to; a separate "is" entry would only show the IS acronym.
    "am", "is", "are", "was", "were", "been", "being",
    "has", "had", "having", "does", "did", "done", "doing",
    "took", "taken", "came", "went", "gone", "said", "made",
    "gave", "given", "told", "knew", "known", "got", "gotten", "seen",
}

# ---------------------------------------------------------------- ARPAbet → IPA

ARPABET = {
    "AA": "ɑ", "AE": "æ", "AH": "ʌ", "AO": "ɔ", "AW": "aʊ", "AY": "aɪ",
    "B": "b", "CH": "tʃ", "D": "d", "DH": "ð", "EH": "ɛ", "ER": "ɝ",
    "EY": "eɪ", "F": "f", "G": "ɡ", "HH": "h", "IH": "ɪ", "IY": "i",
    "JH": "dʒ", "K": "k", "L": "l", "M": "m", "N": "n", "NG": "ŋ",
    "OW": "oʊ", "OY": "ɔɪ", "P": "p", "R": "ɹ", "S": "s", "SH": "ʃ",
    "T": "t", "TH": "θ", "UH": "ʊ", "UW": "u", "V": "v", "W": "w",
    "Y": "j", "Z": "z", "ZH": "ʒ",
}
VOWELS = {"AA", "AE", "AH", "AO", "AW", "AY", "EH", "ER", "EY", "IH", "IY",
          "OW", "OY", "UH", "UW"}
# Stressed tense vowels are long in English IPA. Without the length mark,
# "God" comes out as /ɡɑd/, which is the spelling of a different vowel.
LONG = {"AA", "AO", "ER", "IY", "UW"}


def arpabet_to_ipa(phones):
    segs = []          # (ipa, stress) in order
    for ph in phones:
        base = ph.rstrip("012")
        stress = ph[-1] if ph[-1].isdigit() else ""
        ipa = ARPABET.get(base)
        if ipa is None:
            continue
        if base == "AH" and stress == "0":
            ipa = "ə"
        elif base == "ER" and stress == "0":
            ipa = "ɚ"
        elif base in LONG and stress in ("1", "2"):
            ipa += "ː"
        segs.append((ipa, stress if base in VOWELS else ""))
    vowels = [i for i, (_, st) in enumerate(segs) if st != ""]
    # A word of one syllable carries no stress mark: /laɪt/, not /lˈaɪt/.
    marks = {}
    if len(vowels) > 1:
        for i in vowels:
            st = segs[i][1]
            if st == "1":
                mark = "ˈ"
            elif st == "2":
                mark = "ˌ"
            else:
                continue
            # the mark belongs at the onset of the syllable, before its first
            # consonant, not in front of the vowel
            start = i
            while start > 0 and segs[start - 1][1] == "":
                start -= 1
            marks[start] = mark
    out = []
    for i, (ipa, _) in enumerate(segs):
        out.append(marks.get(i, "") + ipa)
    return "/" + "".join(out) + "/"


# ---------------------------------------------------------------- word forms

IRREGULAR = {
    "be": ["am", "is", "are", "was", "were", "been", "being"],
    "beget": ["begat", "begotten", "begets", "begetting"],
    "have": ["has", "had", "having"],
    "do": ["does", "did", "done", "doing"],
    "go": ["goes", "went", "gone", "going"],
    "say": ["said", "says", "saying"],
    "make": ["made", "makes", "making"],
    "take": ["took", "taken", "takes", "taking"],
    "come": ["came", "comes", "coming"],
    "see": ["saw", "seen", "sees", "seeing"],
    "know": ["knew", "known", "knows", "knowing"],
    "give": ["gave", "given", "gives", "giving"],
    "think": ["thought", "thinks", "thinking"],
    "find": ["found", "finds", "finding"],
    "tell": ["told", "tells", "telling"],
    "become": ["became", "becomes", "becoming"],
    "leave": ["left", "leaves", "leaving"],
    "feel": ["felt", "feels", "feeling"],
    "bring": ["brought", "brings", "bringing"],
    "begin": ["began", "begun", "begins", "beginning"],
    "keep": ["kept", "keeps", "keeping"],
    "hold": ["held", "holds", "holding"],
    "write": ["wrote", "written", "writes", "writing"],
    "stand": ["stood", "stands", "standing"],
    "hear": ["heard", "hears", "hearing"],
    "let": ["lets", "letting"],
    "mean": ["meant", "means", "meaning"],
    "set": ["sets", "setting"],
    "meet": ["met", "meets", "meeting"],
    "run": ["ran", "runs", "running"],
    "pay": ["paid", "pays", "paying"],
    "sit": ["sat", "sits", "sitting"],
    "speak": ["spoke", "spoken", "speaks", "speaking"],
    "lie": ["lay", "lain", "lies", "lying"],
    "lead": ["led", "leads", "leading"],
    "grow": ["grew", "grown", "grows", "growing"],
    "lose": ["lost", "loses", "losing"],
    "fall": ["fell", "fallen", "falls", "falling"],
    "send": ["sent", "sends", "sending"],
    "build": ["built", "builds", "building"],
    "understand": ["understood", "understands", "understanding"],
    "draw": ["drew", "drawn", "draws", "drawing"],
    "break": ["broke", "broken", "breaks", "breaking"],
    "spend": ["spent", "spends", "spending"],
    "drive": ["drove", "driven", "drives", "driving"],
    "bear": ["bore", "born", "borne", "bears", "bearing"],
    "eat": ["ate", "eaten", "eats", "eating"],
    "teach": ["taught", "teaches", "teaching"],
    "seek": ["sought", "seeks", "seeking"],
    "choose": ["chose", "chosen", "chooses", "choosing"],
    "forgive": ["forgave", "forgiven", "forgives", "forgiving"],
    "rise": ["rose", "risen", "rises", "rising"],
    "die": ["died", "dies", "dying"],
    "live": ["lived", "lives", "living"],
    "man": ["men"],
    "woman": ["women"],
    "child": ["children"],
    "foot": ["feet"],
    "tooth": ["teeth"],
    "person": ["people"],
    "sheep": ["sheep"],
    "ox": ["oxen"],
    "get": ["gets", "got", "gotten", "getting"],
}

# form → lemma. The reader sees "went" far more often than "go", so these
# cannot be left to the suffix rules.
IRREGULAR_FORM = {form: lemma for lemma, forms in IRREGULAR.items() for form in forms}


# ---------------------------------------------------------------- sources


def open_group_score(group):
    """How likely a part-of-speech group is the one the reader wants.

    Core senses outrank common ones; rare senses are mostly technical. Names
    and interjections lose to nouns and verbs unless nothing else exists.
    """
    counts = Counter(m.get("priority") or "rare" for m in group.get("meanings") or [])
    return (counts["core"] * 4 + counts["common"] * 1.5
            + counts["rare"] * 0.15 + OPEN_POS_WEIGHT.get(group.get("pos"), 1))


def load_open_dict():
    """Open Dictionary: single-word headwords, Chinese glosses and forms.

    The release has 84,212 entries, but 50,248 of them are phrases, symbols or
    hyphenated compounds that this reader can never click (the tokenizer emits
    one alphabetic token at a time). Keeping those would grow the index without
    changing a single lookup, so only single alphabetic headwords are used.
    """
    path = os.environ.get("OPEN_DICT", os.path.join(HERE, "open-dict.jsonl.gz"))
    out, form_sources = {}, defaultdict(Counter)
    if not os.path.exists(path):
        return out, {}
    with gzip.open(path, "rt", encoding="utf-8") as fh:
        for line in fh:
            try:
                row = json.loads(line)
            except json.JSONDecodeError:
                continue
            word = (row.get("normalized_headword") or row.get("headword") or "").strip().lower()
            if not OPEN_WORD.fullmatch(word):
                continue
            groups = [g for g in row.get("pos_groups") or [] if g.get("meanings")]
            if not groups:
                continue
            proper = row.get("entry_type") == "proper_name"
            if proper:
                groups = [g for g in groups if g.get("pos") == "name"] or groups
            else:
                # A proper-name group on a common word ("Church", "Temple")
                # describes a surname or institution, not the word in the text.
                groups = [g for g in groups if g.get("pos") not in ("name", "intj")] or groups
            group = max(groups, key=open_group_score)
            glosses = []
            for meaning in sorted(group.get("meanings") or [],
                                  key=lambda m: {"core": 0, "common": 1}.get(m.get("priority"), 2)):
                gloss = (meaning.get("short_gloss") or "").strip()
                if gloss and gloss not in glosses:
                    glosses.append(gloss)
                if len(glosses) == 3:
                    break
            ipa = next((p.get("ipa") for p in group.get("pronunciations") or []
                        if p.get("ipa")), "")
            out.setdefault(word, {
                "pos": OPEN_POS.get(group.get("pos"), ""),
                "ipa": ipa,
                "zh": glosses,
                "proper": proper,
            })
            # Forms belong to the headword, not to the one group we picked
            # for the card: "sin" is glossed as a noun but its verb supplies
            # "sinned".
            for group in row.get("pos_groups") or []:
                for form in group.get("forms") or []:
                    text = (form.get("text") or "").strip().lower()
                    tags = set(form.get("tags") or [])
                    if (text != word and OPEN_WORD.fullmatch(text)
                            and tags & OPEN_FORM_TAGS):
                        form_sources[text][word] += 1
    # One form can be listed under more than one headword ("sinned" under both
    # sin and the dialectal sinn). Prefer a headword we actually kept, then the
    # shorter base, then the most frequent listing.
    forms = {}
    for form, bases in form_sources.items():
        held = [b for b in bases if b in out]
        base = min(held or list(bases), key=lambda b: (len(b), -bases[b], b))
        forms[form] = base
    return out, forms


def load_acai_names():
    """ACAI's preferred English → Chinese labels, for biblical proper names.

    Only the preferred label is used. Alternate labels routinely belong to a
    different entity, and ACAI has several people named Jehoshaphat; the card
    only needs a short Chinese name, so the first preferred label wins.
    """
    path = os.path.join(HERE, "acai-names.json")
    if not os.path.exists(path):
        return {}
    with open(path, encoding="utf-8") as fh:
        names = json.load(fh)
    return {key.lower(): value for key, value in names.items()
            if OPEN_WORD.fullmatch(key.lower()) and value}


def load_frequency():
    rank = {}
    with open(os.path.join(HERE, "freq-en.txt"), encoding="utf-8") as fh:
        for i, line in enumerate(fh):
            parts = line.split()
            if len(parts) == 2:
                rank.setdefault(parts[0].lower(), i + 1)
    return rank


def load_bible():
    """Every English word form, and BSB statistics for display.

    KJV and WEBP are tokenized too: "hath", "saith" and "dwelleth" only occur
    there, and a reader who clicks them must still get a card. Counts and
    collocations stay on BSB so the numbers keep one meaning.
    """
    db = sqlite3.connect(f"file:{os.path.join(ROOT, 'data', 'db', 'bible-en-study.db')}?mode=ro", uri=True)
    token = re.compile(r"[A-Za-z]+(?:'[A-Za-z]+)?")
    counts = Counter()
    seen = set()
    bigrams = defaultdict(Counter)
    previous = defaultdict(Counter)
    cuv = []
    for translation, text in db.execute(
            "select translation_id, text from verses "
            "where translation_id in ('BSB','WEBP','KJV')"):
        words = [w.lower() for w in token.findall(text)]
        seen.update(words)
        if translation != "BSB":
            continue
        counts.update(words)
        for a, b in zip(words, words[1:]):
            bigrams[a][b] += 1
            previous[b][a] += 1
    # CUV is the only Chinese text we have; it ranks Chinese glosses by how
    # often the translation actually uses them.
    for text in db.execute("select text from verses where translation_id='CUV'"):
        cuv.append(text[0])
    return counts, seen, bigrams, previous, cuv


def load_cedict():
    """CC-CEDICT, inverted: English gloss → Chinese headword.

    Whole gloss chunks only ("to create" ⇒ create), split by the kind of word
    the chunk describes so a noun entry is not glossed with a verb. Names are
    kept apart: they are the only Chinese we have for biblical proper nouns.
    """
    out = {"n": defaultdict(Counter), "v": defaultdict(Counter),
           "name": defaultdict(Counter)}
    candidates = set()
    tgz = os.path.join(HERE, "cc-cedict-1.1.1.tgz")
    js = os.path.join(HERE, "cedict.js")
    if not os.path.exists(js):
        if not os.path.exists(tgz):
            return out, candidates
        import tarfile
        with tarfile.open(tgz) as tf:
            open(js, "wb").write(tf.extractfile("package/data/all.js").read())
    raw = open(js, encoding="utf-8").read().split("export default", 1)[1].strip().rstrip(";")
    for row in json.loads(raw)["all"]:
        simp, english = row[1] or row[0], row[3]
        if isinstance(english, list):
            english = "; ".join(english)
        # Single-character words (神, 主, 罪, 灵) are the most useful glosses
        # for biblical terms and were being dropped by the old 2..4 filter.
        if not isinstance(english, str) or not re.fullmatch(r"[\u4e00-\u9fff]{1,4}", simp):
            continue
        candidates.add(simp)
        for i, chunk in enumerate(english.split(";")):
            chunk = re.sub(r"\(.*?\)", "", chunk).strip()
            verb = chunk.lower().startswith("to ")
            chunk = re.sub(r"^(to|a|an|the)\s+", "", chunk, flags=re.I).strip()
            if not chunk:
                continue
            weight = 3 if i == 0 else 1
            if len(simp) <= 3:
                weight += 1
            if chunk[:1].isupper() and re.fullmatch(r"[A-Za-z][A-Za-z .'-]*", chunk):
                key = chunk.split(",")[0].strip().lower()
                if len(key) > 1:
                    out["name"][key][simp] += weight
                continue
            if not re.fullmatch(r"[a-z][a-z-]*", chunk) or len(chunk) < 2:
                continue
            out["v" if verb else "n"][chunk][simp] += weight
    return out, candidates


def load_cuv_freq(verses, candidates):
    """How often each candidate Chinese word occurs in the CUV.

    A gloss that the Chinese Bible actually uses is a better translation for
    this reader than one that merely wins CC-CEDICT's alphabetical order.
    """
    freq = Counter()
    if not candidates:
        return freq
    cjk = re.compile(r"[\u4e00-\u9fff]+")
    for text in verses:
        for run in cjk.findall(text):
            for i in range(len(run)):
                for n in (1, 2, 3, 4):
                    if i + n > len(run):
                        break
                    word = run[i:i + n]
                    if word in candidates:
                        freq[word] += 1
    return freq


def load_wordnet():
    """lemma → [(pos, gloss, examples)] from the WordNet 3.0 database."""
    pos_of = {"noun": "n", "verb": "v", "adj": "a", "adv": "r"}
    index, data = {}, {}
    for name, pos in pos_of.items():
        with open(os.path.join(HERE, "package", "dict", f"index.{name}"), encoding="latin-1") as fh:
            for line in fh:
                if line.startswith("  ") or not line.strip():
                    continue
                parts = line.split()
                lemma = parts[0].replace("_", " ").lower()
                p_cnt = int(parts[3])
                offsets = parts[6 + p_cnt:]
                index.setdefault(lemma, []).extend((pos, int(o)) for o in offsets)
        with open(os.path.join(HERE, "package", "dict", f"data.{name}"), encoding="latin-1") as fh:
            for line in fh:
                if line.startswith("  ") or not line.strip():
                    continue
                head, _, gloss = line.partition("|")
                parts = head.split()
                offset = int(parts[0])
                data[(pos, offset)] = gloss.strip()
    return index, data


def load_cmudict():
    out = {}
    path = os.path.join(HERE, "cmudict.dict")
    if not os.path.exists(path):
        return out
    with open(path, encoding="utf-8", errors="ignore") as fh:
        for line in fh:
            parts = line.split()
            if len(parts) < 2 or parts[0].startswith(";;;"):
                continue
            word = parts[0].split("(")[0].lower()
            if not WORD.match(word) or word in out:
                continue
            out[word] = arpabet_to_ipa(parts[1:])
    return out


# ---------------------------------------------------------------- entry

# Words whose meaning is grammar, not a WordNet sense. WordNet either has no
# entry for them or offers the wrong homonym ("a" as a letter, "in" as an
# adjective), which is how they ended up missing or glossed as metric units.
FUNCTION = {
    "the": ("art.", "定冠词（特指）", "used before a noun to refer to a particular person or thing"),
    "a": ("art.", "不定冠词（一个）", "used before a singular noun to refer to one of a class"),
    "an": ("art.", "不定冠词（一个）", "used before a vowel sound to refer to one of a class"),
    "and": ("conj.", "和；与；而且", "used to join words, phrases, or clauses"),
    "or": ("conj.", "或者；否则", "used to connect alternatives"),
    "nor": ("conj.", "也不；也没有", "used to introduce a further negative statement"),
    "but": ("conj.", "但是；除了", "used to introduce a contrast"),
    "of": ("prep.", "……的；属于", "belonging to, relating to, or coming from"),
    "to": ("prep.", "到；向；朝；（不定式标志）", "in the direction of; used before a verb in the infinitive"),
    "in": ("prep.", "在……里；在……期间", "inside something; during a period"),
    "on": ("prep.", "在……上；关于", "touching the surface of; about"),
    "at": ("prep.", "在（某处/某时）", "in or near a place or time"),
    "by": ("prep.", "被；由；通过；在……旁", "near; through the action of"),
    "from": ("prep.", "从；来自", "starting at; originating in"),
    "with": ("prep.", "和……一起；用", "together with; using"),
    "without": ("prep.", "没有；不用", "not having; not using"),
    "into": ("prep.", "进入；变成", "to the inside of; becoming"),
    "unto": ("prep.", "到；向（古）", "to; toward (archaic)"),
    "upon": ("prep.", "在……上；当……时（古）", "on; at the time of (archaic)"),
    "through": ("prep.", "通过；穿过", "from one side to the other; by means of"),
    "as": ("conj.", "作为；当……时；因为", "in the role of; while; because"),
    "than": ("conj.", "比", "used to compare two things"),
    "that": ("pron.", "那；那个；（引导从句）", "used to introduce a clause; the thing over there"),
    "this": ("pron.", "这；这个", "the person or thing near the speaker"),
    "these": ("pron.", "这些", "plural of this"),
    "those": ("pron.", "那些", "plural of that"),
    "there": ("adv.", "那里；（there be）有", "in that place; used in there is/are"),
    "then": ("adv.", "然后；那时", "at that time; next"),
    "so": ("adv.", "因此；如此", "for that reason; to such a degree"),
    "if": ("conj.", "如果；是否", "used to introduce a condition"),
    "because": ("conj.", "因为", "for the reason that"),
    "when": ("conj.", "当……时；什么时候", "at the time that; at what time"),
    "while": ("conj.", "当……时；然而", "during the time that; although"),
    "though": ("conj.", "虽然；尽管", "despite the fact that"),
    "although": ("conj.", "虽然；尽管", "despite the fact that"),
    "since": ("conj.", "自从；因为", "from a past time; because"),
    "until": ("conj.", "直到", "up to the time that"),
    "whether": ("conj.", "是否", "used to introduce a choice between possibilities"),
    "not": ("adv.", "不；没有", "used to make a negative"),
    "no": ("adv.", "不；没有；禁止", "used to give a negative answer"),
    "never": ("adv.", "从不；绝不", "not at any time"),
    "be": ("v.", "是；存在", "to exist or to have a quality"),
    "have": ("v.", "有；已经（助动词）", "to own or hold; used to form perfect tenses"),
    "do": ("v.", "做；（助动词）", "to perform; used to form questions and negatives"),
    "i": ("pron.", "我", "the speaker or writer"),
    "you": ("pron.", "你；你们", "the person or people spoken to"),
    "he": ("pron.", "他", "the male person spoken about"),
    "she": ("pron.", "她", "the female person spoken about"),
    "it": ("pron.", "它；这", "the thing spoken about"),
    "we": ("pron.", "我们", "the speaker and others"),
    "they": ("pron.", "他们；她们；它们", "the people or things spoken about"),
    "me": ("pron.", "我（宾格）", "the object form of I"),
    "him": ("pron.", "他（宾格）", "the object form of he"),
    "her": ("pron.", "她；她的", "the object form of she; belonging to her"),
    "us": ("pron.", "我们（宾格）", "the object form of we"),
    "them": ("pron.", "他们（宾格）", "the object form of they"),
    "my": ("det.", "我的", "belonging to me"),
    "your": ("det.", "你的；你们的", "belonging to you"),
    "his": ("det.", "他的", "belonging to him"),
    "its": ("det.", "它的", "belonging to it"),
    "our": ("det.", "我们的", "belonging to us"),
    "their": ("det.", "他们的", "belonging to them"),
    "who": ("pron.", "谁（关系代词）", "which person; used to introduce a relative clause"),
    "whom": ("pron.", "谁（宾格）", "the object form of who"),
    "whose": ("pron.", "谁的", "belonging to which person"),
    "which": ("pron.", "哪个；……的（关系代词）", "used to ask or introduce a choice"),
    "what": ("pron.", "什么", "used to ask about a thing"),
    "will": ("modal", "将；会；意愿", "used to talk about the future; a desire or wish"),
    "would": ("modal", "将会；愿意", "used in conditional statements; past of will"),
    "shall": ("modal", "将；应当", "used to talk about the future or an obligation"),
    "should": ("modal", "应该", "used to give advice or say what is right"),
    "can": ("modal", "能；可以", "to be able to; to be allowed to"),
    "could": ("modal", "能；可以（过去式）", "past of can; used for possibility"),
    "may": ("modal", "可以；也许", "to be allowed to; to be possible"),
    "might": ("modal", "也许；可以；力量", "used for possibility; power or strength"),
    "must": ("modal", "必须", "used to say something is necessary"),
    "myself": ("pron.", "我自己", "used to refer to the speaker as the object of a verb"),
    "yourself": ("pron.", "你自己", "used to refer to the person spoken to"),
    "himself": ("pron.", "他自己", "used to refer to a male person already mentioned"),
    "herself": ("pron.", "她自己", "used to refer to a female person already mentioned"),
    "itself": ("pron.", "它自己", "used to refer to a thing already mentioned"),
    "ourselves": ("pron.", "我们自己", "used to refer to the speaker and others"),
    "yourselves": ("pron.", "你们自己", "used to refer to the people spoken to"),
    "themselves": ("pron.", "他们自己", "used to refer to people or things already mentioned"),
    "oneself": ("pron.", "自己", "used to refer to a person in general"),
    "one": ("num./pron.", "一个；人", "a single person or thing"),
    "well": ("adv./n.", "好；健康；井", "in a good way; a deep hole dug for water"),
    "still": ("adv.", "仍然；还", "continuing up to now"),
    "even": ("adv.", "甚至；连", "used to emphasize something surprising"),
    "very": ("adv.", "很；非常", "to a high degree"),
    "own": ("adj./v.", "自己的；拥有", "belonging to oneself; to possess"),
    "why": ("adv.", "为什么；原因", "for what reason"),
    "how": ("adv.", "怎样；如何", "in what way"),
    "where": ("adv.", "哪里；在哪儿", "in or to what place"),
    "yes": ("int.", "是；对", "used to give an affirmative answer"),
    "oh": ("int.", "哦；啊", "used to express surprise or emotion"),
    "ah": ("int.", "啊", "used to express surprise, pleasure, or sympathy"),
    "thou": ("pron.", "你（古）", "you, singular (archaic)"),
    "thee": ("pron.", "你（宾格，古）", "the object form of thou (archaic)"),
    "thy": ("det.", "你的（古）", "belonging to thou (archaic)"),
    "thine": ("det.", "你的（古，元音前）", "belonging to thou (archaic)"),
    "ye": ("pron.", "你们（古）", "you, plural (archaic)"),
    "behold": ("v.", "看哪；注意看（古）", "look; see (archaic)"),
    "amen": ("int.", "阿们", "so be it; truly"),
}

# KJV forms that no suffix rule reaches.
ARCHAIC_FORM = {
    "hath": "have", "hast": "have", "doth": "do", "dost": "do", "didst": "do",
    "saith": "say", "sayest": "say", "art": "be", "wast": "be", "wert": "be",
    "shalt": "shall", "wilt": "will", "canst": "can", "couldst": "could",
    "wouldst": "would", "shouldst": "should", "mayest": "may", "mightest": "might",
    "shouldest": "should", "wouldest": "would", "couldest": "could",
}
IRREGULAR_FORM.update(ARCHAIC_FORM)

# A short curated set where the Chinese Bible's own wording matters more than
# the general dictionary order.
ZH_OVERRIDE = {
    "god": ["神", "上帝"],
    "lord": ["主", "上帝"],
    "grace": ["恩典", "恩惠"],
    "temple": ["圣殿", "殿", "庙宇"],
    "spirit": ["灵", "灵魂", "精神"],
    "heaven": ["天", "天堂", "天空"],
    "church": ["教会", "教堂"],
    "apostle": ["使徒"],
    "abram": ["亚伯兰"],
    "christ": ["基督"],
    "satan": ["撒但", "魔鬼"],
    "holy": ["神圣的", "圣洁的"],
    "righteous": ["公义", "正义", "正直"],
    "flock": ["羊群", "群羊"],
    "fool": ["愚人", "愚昧人", "愚妄人"],
    "seed": ["种子", "后裔"],
    "wealth": ["财富", "财宝"],
    "wise": ["智慧", "聪明"],
    "brother": ["兄弟", "弟兄"],
    "sister": ["姐妹", "姊妹"],
    "father": ["父亲", "父"],
    "mother": ["母亲", "母"],
    "judgment": ["审判", "判断"],
    "levite": ["利未人"],
    "iniquity": ["罪孽", "不义"],
    "cubit": ["肘（长度单位）"],
    "chaldean": ["迦勒底人"],
    "shekel": ["舍客勒"],
    "defiled": ["污秽的", "玷污"],
    "unleavened": ["无酵的"],
    "laban": ["拉班"],
    "ephah": ["伊法"],
    "israelites": ["以色列人"],
    "afterward": ["后来", "此后"],
    "fatherless": ["无父的", "孤儿"],
    "selah": ["细拉"],
    "sheol": ["阴间"],
    "cherub": ["基路伯"],
    "cherubim": ["基路伯"],
    "uncircumcised": ["未受割礼的"],
    "wherewith": ["用……；借以"],
    "howbeit": ["然而"],
    "shekels": ["舍客勒"],
    "sanctified": ["分别为圣", "使成圣"],
    "thither": ["到那里"],
    "threescore": ["六十"],
    "uncleanness": ["不洁"],
    "threshing": ["打谷"],
    "hittite": ["赫人"],
    "elam": ["以拦"],
    "exalt": ["尊崇", "升高"],
    "afflict": ["使受苦", "苦待"],
    "consecrated": ["分别为圣的"],
    "wickedly": ["邪恶地"],
    "abimelech": ["亚比米勒"],
    "gibeah": ["基比亚"],
    "gilgal": ["吉甲"],
    "ramah": ["拉玛"],
    "mizpah": ["米斯巴"],
    "kadesh": ["加低斯"],
    "adonijah": ["亚多尼雅"],
    "beersheba": ["别是巴"],
    "reubenites": ["流便人"],
    "beget": ["生", "生育"],
    "encamp": ["安营"],
    "maidservant": ["婢女", "使女"],
    "besiege": ["围困"],
    "canaanite": ["迦南人"],
    "succoth": ["疏割"],
    "lamentation": ["哀歌", "哀哭"],
    "raiment": ["衣服"],
    "unrighteousness": ["不义"],
    "sepulchre": ["坟墓"],
    "ephesus": ["以弗所"],
    "travail": ["劳苦", "产痛"],
    "harlot": ["妓女"],
    "uprightness": ["正直"],
    "wroth": ["发怒的"],
    "ramoth": ["拉末"],
    "shemesh": ["示麦"],
    "kiriath": ["基列"],
    "jearim": ["耶琳"],
    "libnah": ["立拿"],
    "whoso": ["凡……的人"],
    "benjamites": ["便雅悯人"],
    "horeb": ["何烈"],
    "sarai": ["撒莱"],
    "facedown": ["俯伏"],
    "amorite": ["亚摩利人"],
    "aroer": ["亚罗珥"],
    "hermon": ["黑门"],
    "shadrach": ["沙得拉"],
    "meshach": ["米煞"],
    "firstfruits": ["初熟之物"],
    "cherubims": ["基路伯"],
    "jehoash": ["约阿施"],
    "whereof": ["其中", "关于什么"],
    "forasmuch": ["既然", "因为"],
    "besieged": ["被围困的"],
    "tidings": ["消息", "音信"],
    "godliness": ["敬虔"],
    "handmaid": ["婢女", "使女"],
    "negev": ["南地"],
    "shittim": ["什亭"],
    "unblemished": ["无残疾的", "无瑕疵的"],
    "leavened": ["有酵的"],
    "kinsman": ["亲属", "至亲"],
    "susa": ["书珊"],
    "treacherously": ["诡诈地"],
    "straightway": ["立刻"],
    "sin": ["罪", "罪恶"],
    "faith": ["信心", "信仰"],
    "mercy": ["怜悯", "慈悲"],
    "covenant": ["约", "盟约"],
    "prophet": ["先知"],
    "disciple": ["门徒"],
    "angel": ["天使"],
    "gospel": ["福音"],
    "righteousness": ["义", "公义"],
    "salvation": ["救恩", "拯救"],
    "repent": ["悔改"],
    "baptize": ["施洗"],
    "passover": ["逾越节"],
    "sabbath": ["安息日"],
    "tabernacle": ["帐幕"],
    "altar": ["祭坛"],
    "priest": ["祭司"],
    "sacrifice": ["祭物", "牺牲"],
    "resurrection": ["复活"],
    "people": ["人", "人们", "百姓"],
    "something": ["某事", "某物"],
    "anything": ["任何事", "任何东西"],
    "nothing": ["没有什么", "无事"],
    "everything": ["一切", "每件事"],
    "everyone": ["每个人", "众人"],
    "someone": ["有人", "某人"],
    "anyone": ["任何人"],
    "world": ["世界"],
    "earth": ["地", "大地", "世界"],
    "king": ["王", "国王"],
    "kingdom": ["国", "国度", "王国"],
    "city": ["城", "城市"],
    "house": ["房屋", "家"],
    "day": ["日", "日子", "白天"],
    "night": ["夜", "夜晚"],
    "water": ["水"],
    "bread": ["饼", "面包"],
    "wine": ["酒", "葡萄酒"],
    "flesh": ["肉体", "血肉"],
    "soul": ["灵魂", "生命"],
    "heart": ["心", "内心"],
    "word": ["话", "道", "言语"],
    "law": ["律法", "法律"],
    "truth": ["真理", "真相"],
    "life": ["生命", "生活"],
    "death": ["死", "死亡"],
    "light": ["光", "光明"],
    "darkness": ["黑暗"],
    "peace": ["平安", "和平"],
    "love": ["爱", "慈爱"],
    "hope": ["盼望", "希望"],
    "fear": ["惧怕", "敬畏"],
    "wisdom": ["智慧"],
    "power": ["能力", "权柄", "力量"],
    "glory": ["荣耀"],
    "blessed": ["有福的", "蒙福的"],
    "pray": ["祷告", "祈祷"],
    "worship": ["敬拜", "崇拜"],
    "save": ["拯救", "救"],
    "heal": ["医治", "治愈"],
    "clean": ["洁净", "干净"],
    "unclean": ["不洁净"],
}

SENSE_OVERRIDE = {
    "god": ["the supreme being; the Creator and ruler of the universe"],
    "lord": ["God; a master or ruler", "a titled nobleman"],
    "jesus": ["the central figure of the New Testament; Christians believe he is the Son of God"],
    "christ": ["the Messiah; the anointed one"],
    "heaven": ["the sky; the place where God is said to live"],
    "spirit": ["the non-physical part of a person; the Holy Spirit"],
    "grace": ["God's free and unearned favor"],
    "church": ["the community of Christians; a building for Christian worship"],
    "temple": ["a building used for worship, especially in Jerusalem"],
    "satan": ["the chief spirit of evil; the adversary of God"],
    "holy": ["dedicated to God; sacred; morally pure"],
    "righteous": ["morally right; just; conforming to God's law"],
    "flock": ["a group of sheep or birds; a congregation"],
    "fool": ["a person who lacks good sense; one who denies God"],
    "seed": ["the part of a plant from which a new plant grows; offspring"],
    "wealth": ["a large amount of money or possessions"],
    "wise": ["having or showing good judgment and knowledge"],
    "brother": ["a male sibling; a fellow believer"],
    "sister": ["a female sibling; a fellow believer"],
    "father": ["a male parent; God as Father"],
    "mother": ["a female parent"],
    "judgment": ["the act of judging; God's judgment"],
}

# A few common words whose WordNet part of speech is not the one the Bible
# reader meets first.
POS_OVERRIDE = {
    "holy": "adj.", "righteous": "adj.", "flock": "n.", "fool": "n.",
    "seed": "n.", "wealth": "n.", "wise": "adj.", "brother": "n.",
    "sister": "n.", "father": "n.", "mother": "n.", "judgment": "n.",
}

# CMUdict gives the strong form. These are the weak forms a reader actually
# hears in running text.
IPA_OVERRIDE = {
    "the": "/ðə/", "a": "/ə/", "an": "/ən/", "and": "/ənd/", "of": "/əv/",
    "to": "/tə/", "was": "/wʌz/", "were": "/wɜːr/", "been": "/biːn/",
    "are": "/ɑːr/",
}

STOP = set(FUNCTION) | {
    "has", "had", "having", "does", "did", "done", "doing", "is", "are",
    "was", "were", "been", "being", "am", "hath", "doth", "saith", "shalt",
    "wilt", "art", "ye", "thee", "thou", "thy", "thine",
}


def collocations(word, forward, backward, limit=2):
    """Bigrams worth showing: the partner carries meaning, not just grammar."""
    if word in FUNCTION:
        return []
    picked = []
    for partner, n in forward.get(word, Counter()).most_common(30):
        if len(partner) > 2 and partner not in STOP and n > 1:
            picked.append((n, f"{word} {partner}"))
    for partner, n in backward.get(word, Counter()).most_common(30):
        if len(partner) > 2 and partner not in STOP and n > 1:
            picked.append((n, f"{partner} {word}"))
    picked.sort(key=lambda pair: (-pair[0], pair[1]))
    out = []
    for _, phrase in picked:
        if phrase not in out:
            out.append(phrase)
        if len(out) == limit:
            break
    return out


def choose_zh(lemma, dom, cedict, cuv_freq):
    """Chinese glosses, ranked by what the CUV actually uses."""
    if lemma in ZH_OVERRIDE:
        return ZH_OVERRIDE[lemma]
    candidates = Counter()
    candidates.update(cedict["v" if dom == "v" else "n"].get(lemma, Counter()))
    candidates.update(cedict["name"].get(lemma, Counter()))
    if not candidates:
        candidates.update(cedict["n"].get(lemma, Counter()))
        candidates.update(cedict["v"].get(lemma, Counter()))
    if not candidates:
        return []
    scored = []
    for zh, cedict_weight in candidates.items():
        used = cuv_freq.get(zh, 0)
        bonus = 0.7 if len(zh) == 1 else 1.1 if len(zh) == 2 else 1.15
        score = math.log1p(used) * bonus + 0.12 * math.log1p(cedict_weight)
        scored.append((score, used, cedict_weight, zh))
    scored.sort(reverse=True)
    return [zh for _, _, _, zh in scored[:3]]


def clean_gloss(gloss, limit=110):
    text = gloss.split(";")[0].strip()
    text = re.sub(r"\s+", " ", text)
    if len(text) > limit:
        text = text[: limit - 1].rsplit(" ", 1)[0] + "…"
    return text


# WordNet describes biblical people and places with a handful of recurring
# phrases. They let ACAI take over the Chinese label without turning ordinary
# words that happen to be names ("job", "temple") into proper nouns.
PROPER_SENSE = re.compile(
    r"\((Old|New) Testament\)|\bking of\b|\bson of\b|\bprophet\b|"
    r"\bapostle\b|\bdisciple\b|\bemperor\b|\bkingdom\b|\bhigh priest\b",
    re.I,
)


def looks_proper(senses):
    return bool(senses) and any(PROPER_SENSE.search(s) for s in senses)


def zh_for(lemma, dom, open_entry, cedict, cuv_freq, acai, proper):
    """Chinese senses: manual biblical wording, ACAI names, Open Dictionary, CC-CEDICT."""
    if lemma in ZH_OVERRIDE:
        return ZH_OVERRIDE[lemma]
    # A name Open Dictionary classified as a common noun ("Moab" as the bomb,
    # "Gilead" as a novel's country) must lose to ACAI when the word is a name
    # in this text.
    if proper and lemma in acai:
        return [acai[lemma]]
    if open_entry and open_entry.get("zh"):
        return open_entry["zh"]
    if open_entry and open_entry.get("proper") and lemma in acai:
        return [acai[lemma]]
    return choose_zh(lemma, dom, cedict, cuv_freq)


def build(vocab):
    freq = load_frequency()
    bible, seen, bigrams, previous, cuv = load_bible()
    index, data = load_wordnet()
    ipa = load_cmudict()
    cedict, zh_candidates = load_cedict()
    cuv_freq = load_cuv_freq(cuv, zh_candidates)
    open_entries, open_forms = load_open_dict()
    acai = load_acai_names()

    if vocab == "bible":
        wanted = set(seen) | set(acai) | set(FUNCTION)
    elif vocab == "top33k":
        wanted = set(freq) | set(seen) | set(acai) | set(FUNCTION)
    elif vocab == "top10k":
        wanted = {w for w, r in freq.items() if r <= 10000} | set(seen) | set(acai) | set(FUNCTION)
    else:
        # "all" is every single-word headword in Open Dictionary, plus the
        # Bible forms and names no general dictionary has.
        wanted = set(open_entries) | set(seen) | set(acai) | set(FUNCTION) | set(ZH_OVERRIDE) | set(ARCHAIC_FORM.values())

    # form → lemma. A lemma is a WordNet headword, a known base form, or the
    # word itself.
    form_of = {}
    lemmas = set()

    def candidates(w):
        out = [w]
        if len(w) > 3 and w.endswith("ies"):
            out.append(w[:-3] + "y")
        if len(w) > 3 and w.endswith("es"):
            out += [w[:-2], w[:-1]]
        elif len(w) > 3 and w.endswith("s"):
            out.append(w[:-1])
        if len(w) > 4 and w.endswith("ied"):
            out.append(w[:-3] + "y")
        if len(w) > 4 and w.endswith("ed"):
            out += [w[:-2], w[:-1]]
        if len(w) > 4 and w.endswith("ing"):
            out += [w[:-3], w[:-3] + "e"]
            if w[-4] == w[-5]:
                out.append(w[:-4])
        # KJV: giveth, sitteth, knowest. The bare stem and the silent-e stem
        # cover both, and doubled consonants need one more cut.
        if len(w) > 4 and w.endswith("eth"):
            out += [w[:-3], w[:-3] + "e"]
            if w[-4] == w[-5]:
                out.append(w[:-4])
        if len(w) > 4 and w.endswith("est"):
            out += [w[:-3], w[:-3] + "e"]
            if w[-4] == w[-5]:
                out.append(w[:-4])
        return [c for c in dict.fromkeys(out) if c]

    for w in wanted:
        base = IRREGULAR_FORM.get(w)
        if base:
            form_of[w] = base
            lemmas.add(base)
            continue
        for cand in candidates(w):
            if cand in index:
                form_of[w] = cand
                lemmas.add(cand)
                break
        else:
            form_of[w] = w

    # Open Dictionary knows the irregular forms WordNet's suffix rules miss:
    # spake, dwelt, smote, arose, sinned. Only forms that occur in the text are
    # indexed, so the form table stays a lookup aid, not a second dictionary.
    for form, base in open_forms.items():
        if form in seen and (form not in form_of or form_of[form] == form) and base != form:
            form_of[form] = base
            lemmas.add(base)

    entries = {}
    for lemma in sorted(lemmas | wanted):
        fn = FUNCTION.get(lemma)
        if fn:
            pos, zh, sense = fn
            entries[lemma] = {
                "w": lemma,
                "pos": pos,
                "uk": IPA_OVERRIDE.get(lemma, ipa.get(lemma, "")),
                "s": [sense],
                "zh": [zh],
                "c": [],
                "n": bible.get(lemma, 0),
                "r": freq.get(lemma, 0),
            }
            continue

        # "wilt" and "art" are ordinary WordNet nouns; in this text they are
        # KJV forms of "will" and "be", so let the pointer entry win.
        if lemma in ARCHAIC_FORM:
            continue

        # In the reader, "saw" is far more often the past of "see" than a
        # carpenter's tool. The pointer entry makes that the card's reading.
        # The same applies to every other inflected form: "turned" should show
        # turn, not an adjective sense WordNet happens to record. Forms that are
        # themselves names (Israelites) keep their own entry.
        if (form_of.get(lemma, lemma) != lemma
                and (lemma in INFLECTION_FIRST
                     or (lemma not in acai and lemma not in open_entries))):
            continue

        open_entry = open_entries.get(lemma)
        by_pos = {}
        for p, off in index.get(lemma, [])[:12]:
            gloss = data.get((p, off))
            if gloss:
                by_pos.setdefault(p, []).append(clean_gloss(gloss))
        senses = []
        dom = ""
        if by_pos:
            # the part of speech with the most senses is what the word mostly is,
            # which keeps "begin" a verb and stops the politician noun winning
            dom = max(by_pos, key=lambda p: len(by_pos[p]))
            pos = {"n": "n.", "v": "v.", "a": "adj.", "r": "adv."}.get(dom, "")
            senses = [g for g in dict.fromkeys(by_pos[dom]) if g][:3]
            if lemma in SENSE_OVERRIDE:
                senses = SENSE_OVERRIDE[lemma]
        else:
            pos = open_entry["pos"] if open_entry else ""
        proper = lemma in acai and (not senses or looks_proper(senses))
        if proper:
            pos = "专名"
        pos = POS_OVERRIDE.get(lemma, pos)
        zh = zh_for(lemma, dom, open_entry, cedict, cuv_freq, acai, proper)
        if not zh and not senses:
            continue
        entries[lemma] = {
            "w": lemma,
            "pos": pos,
            "uk": IPA_OVERRIDE.get(lemma, ipa.get(lemma, "") or (open_entry["ipa"] if open_entry else "")),
            "s": senses,
            "zh": zh,
            "c": collocations(lemma, bigrams, previous),
            "n": bible.get(lemma, 0),
            "r": freq.get(lemma, 0),
        }

    # Every wanted word that is only an inflected form gets a pointer entry so
    # the card can say what it is a form of.
    for form, lemma in form_of.items():
        if form in entries or lemma not in entries or form == lemma:
            continue
        entries[form] = {"w": form, "lemma": lemma,
                         "uk": IPA_OVERRIDE.get(form, ipa.get(form, "")),
                         "n": bible.get(form, 0), "r": freq.get(form, 0)}

    write(entries, form_of)


def write(entries, form_of):
    words = sorted(entries)
    index = {"words": words, "forms": {f: l for f, l in form_of.items() if f != l and l in entries}, "entries": {}}
    payload = bytearray()
    for w in words:
        blob = zlib.compress(json.dumps(entries[w], ensure_ascii=False, separators=(",", ":")).encode(), 9)
        index["entries"][w] = [len(payload), len(blob)]
        payload += blob
    idx = zlib.compress(json.dumps(index, ensure_ascii=False, separators=(",", ":")).encode(), 9)
    with open(OUT, "wb") as fh:
        fh.write(b"LEXI2\n")
        fh.write(struct.pack("<I", len(idx)))
        fh.write(idx)
        fh.write(payload)
    print(f"words {len(words)} · lemmas {len(index['forms'])} forms · index {len(idx)/1e6:.2f} MB · payload {len(payload)/1e6:.2f} MB · file {os.path.getsize(OUT)/1e6:.2f} MB")

if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("-vocab", default="all", choices=["all", "bible", "top10k", "top33k"])
    build(ap.parse_args().vocab)
