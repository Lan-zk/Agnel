package vocab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withTempDir(t *testing.T) {
	t.Helper()
	old := Dir
	Dir = t.TempDir()
	t.Cleanup(func() { Dir = old })
}

func TestLoadMissingIsEmpty(t *testing.T) {
	withTempDir(t)
	f, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if f.Count() != 0 || f.Version != Version {
		t.Fatalf("first run should be an empty v%d list, got v%d with %d words", Version, f.Version, f.Count())
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	withTempDir(t)
	f := Empty()
	f.Add(Word{Word: "Created", IPA: "kriˈeɪtɪd", Zh: []string{"创造"}, Stage: StageNew})
	if err := Save(f); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Count() != 1 || got.Words[0].Word != "Created" || got.Words[0].IPA != "kriˈeɪtɪd" {
		t.Fatalf("round trip lost the word: %+v", got.Words)
	}
}

// A damaged file is reported and left alone, so the words can still be
// recovered by hand instead of being replaced by an empty list.
func TestDamagedFileIsKept(t *testing.T) {
	withTempDir(t)
	path := Path()
	if err := os.WriteFile(path, []byte("{ this is not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Load()
	if err == nil {
		t.Fatal("a damaged file should report an error")
	}
	if f.Count() != 0 {
		t.Fatalf("a damaged file should read as empty, got %d words", f.Count())
	}
	b, readErr := os.ReadFile(path)
	if readErr != nil || string(b) != "{ this is not json" {
		t.Fatalf("Load changed the damaged file: %q, %v", b, readErr)
	}
}

func TestNewerVersionIsRefused(t *testing.T) {
	withTempDir(t)
	if err := os.WriteFile(Path(), []byte(`{"version": 99, "words": []}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("a newer version should not be opened")
	}
}

// Saving twice keeps one entry, matched without case, and keeps the surface
// form the reader first met.
func TestAddKeepsOneEntryPerWord(t *testing.T) {
	withTempDir(t)
	f := Empty()
	f.Add(Word{Word: "Created", Stage: StageNew})
	f.Add(Word{Word: "created", Stage: StageKnown})
	if f.Count() != 1 || f.Words[0].Word != "Created" {
		t.Fatalf("expected one entry keeping \"Created\", got %+v", f.Words)
	}
	if f.Words[0].Stage != StageNew {
		t.Fatalf("adding again should not change the stage, got %q", f.Words[0].Stage)
	}
}

func TestContextsDoNotRepeat(t *testing.T) {
	withTempDir(t)
	f := Empty()
	ctx := Context{Translation: "BSB", Book: "Gen", Chapter: 1, Verse: 1, EN: "In the beginning", ZH: "起初"}
	f.Add(Word{Word: "beginning", Contexts: []Context{ctx}})
	f.Add(Word{Word: "beginning", Contexts: []Context{ctx}})
	if n := len(f.Words[0].Contexts); n != 1 {
		t.Fatalf("the same verse should be recorded once, got %d", n)
	}
	other := Context{Translation: "BSB", Book: "Gen", Chapter: 1, Verse: 2}
	if !f.AddContext("beginning", other) || len(f.Words[0].Contexts) != 2 {
		t.Fatalf("a new verse should be added, got %+v", f.Words[0].Contexts)
	}
}

func TestStagesAndRemove(t *testing.T) {
	withTempDir(t)
	f := Empty()
	f.Add(Word{Word: "light"})
	if !f.SetStage("light", StageLearning) || !f.Known("light") {
		t.Fatal("SetStage should find the word")
	}
	if f.SetStage("light", "forgotten") {
		t.Fatal("an unknown stage should be refused")
	}
	if f.Words[0].Stage != StageLearning {
		t.Fatalf("stage should be learning, got %q", f.Words[0].Stage)
	}
	if !f.Remove("LIGHT") || f.Count() != 0 {
		t.Fatal("Remove should match without case")
	}
	if f.Remove("light") {
		t.Fatal("removing twice should report nothing removed")
	}
}

// Save leaves no temporary file behind.
func TestSaveCleansUp(t *testing.T) {
	withTempDir(t)
	if err := Save(Empty()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entries, err := os.ReadDir(Dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 1 || strings.HasSuffix(names[0], ".tmp") {
		t.Fatalf("expected only vocab.json, got %v", names)
	}
	if _, err := os.Stat(filepath.Join(Dir, "vocab.json")); err != nil {
		t.Fatalf("vocab.json missing: %v", err)
	}
}
