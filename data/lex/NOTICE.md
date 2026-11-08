# Dictionary data notice

`app/internal/dict/dict.bin` is a modified, merged dictionary derived from the
following sources. It is distributed under the **Creative Commons
Attribution-ShareAlike 4.0 International** license (CC BY-SA 4.0), except where
a source below says otherwise.

- **Open Dictionary v2.0** — <https://github.com/ahpxex/open-dictionary>
  - Data: CC BY-SA 4.0. Upstream material: English Wiktionary and Wiktextract.
  - Changes: selected single-word headwords; selected one part-of-speech group
    per headword; limited to three short Chinese glosses; merged word forms;
    removed examples, relations, etymology and pronunciation variants.
- **ACAI (Aquifer Church Answers Index)** — <https://github.com/BibleAquifer/ACAI>
  - Data: CC BY-SA 4.0, © BiblioNexus 2024.
  - Changes: extracted preferred English and Simplified Chinese labels only;
    excluded alternate labels and entity descriptions.
- **CC-CEDICT** — <https://cc-cedict.org/>
  - Data: CC BY-SA 4.0. Changes: inverted English gloss index; ranked Chinese
    candidates against the Chinese Union Version; used as a fallback.
- **Princeton WordNet 3.0** — <https://wordnet.princeton.edu/>
  - License: Princeton WordNet License (permissive). Changes: selected up to
    three glosses for the dominant part of speech; truncated long glosses.
- **CMU Pronouncing Dictionary** — <http://www.speech.cs.cmu.edu/cgi-bin/cmudict>
  - License: BSD-style. Changes: converted ARPAbet to a compact IPA string and
    marked long vowels.
- **FrequencyWords `en_50k`** — <https://github.com/hermitdave/FrequencyWords>
  - License: CC BY 4.0. Used for headword rank only.

License texts and attribution URLs are included above. If you redistribute a
modified `dict.bin`, keep this notice, state the changes, and license the
resulting data under CC BY-SA 4.0.

The Go application code is separate from this data and may be licensed
independently.
