package dict

import (
	"strings"
	"testing"
)

func TestTheDictionaryLoads(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	if n := Words(); n < 30000 {
		t.Fatalf("words = %d, want at least 30000", n)
	}
}

func TestAnInflectedFormBorrowsTheMeaning(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	e, ok := Lookup("created")
	if !ok {
		t.Fatal("created is missing")
	}
	if e.Lemma != "create" {
		t.Fatalf("lemma = %q, want create", e.Lemma)
	}
	if len(e.Zh) == 0 || len(e.Senses) == 0 {
		t.Fatalf("created has no meaning: %+v", e)
	}
	if e.IPA == "" {
		t.Error("created has no pronunciation of its own")
	}
}

func TestAHeadwordCarriesGlossPronunciationAndCount(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	e, ok := Lookup("shepherd")
	if !ok {
		t.Fatal("shepherd is missing")
	}
	if e.Pos == "" || e.IPA == "" || len(e.Zh) == 0 {
		t.Fatalf("shepherd is thin: %+v", e)
	}
	if e.Verses == 0 {
		t.Error("shepherd was never counted in the text")
	}
}

func TestAWordOutsideTheDictionaryIsReportedMissing(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	if _, ok := Lookup("zzzqqq"); ok {
		t.Fatal("a nonsense word was found")
	}
}

func TestFunctionWordsAreDefined(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	for _, word := range []string{
		"the", "a", "an", "and", "of", "to", "in", "on",
		"yourselves", "themselves", "one", "well", "still", "very",
	} {
		e, ok := Lookup(word)
		if !ok {
			t.Fatalf("%s is missing", word)
		}
		if e.Pos == "" || len(e.Zh) == 0 || len(e.Senses) == 0 {
			t.Fatalf("%s is thin: %+v", word, e)
		}
	}
}

func TestBiblicalTermsUseTheChineseBibleWording(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	e, ok := Lookup("God")
	if !ok {
		t.Fatal("God is missing")
	}
	if len(e.Zh) == 0 || e.Zh[0] != "神" {
		t.Fatalf("God zh = %v, want 神 first", e.Zh)
	}
	if !strings.Contains(e.IPA, "ɑː") {
		t.Fatalf("God IPA = %q, want a long vowel", e.IPA)
	}
	for word, want := range map[string][2]string{
		"righteous": {"adj.", "公义"},
		"holy":      {"adj.", "神圣的"},
		"flock":     {"n.", "羊群"},
	} {
		pos, zh := want[0], want[1]
		e, ok := Lookup(word)
		if !ok {
			t.Fatalf("%s is missing", word)
		}
		if e.Pos != pos || len(e.Zh) == 0 || e.Zh[0] != zh {
			t.Fatalf("%s = %q %v, want %q %q", word, e.Pos, e.Zh, pos, zh)
		}
	}
}

func TestOpenDictionaryCoversGeneralWords(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	for word, want := range map[string]string{
		"against": "逆着",
		"among":   "混在",
		"church":  "教会",
		"didn’t":  "过去没有",
		"God's":   "神",
	} {
		e, ok := Lookup(word)
		if !ok {
			t.Fatalf("%s is missing", word)
		}
		if len(e.Zh) == 0 || !strings.Contains(e.Zh[0], want) {
			t.Fatalf("%s zh = %v, want %q", word, e.Zh, want)
		}
	}
}

func TestBiblicalNamesUseTheChineseBibleWording(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	for word, want := range map[string]string{
		"gilead": "基列",
		"joab":   "约押",
		"moab":   "摩押",
	} {
		e, ok := Lookup(word)
		if !ok {
			t.Fatalf("%s is missing", word)
		}
		if e.Pos != "专名" || len(e.Zh) == 0 || e.Zh[0] != want {
			t.Fatalf("%s = %q %v, want 专名 %q", word, e.Pos, e.Zh, want)
		}
	}
}

func TestHomographsPreferTheTextReading(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	for word, lemma := range map[string]string{
		"saw": "see", "rose": "rise", "left": "leave", "found": "find",
	} {
		e, ok := Lookup(word)
		if !ok {
			t.Fatalf("%s is missing", word)
		}
		if e.Lemma != lemma {
			t.Fatalf("%s lemma = %q, want %q", word, e.Lemma, lemma)
		}
		if len(e.Zh) == 0 || len(e.Senses) == 0 {
			t.Fatalf("%s has no meaning: %+v", word, e)
		}
	}
}

func TestKJVFormsResolveToTheirBase(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	forms := map[string]string{
		"hath": "have", "saith": "say", "giveth": "give", "sitteth": "sit",
		"wilt": "will", "art": "be", "shalt": "shall", "knowest": "know",
		"spake": "speak", "dwelt": "dwell", "smote": "smite", "sinned": "sin",
		"arose": "arise",
	}
	for form, lemma := range forms {
		e, ok := Lookup(form)
		if !ok {
			t.Fatalf("%s is missing", form)
		}
		if e.Lemma != lemma {
			t.Fatalf("%s lemma = %q, want %q", form, e.Lemma, lemma)
		}
		if len(e.Zh) == 0 || len(e.Senses) == 0 {
			t.Fatalf("%s has no meaning: %+v", form, e)
		}
	}
}
