package bible

import "testing"

func TestTheStoreHoldsTheWholeBible(t *testing.T) {
	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(s.Books()), 66; got != want {
		t.Fatalf("books = %d, want %d", got, want)
	}
	if got, want := len(s.Translations()), 4; got != want {
		t.Fatalf("translations = %d, want %d", got, want)
	}
}

func TestAChineseChapterDecodes(t *testing.T) {
	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := s.Chapter("CUV", "Gen", 1)
	if !ok {
		t.Fatal("the Chinese Genesis 1 is missing")
	}
	if len(ch.Verses) < 30 {
		t.Fatalf("the Chinese Genesis 1 has %d verses", len(ch.Verses))
	}
	if ch.Verses[0].Text == "" {
		t.Fatal("the Chinese Genesis 1:1 is empty")
	}
	if t1, ok := s.Translation("CUV"); !ok || !t1.IsChinese() {
		t.Fatalf("CUV is not marked as Chinese: %+v", t1)
	}
	if got := len(s.InLanguage("en")); got != 3 {
		t.Fatalf("English translations = %d, want 3", got)
	}
}

func TestAChapterDecodes(t *testing.T) {
	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := s.Chapter("BSB", "Gen", 1)
	if !ok {
		t.Fatal("Genesis 1 is missing")
	}
	if len(ch.Verses) != 31 {
		t.Fatalf("Genesis 1 has %d verses, want 31", len(ch.Verses))
	}
	if len(ch.Headings) == 0 {
		t.Fatal("Genesis 1 has no headings")
	}
	if len(ch.Verses[0].Text) == 0 {
		t.Fatal("Genesis 1:1 is empty")
	}
}

func TestTheLongestPsalmDecodes(t *testing.T) {
	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := s.Chapter("KJV", "Ps", 119)
	if !ok {
		t.Fatal("Psalm 119 is missing")
	}
	if len(ch.Verses) != 176 {
		t.Fatalf("Psalm 119 has %d verses, want 176", len(ch.Verses))
	}
}

func TestEveryBookHasAChineseName(t *testing.T) {
	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range s.Books() {
		if ZH(b.OSIS) == "" || Mark(b.OSIS) == "" {
			t.Errorf("%d (%s) has no Chinese name", b.ID, b.OSIS)
		}
	}
}
