package desk

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"

	"app/internal/store"
	"app/internal/theme"
	"app/internal/vocab"
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

// clickRightmost picks the last control with this label to the right, for
// labels that exist both in the filter column and in the detail pane.
func clickRightmost(t *testing.T, d *app.Driver, label string) {
	t.Helper()
	var best ui.Hit
	found := false
	for _, h := range d.Snapshot().Hits {
		if h.Label == label && (!found || h.Bounds.X > best.Bounds.X) {
			best, found = h, true
		}
	}
	if !found {
		t.Fatalf("no control labelled %q on screen", label)
	}
	b := best.Bounds
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
	if !paintedMatch(d, `\d+ / \d+ 页`) {
		t.Error("the position line is missing")
	}
	// One spread is one page: both halves print the same folio.
	if m.folio(true) != m.spread+1 || m.folio(false) != m.spread+1 {
		t.Errorf("folios = %d/%d, want %d on both halves", m.folio(true), m.folio(false), m.spread+1)
	}
	if painted(d, "3–4") || painted(d, "1–2") {
		t.Error("the position line still prints a folio range")
	}
}

// TestTheWindowOwnsItsTitlebar checks the strip above the masthead: the app
// paints the window's own titlebar there, so the strip moves the window and
// carries the imprint, while the masthead below takes its own presses.
func TestTheWindowOwnsItsTitlebar(t *testing.T) {
	d, _ := open(t)
	if !painted(d, "Agnel") {
		t.Fatal("the titlebar has no imprint")
	}
	if got := d.PressAt(720, titlebarH/2); got != app.PressMove {
		t.Errorf("press in the titlebar = %v, want a window move", got)
	}
	if got := d.PressAt(720, titlebarH+20); got == app.PressMove {
		t.Error("the masthead below the titlebar still moves the window")
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

func TestBlankPageClickClosesTheCard(t *testing.T) {
	d, m := open(t)
	click(t, d, "created")
	if m.lookup == nil {
		t.Fatal("lookup did not open")
	}
	// A click in the page margin, inside the reading area but not on a word.
	area := m.lastArea
	if area.Width == 0 {
		t.Fatal("no page area was laid out")
	}
	if err := d.Click(area.X+6, area.Y+area.Height-6); err != nil {
		t.Fatal(err)
	}
	if m.lookup != nil {
		t.Fatalf("lookup stayed open after a blank page click: %q", m.lookup.word)
	}
}

func TestClickingAnotherWordSwitchesTheCard(t *testing.T) {
	d, m := open(t)
	click(t, d, "created")
	click(t, d, "beginning")
	if m.lookup == nil || m.lookup.word != "beginning" {
		t.Fatalf("lookup = %+v, want the word 'beginning'", m.lookup)
	}
}

func TestSwitchingToTheBankClosesTheCard(t *testing.T) {
	d, m := open(t)
	click(t, d, "created")
	if m.lookup == nil {
		t.Fatal("lookup did not open")
	}
	click(t, d, "词库")
	if m.pane != paneBank {
		t.Fatalf("pane = %v, want the bank", m.pane)
	}
	if m.lookup != nil {
		t.Fatalf("the card stayed open over the bank: %q", m.lookup.word)
	}
}

// openWithVocab starts the app over a word list written to the temporary place,
// so the bank tests begin with words instead of an empty file.
func openWithVocab(t *testing.T, words ...vocab.Word) (*app.Driver, *State) {
	t.Helper()
	store.Dir = t.TempDir()
	vocab.Dir = store.Dir
	f := vocab.Empty()
	for _, w := range words {
		f.Add(w)
	}
	if err := vocab.Save(f); err != nil {
		t.Fatal(err)
	}
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

func savedWords(t *testing.T) *vocab.File {
	t.Helper()
	f, err := vocab.Load()
	if err != nil {
		t.Fatalf("vocab.Load: %v", err)
	}
	return f
}

// The card's button saves the word with the verse it was met in, and the same
// button takes it out again.
func TestTheCardSavesAndUnsaves(t *testing.T) {
	d, m := open(t)
	click(t, d, "beginning")
	if !painted(d, "收藏这个词") {
		t.Fatal("the card should offer to save the word")
	}
	click(t, d, "收藏这个词")
	if !painted(d, "已收藏") {
		t.Fatal("the card should say the word is saved")
	}
	onDisk := savedWords(t)
	if onDisk.Count() != 1 || onDisk.Words[0].Word != "beginning" {
		t.Fatalf("vocab.json should hold beginning, got %+v", onDisk.Words)
	}
	if len(onDisk.Words[0].Contexts) != 1 {
		t.Fatalf("the verse should be saved with the word, got %+v", onDisk.Words[0].Contexts)
	}
	ctx := onDisk.Words[0].Contexts[0]
	if ctx.Book != "Gen" || ctx.Chapter != 1 || ctx.Verse != 1 || !strings.Contains(ctx.EN, "In the beginning") {
		t.Fatalf("unexpected context %+v", ctx)
	}
	if m.vocab.Count() != 1 {
		t.Fatalf("memory and disk disagree: %d words", m.vocab.Count())
	}
	click(t, d, "已收藏 · 移出")
	if m.vocab.Count() != 0 || savedWords(t).Count() != 0 {
		t.Fatal("a second click should take the word out")
	}
}

// The bookmark-era word list moves into the word bank once, and the old field is
// cleared so it does not come back.
func TestOldSavedWordsMoveOver(t *testing.T) {
	store.Dir = t.TempDir()
	if err := store.Save(store.State{Saved: []string{"light", "beginning"}}); err != nil {
		t.Fatal(err)
	}
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	m := New()
	Config(m)
	if m.vocab == nil || m.vocab.Count() != 2 {
		t.Fatalf("two words should move over, got %+v", m.vocab)
	}
	if w := m.vocab.Words[0]; w.Stage != vocab.StageNew || len(w.Zh) == 0 {
		t.Fatalf("a moved word should take the dictionary entry, got %+v", w)
	}
	st, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Saved) != 0 {
		t.Fatalf("the old list should be cleared, got %v", st.Saved)
	}
}

// The bank lists the saved words and its filters count them.
func TestTheBankListsAndFilters(t *testing.T) {
	d, m := openWithVocab(t,
		vocab.Word{Word: "light", Stage: vocab.StageNew},
		vocab.Word{Word: "void", Stage: vocab.StageKnown},
		vocab.Word{Word: "deep", Stage: vocab.StageLearning},
	)
	click(t, d, "词库")
	if !painted(d, "3 个词") {
		t.Fatal("the bar should count the words")
	}
	for _, word := range []string{"light", "void", "deep"} {
		if !painted(d, word) {
			t.Fatalf("%q should be in the list", word)
		}
	}
	click(t, d, "已掌握")
	if m.filter != 3 {
		t.Fatalf("the known filter should be on, filter=%d", m.filter)
	}
	if m.bankWords()[0].Word != "void" {
		t.Fatalf("the known filter should keep void, got %v", m.bankWords())
	}
}

// The collection is empty: the surface says how words get here and offers the
// two ways out, instead of a bare wall under a filter bar.
func TestTheEmptyBankOffersTheNextStep(t *testing.T) {
	d, m := openWithVocab(t)
	click(t, d, "词库")
	if m.bankCounts()[0] != 0 {
		t.Fatalf("the fixture should start empty, got %d", m.bankCounts()[0])
	}
	if !painted(d, "词库还空着") {
		t.Fatal("an empty collection should say so")
	}
	click(t, d, "手动添加一个词")
	if !m.adding || !painted(d, "添加一个词") {
		t.Fatal("the empty state should open the add form in the drawer")
	}
	click(t, d, "取消")
	click(t, d, "去阅读")
	if m.pane != paneReader {
		t.Fatal("去阅读 should return to the reading pane")
	}
}

// Choosing a stage on the detail writes it through to disk.
func TestStageChangeIsSaved(t *testing.T) {
	d, m := openWithVocab(t, vocab.Word{Word: "light", Stage: vocab.StageNew})
	click(t, d, "词库")
	click(t, d, "light")
	clickRightmost(t, d, "已掌握")
	onDisk := savedWords(t)
	if onDisk.Count() != 1 || onDisk.Words[0].Stage != vocab.StageKnown {
		t.Fatalf("stage should be known on disk, got %+v", onDisk.Words)
	}
	_ = m
}

// The search box looks through the word, its Chinese and the verses it was
// saved from.
func TestSearchFindsWordsByChinese(t *testing.T) {
	d, m := openWithVocab(t,
		vocab.Word{Word: "light", Zh: []string{"光；光明"}},
		vocab.Word{Word: "darkness", Zh: []string{"黑暗"}},
	)
	click(t, d, "词库")
	if len(m.bankWords()) != 2 {
		t.Fatalf("both words should be listed, got %d", len(m.bankWords()))
	}
	click(t, d, "搜索词库")
	d.Type("光")
	if got := m.bankWords(); len(got) != 1 || got[0].Word != "light" {
		t.Fatalf("searching 光 should leave light, got %+v", got)
	}
}

// Deleting asks first; cancelling keeps the word, confirming removes it.
func TestDeleteAsksFirst(t *testing.T) {
	d, m := openWithVocab(t, vocab.Word{Word: "light"})
	click(t, d, "词库")
	click(t, d, "light")
	click(t, d, "移出词库")
	if !painted(d, "移出词库？") {
		t.Fatal("a confirmation should be on screen")
	}
	click(t, d, "取消")
	if m.vocab.Count() != 1 {
		t.Fatal("cancelling should keep the word")
	}
	click(t, d, "移出词库")
	click(t, d, "删除")
	if m.vocab.Count() != 0 || savedWords(t).Count() != 0 {
		t.Fatal("confirming should delete the word")
	}
}

// A word typed in by hand is saved, and a word the dictionary does not have is
// kept anyway.
func TestManualAdd(t *testing.T) {
	d, m := open(t)
	click(t, d, "词库")
	click(t, d, "手动添加")
	click(t, d, "单词")
	if err := d.Type("covenant"); err != nil {
		t.Fatal(err)
	}
	click(t, d, "中文释义")
	if err := d.Type("约"); err != nil {
		t.Fatal(err)
	}
	click(t, d, "添加")
	if m.vocab.Count() != 1 {
		t.Fatalf("the word should be added, got %+v", m.vocab.Words)
	}
	w := m.vocab.Words[0]
	if w.Word != "covenant" || len(w.Zh) == 0 || w.Zh[0] != "约" {
		t.Fatalf("unexpected word %+v", w)
	}
}

// 回到原文 closes the bank, opens the reader at the word's chapter, and lands on
// the page where that verse is.
func TestReturnToTheVerse(t *testing.T) {
	d, m := openWithVocab(t, vocab.Word{
		Word: "formless", Stage: vocab.StageNew,
		Contexts: []vocab.Context{{Translation: "BSB", Book: "Gen", Chapter: 1, Verse: 2,
			EN: "Now the earth was formless and void", ZH: "地是空虚混沌"}},
	})
	click(t, d, "词库")
	click(t, d, "formless")
	click(t, d, "回到原文")
	if m.pane != paneReader {
		t.Fatal("the reader should be on screen")
	}
	if m.osis != "Gen" || m.chapter != 1 {
		t.Fatalf("should be Genesis 1, got %s %d", m.osis, m.chapter)
	}
	if !painted(d, "formless") {
		t.Fatal("the verse should be on the page")
	}
	if m.landVerse != 0 {
		t.Fatalf("the landing verse should be consumed, got %d", m.landVerse)
	}
	// The card opens on the word itself, in the verse it was taken from.
	if m.lookup == nil || m.lookup.word != "formless" || m.lookup.v != 2 {
		t.Fatalf("lookup = %+v, want 'formless' at Genesis 1:2", m.lookup)
	}
	if err := d.Advance(20 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if !painted(d, "已收藏 · 移出") {
		t.Fatal("the card for the landing word should be on screen")
	}
}

// marked counts painted rectangles of a colour; the saved-word marks are the
// only thing on the page using these two colours.
func marked(d *app.Driver, color ui.Color) int {
	n := 0
	for _, p := range d.Snapshot().Primitives {
		if p.Color == color {
			n++
		}
	}
	return n
}

func TestSavedWordsAreMarkedByStage(t *testing.T) {
	for _, tc := range []struct {
		name      string
		stage     string
		wantNew   bool
		wantLearn bool
	}{
		{"new", vocab.StageNew, true, false},
		{"learning", vocab.StageLearning, false, true},
		{"known", vocab.StageKnown, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, _ := openWithVocab(t, vocab.Word{Word: "created", Stage: tc.stage})
			if got := marked(d, theme.PaperMarkNew) > 0; got != tc.wantNew {
				t.Errorf("new-word mark painted = %v, want %v", got, tc.wantNew)
			}
			if got := marked(d, theme.PaperMarkLearning) > 0; got != tc.wantLearn {
				t.Errorf("learning-word mark painted = %v, want %v", got, tc.wantLearn)
			}
		})
	}
}

// --- the whole dictionary, beside the saved words ---------------------------

func TestTheDictionaryViewListsEveryHeadword(t *testing.T) {
	d, m := open(t)
	click(t, d, "词库")
	click(t, d, "全部词典")
	if !m.dictMode {
		t.Fatal("the dictionary view did not open")
	}
	if !painted(d, "aback") {
		t.Error("the first page does not start at the beginning of the alphabet")
	}
	if painted(d, "zuzim") {
		t.Error("the whole lexicon was laid out at once instead of the current page")
	}
	if got := len(m.dictWords()); got < 41_000 {
		t.Errorf("dictWords() has %d entries, want all 41,873 words", got)
	}
	// Clicking "下一页" advances to page 2 (index 1) and replaces page 1's words.
	click(t, d, "下一页")
	if m.dictPg != 1 {
		t.Fatalf("dictPg after 下一页 = %d, want 1", m.dictPg)
	}
	if painted(d, "aback") {
		t.Error("page 1 word 'aback' is still painted on page 2")
	}
}

func TestALetterJumpsToItsWords(t *testing.T) {
	d, m := open(t)
	click(t, d, "词库")
	click(t, d, "全部词典")
	click(t, d, "S")
	d.Advance(40 * time.Millisecond)
	if !painted(d, "sabbath") {
		t.Error("the list did not jump to S")
	}
	if painted(d, "aback") {
		t.Error("the list is still showing the start of the alphabet")
	}
	if m.selected != "" {
		t.Error("jumping alone opened a word")
	}
}

func TestDictionarySearchFindsPrefixes(t *testing.T) {
	d, m := open(t)
	click(t, d, "词库")
	click(t, d, "全部词典")
	click(t, d, "搜索词库")
	d.Type("covenant")
	d.Advance(40 * time.Millisecond)
	words := m.dictWords()
	if len(words) == 0 || words[0] != "covenant" {
		t.Fatalf("dictWords = %v, want the prefix matches starting at covenant", words)
	}
	for _, w := range words {
		if !strings.HasPrefix(w, "covenant") {
			t.Errorf("%q does not start with the query", w)
		}
	}
	if !painted(d, "covenant") {
		t.Error("the match is not on the page")
	}
}

func TestTheDictionarySavesToTheBank(t *testing.T) {
	d, m := open(t)
	click(t, d, "词库")
	click(t, d, "全部词典")
	click(t, d, "搜索词库")
	d.Type("covenant")
	d.Advance(40 * time.Millisecond)
	click(t, d, "covenant")
	click(t, d, "收藏这个词")
	if m.vocab == nil || !m.vocab.Known("covenant") {
		t.Fatal("the word was not saved")
	}
	if !painted(d, "已收藏 · 移出") {
		t.Error("the detail still offers to save it")
	}
	click(t, d, "我的词")
	if m.dictMode {
		t.Fatal("the bank did not come back")
	}
	if !painted(d, "covenant") {
		t.Error("the saved word is not in the bank")
	}
}

func TestLookupCardLemmaJumpAndOpenInBank(t *testing.T) {
	d, m := openWithVocab(t, vocab.Word{Word: "created", Stage: vocab.StageNew})
	click(t, d, "created")
	if m.lookup == nil || m.lookup.word != "created" {
		t.Fatalf("lookup = %+v, want created", m.lookup)
	}
	// Clicking "在词库查看" opens the Word Bank with the word selected.
	click(t, d, "在词库查看")
	if m.pane != paneBank || m.selected != "created" {
		t.Fatalf("pane = %v, selected = %q, want paneBank and 'created'", m.pane, m.selected)
	}
	click(t, d, "阅读")
	click(t, d, "created")
	// Clicking the lemma chip switches the lookup card in-place to the base word.
	click(t, d, "这是 create 的变形")
	if m.lookup == nil || m.lookup.word != "create" {
		t.Fatalf("lookup after lemma click = %+v, want 'create'", m.lookup)
	}
}

func TestBankAutoSelectAndArrowKeyNavigation(t *testing.T) {
	t0 := time.Now()
	d, m := openWithVocab(t,
		vocab.Word{Word: "void", Stage: vocab.StageNew, Added: t0.Add(-2 * time.Minute)},
		vocab.Word{Word: "beginning", Stage: vocab.StageLearning, Added: t0.Add(-1 * time.Minute)},
		vocab.Word{Word: "created", Stage: vocab.StageKnown, Added: t0},
	)
	click(t, d, "词库")
	if m.selected != "created" {
		t.Fatalf("auto-selected word = %q, want newest word 'created'", m.selected)
	}
	if !m.stepBankSelection(1) || m.selected != "beginning" {
		t.Fatalf("step down selected = %q, want 'beginning'", m.selected)
	}
	if !m.stepBankSelection(1) || m.selected != "void" {
		t.Fatalf("step down selected = %q, want 'void'", m.selected)
	}
	if !m.stepBankSelection(-1) || m.selected != "beginning" {
		t.Fatalf("step up selected = %q, want 'beginning'", m.selected)
	}
	// Clicking "收起详情" closes the right-hand slide-over drawer so the Bento wall expands.
	click(t, d, "收起详情")
	if m.selected != "" {
		t.Fatalf("selected after closing drawer = %q, want empty", m.selected)
	}
	// Clicking any card re-opens the drawer on that word.
	click(t, d, "void")
	if m.selected != "void" {
		t.Fatalf("selected after clicking card = %q, want 'void'", m.selected)
	}
}

func TestBankPagination(t *testing.T) {
	t0 := time.Now()
	words := make([]vocab.Word, 0, 15)
	for i := 0; i < 15; i++ {
		words = append(words, vocab.Word{
			Word:  fmt.Sprintf("word%02d", i+1),
			Stage: vocab.StageNew,
			Added: t0.Add(time.Duration(i) * time.Minute),
		})
	}
	d, m := openWithVocab(t, words...)
	click(t, d, "词库")
	// Newest is word15 (page 1 shows word15..word04, 12 cards per page; page 2 shows word03..word01).
	if m.bankPg != 0 {
		t.Fatalf("initial bankPg = %d, want 0", m.bankPg)
	}
	if !painted(d, "word15") || !painted(d, "word04") {
		t.Fatal("page 1 should display word15 through word04")
	}
	if painted(d, "word03") {
		t.Fatal("word03 belongs to page 2 and should not be painted on page 1")
	}
	click(t, d, "下一页")
	if m.bankPg != 1 {
		t.Fatalf("bankPg after 下一页 = %d, want 1", m.bankPg)
	}
	if !painted(d, "word03") || !painted(d, "word01") {
		t.Fatal("page 2 should display word03 through word01")
	}
	if painted(d, "word15") {
		t.Fatal("word15 belongs to page 1 and should not be painted on page 2")
	}
	click(t, d, "上一页")
	if m.bankPg != 0 {
		t.Fatalf("bankPg after 上一页 = %d, want 0", m.bankPg)
	}
}

func TestBookTurnAnimationPaintsBothPhasesCleanly(t *testing.T) {
	d, m := open(t)
	if err := d.Key(ui.KeyRight, 0); err != nil {
		t.Fatal(err)
	}
	if !m.turning || m.turnDir != 1 {
		t.Fatalf("turning=%v dir=%d, want true and 1", m.turning, m.turnDir)
	}
	// Phase 1 (0 < p < 0.5): right leaf lifting toward the spine.
	if err := d.Advance(220 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if !m.turning {
		t.Fatal("expected turn animation to still be running in phase 1")
	}
	// Phase 2 (0.5 <= p < 1.0): leaf unfolding across the spine onto the left page.
	if err := d.Advance(220 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	// Finish the turn.
	if err := d.Advance(300 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if m.turning {
		t.Fatal("expected turn animation to finish after turnDuration")
	}
}
