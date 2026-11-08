package desk

import (
	"regexp"
	"strings"
	"testing"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"

	"app/internal/store"
)

// open builds a driver over the real store, with the saved place in a temporary
// directory so a test never moves the reader's own bookmark.
func open(t *testing.T) (*app.Driver, *State) {
	t.Helper()
	store.Dir = t.TempDir()
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	m := New()
	driver, err := app.NewDriver(Config(m), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(driver.Close)
	return driver, m
}

func painted(d *app.Driver, want string) bool {
	for _, text := range d.Snapshot().Texts {
		if strings.Contains(text.Value, want) {
			return true
		}
	}
	return false
}

func paintedMatch(d *app.Driver, pattern string) bool {
	re := regexp.MustCompile(pattern)
	for _, text := range d.Snapshot().Texts {
		if re.MatchString(text.Value) {
			return true
		}
	}
	return false
}

func hit(t *testing.T, d *app.Driver, label string) (ui.Hit, bool) {
	t.Helper()
	for _, h := range d.Snapshot().Hits {
		if h.Label == label {
			return h, true
		}
	}
	return ui.Hit{}, false
}

func click(t *testing.T, d *app.Driver, label string) {
	t.Helper()
	h, ok := hit(t, d, label)
	if !ok {
		t.Fatalf("no control labelled %q on screen", label)
	}
	b := h.Bounds
	if err := d.Click(b.X+b.Width/2, b.Y+b.Height/2); err != nil {
		t.Fatal(err)
	}
}

func TestTheSpreadPaintsTwoTranslations(t *testing.T) {
	d, m := open(t)
	if m.left == "" || m.right == "" {
		t.Fatal("no default translations")
	}
	if !painted(d, "创世记 1") {
		t.Error("the title is missing")
	}
	if !painted(d, "In the beginning God created") {
		t.Error("the left page has no text")
	}
	if !painted(d, "起初，") || !painted(d, "神创造天地") {
		t.Error("the right page has no Chinese text")
	}
	if !paintedMatch(d, `\d+–\d+ / \d+ 页`) {
		t.Error("the position line is missing")
	}
}

func TestHeadingsArePrinted(t *testing.T) {
	d, _ := open(t)
	if !painted(d, "The Creation") {
		t.Error("the section heading did not print")
	}
}

func TestTheChapterListIsNotBesideThePage(t *testing.T) {
	d, _ := open(t)
	if _, ok := hit(t, d, "创世记 2"); ok {
		t.Error("the chapters are still listed beside the page")
	}
}

func TestThePageTurnSitsAtTheEdges(t *testing.T) {
	d, m := open(t)
	next, ok := hit(t, d, "下一页")
	if !ok {
		t.Fatal("there is no next-page control")
	}
	if next.Bounds.X < m.lastArea.X {
		t.Error("the next-page control is not on the right")
	}
}

func TestTheTurnButtonStaysBesideThePageInSingleMode(t *testing.T) {
	d, m := open(t)
	click(t, d, "单栏")
	next, ok := hit(t, d, "下一页")
	if !ok {
		t.Fatal("there is no next-page control")
	}
	if next.Bounds.X < m.lastArea.X+m.lastArea.Width {
		t.Errorf("the next-page control is at x=%v, over the page", next.Bounds.X)
	}
}

func TestThePickerReturnsFromTheCornerItWasOpenedFrom(t *testing.T) {
	d, _ := open(t)
	click(t, d, "创世记 1")
	back, ok := hit(t, d, "返回阅读")
	if !ok {
		t.Fatal("there is no way back to the page")
	}
	if back.Bounds.X > 200 {
		t.Errorf("the way back is at x=%v, not in the left corner", back.Bounds.X)
	}
}

func TestPickerChoosesABookAndChapter(t *testing.T) {
	d, m := open(t)
	click(t, d, "创世记 1")
	if !m.picker {
		t.Fatal("the book page did not open")
	}
	click(t, d, "诗篇")
	if m.pickerBook != "Ps" {
		t.Fatalf("pickerBook = %q", m.pickerBook)
	}
	click(t, d, "诗篇 23")
	if m.osis != "Ps" || m.chapter != 23 {
		t.Fatalf("went to %s %d", m.osis, m.chapter)
	}
	if !painted(d, "The LORD is my shepherd") && !painted(d, "The LORD Is My Shepherd") {
		t.Error("Psalm 23 is not on the page")
	}
}

// The chapters of a book open where the book is: no second page to go to.
func TestTheChaptersOpenInPlace(t *testing.T) {
	d, m := open(t)
	click(t, d, "创世记 1")
	if _, ok := hit(t, d, "全部书卷"); ok {
		t.Error("there is still a second page to go back to")
	}
	if _, ok := hit(t, d, "创世记 2"); !ok {
		t.Fatal("the chapters of the open book are not on this page")
	}
	click(t, d, "创世记 2")
	if m.osis != "Gen" || m.chapter != 2 {
		t.Fatalf("went to %s %d", m.osis, m.chapter)
	}
}

func TestPageTurnAndKeys(t *testing.T) {
	d, m := open(t)
	before := m.spread
	click(t, d, "下一页")
	if m.spread != before+1 {
		t.Fatalf("spread = %d after the page button", m.spread)
	}
	if err := d.Key(ui.KeyRight, 0); err != nil {
		t.Fatal(err)
	}
	if m.spread != before+2 {
		t.Fatalf("spread = %d after the right arrow", m.spread)
	}
	if err := d.Key(ui.KeyLeft, 0); err != nil {
		t.Fatal(err)
	}
	if m.spread != before+1 {
		t.Fatalf("spread = %d after the left arrow", m.spread)
	}
}

// A turn at the end of a chapter carries on into the chapter beside it, so
// reading never stops at a boundary.
func TestTheTurnCarriesOnIntoTheNextChapter(t *testing.T) {
	d, m := open(t)
	if err := d.Key(ui.KeyEnd, 0); err != nil {
		t.Fatal(err)
	}
	if m.spread != len(m.pages)-1 {
		t.Fatalf("spread = %d, want the last of %d", m.spread, len(m.pages))
	}
	click(t, d, "下一章")
	if m.chapter != 2 || m.spread != 0 {
		t.Fatalf("went to chapter %d spread %d", m.chapter, m.spread)
	}
	if !painted(d, "创世记 2") {
		t.Error("the title still names the first chapter")
	}
}

// Turning back lands on the last spread of the chapter before, which is the
// page a reader coming from its end wants to see.
func TestTurningBackLandsOnTheLastSpreadOfTheChapterBefore(t *testing.T) {
	d, m := open(t)
	for i := 0; i < 20 && m.spread < len(m.pages)-1; i++ {
		click(t, d, "下一页")
	}
	click(t, d, "下一章")
	if m.chapter != 2 || m.spread != 0 {
		t.Fatalf("went to chapter %d spread %d", m.chapter, m.spread)
	}
	click(t, d, "上一章")
	if m.chapter != 1 {
		t.Fatalf("chapter = %d", m.chapter)
	}
	if m.spread != len(m.pages)-1 {
		t.Fatalf("spread = %d of %d, want the last", m.spread, len(m.pages))
	}
}

// Both ends of a book turn into the book beside it.
func TestTheTurnCrossesBetweenBooks(t *testing.T) {
	d, m := open(t)
	click(t, d, "创世记 1")
	click(t, d, "创世记 50")
	if m.chapter != 50 {
		t.Fatalf("chapter = %d", m.chapter)
	}
	if err := d.Key(ui.KeyEnd, 0); err != nil {
		t.Fatal(err)
	}
	click(t, d, "下一卷")
	if m.osis != "Exod" || m.chapter != 1 || m.spread != 0 {
		t.Fatalf("went to %s %d:%d", m.osis, m.chapter, m.spread)
	}
	if !painted(d, "出埃及记 1") {
		t.Error("the page is still in Genesis")
	}
	click(t, d, "上一卷")
	if m.osis != "Gen" || m.chapter != 50 {
		t.Fatalf("went to %s %d", m.osis, m.chapter)
	}
	if m.spread != len(m.pages)-1 {
		t.Fatalf("spread = %d of %d, want the last", m.spread, len(m.pages))
	}
}

// The canon has two ends and both of them hold.
func TestTheTurnStopsAtTheEndsOfTheCanon(t *testing.T) {
	d, m := open(t)
	if _, ok := hit(t, d, "上一页"); ok {
		t.Error("there is a page before the first one")
	}
	m.osis = "Rev"
	m.chapter = 22
	m.spread = len(m.pages) + 10
	if m.canTurn(1) {
		t.Error("the turn runs on past the last chapter of the last book")
	}
	if !m.canTurn(-1) {
		t.Error("there is no way back from the end of the canon")
	}
}

func TestTheRightPageKeepsTheChineseTextWhenTheLeftChanges(t *testing.T) {
	d, m := open(t)
	before := m.right
	for _, abbr := range []string{"KJV", "WEBP"} {
		m.left = abbr
		m.layoutKey = ""
		if err := d.Pump(); err != nil {
			t.Fatal(err)
		}
		if m.left != abbr {
			t.Fatalf("left = %q, want %q", m.left, abbr)
		}
		if m.right != before {
			t.Fatalf("the right page moved to %q", m.right)
		}
		if !painted(d, "起初，") {
			t.Errorf("the right page lost its Chinese text after choosing %s", abbr)
		}
	}
	if !painted(d, "In the beginning God created the heaven") {
		t.Error("the King James text is not on the page")
	}
}

func TestSingleModePaintsOnePage(t *testing.T) {
	d, m := open(t)
	click(t, d, "单栏")
	if m.mode != Single {
		t.Fatalf("mode = %d", m.mode)
	}
	if !painted(d, "Genesis 1") {
		t.Error("the running head is missing")
	}
	if painted(d, "起初，") {
		t.Error("a second column is still painted")
	}
}

func TestBookmarkMarksTheSpread(t *testing.T) {
	d, m := open(t)
	click(t, d, "加书签")
	if !m.bookmarked() {
		t.Fatal("the spread is not bookmarked")
	}
	if !painted(d, "已加书签") {
		t.Error("the button did not change")
	}
	click(t, d, "已加书签")
	if m.bookmarked() {
		t.Error("the bookmark was not removed")
	}
}

func TestTheNextVisitResumesWhereItStopped(t *testing.T) {
	d, m := open(t)
	click(t, d, "创世记 1")
	click(t, d, "约翰福音")
	click(t, d, "约翰福音 3")
	if m.chapter != 3 {
		t.Fatalf("chapter = %d", m.chapter)
	}
	m.remember()

	again := New()
	Load()
	again.restore()
	if again.osis != "John" || again.chapter != 3 {
		t.Fatalf("restored %s %d", again.osis, again.chapter)
	}
}

func TestPickerNamesBooksInEitherLanguage(t *testing.T) {
	d, m := open(t)
	click(t, d, "创世记 1")
	if !painted(d, "创世记") {
		t.Fatal("Chinese names are missing")
	}
	click(t, d, "English")
	if m.lang != 1 {
		t.Fatalf("lang = %d", m.lang)
	}
	if !painted(d, "Genesis") || !painted(d, "Philippians") {
		t.Error("English names are missing after the switch")
	}
	if painted(d, "创世记") {
		t.Error("Chinese names are still showing")
	}
}

func TestWordLookupOpensAndDismissesCleanly(t *testing.T) {
	d, m := open(t)
	click(t, d, "created")
	if m.lookup == nil || m.lookup.word != "created" {
		t.Fatalf("lookup = %+v, want word 'created'", m.lookup)
	}
	if !painted(d, "/kɹiˈeɪtəd/") {
		t.Error("IPA pronunciation is not painted on the lookup card")
	}
	// Clicking outside the lookup card on the scrim dismisses it without panicking.
	if err := d.Click(20, 400); err != nil {
		t.Fatal(err)
	}
	if m.lookup != nil {
		t.Fatalf("lookup remained open after scrim click: %+v", m.lookup)
	}
}
