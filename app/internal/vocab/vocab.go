// Package vocab keeps the words a reader has saved, with the verse each was
// met in. It lives in its own file beside the reading place: the word list
// grows with use, while the place is written on every page turn.
package vocab

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Version is the layout of vocab.json. A file from a newer version is not
// opened, so a newer build never loses words to an older one.
const Version = 1

// The three stages a word moves between. The reader sets them by hand; the
// wrong filter is derived from Wrong, not stored as a stage.
const (
	StageNew      = "new"
	StageLearning = "learning"
	StageKnown    = "known"
)

// Context is one place in the text where the word was met: the verse in both
// translations, so the word can be read again in context and taken back to.
type Context struct {
	Translation string `json:"translation"`
	Book        string `json:"book"` // OSIS id
	Chapter     int    `json:"chapter"`
	Verse       int    `json:"verse"`
	EN          string `json:"en"`
	ZH          string `json:"zh,omitempty"`
}

// Word is one saved word.
type Word struct {
	Word     string    `json:"word"`
	Lemma    string    `json:"lemma,omitempty"`
	Pos      string    `json:"pos,omitempty"`
	IPA      string    `json:"ipa,omitempty"`
	Zh       []string  `json:"zh,omitempty"`
	Senses   []string  `json:"senses,omitempty"`
	Stage    string    `json:"stage"`
	Added    time.Time `json:"added"`
	Contexts []Context `json:"contexts,omitempty"`
	Right    int       `json:"right"`
	Wrong    int       `json:"wrong"`

	LastReview string `json:"lastReview,omitempty"`
	Due        string `json:"due,omitempty"`
}

// File is the whole word list, as it is written to disk.
type File struct {
	Version int    `json:"version"`
	Words   []Word `json:"words"`
}

// Dir is where vocab.json lives. The desk points it at the same directory as
// state.json; tests and the screenshot tool give it a temporary one.
var Dir = defaultDir()

func defaultDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(base, "bible-en-study")
}

func Path() string { return filepath.Join(Dir, "vocab.json") }

// Empty is a word list with nothing in it.
func Empty() *File { return &File{Version: Version, Words: []Word{}} }

// Load reads the word list. A missing file is a first run, not an error. A
// damaged file is reported and left untouched: the desk keeps an empty list in
// memory and does not write over it, so the words can still be recovered by
// hand.
func Load() (*File, error) {
	b, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return Empty(), nil
		}
		return Empty(), err
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return Empty(), err
	}
	if f.Version > Version {
		return Empty(), errors.New("vocab.json is from a newer version")
	}
	if f.Version < 1 {
		f.Version = Version
	}
	if f.Words == nil {
		f.Words = []Word{}
	}
	return &f, nil
}

// Save writes the list through a temporary file in the same directory, then
// renames it into place: a crash mid-write leaves the old list whole.
func Save(f *File) error {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return err
	}
	f.Version = Version
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(Dir, "vocab-*.json")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	return os.Rename(name, Path())
}

// Find is the index of a word, matched without case, or -1.
func (f *File) Find(word string) int {
	want := strings.ToLower(strings.TrimSpace(word))
	if want == "" {
		return -1
	}
	for i, w := range f.Words {
		if strings.ToLower(w.Word) == want {
			return i
		}
	}
	return -1
}

// Add saves a word and returns its index. A word already in the list keeps its
// stage and counts; a new verse is added to its contexts, an old one is not
// repeated. The surface form is whatever the reader met, not the lemma.
func (f *File) Add(w Word) int {
	w.Word = strings.TrimSpace(w.Word)
	if i := f.Find(w.Word); i >= 0 {
		if len(w.Contexts) > 0 {
			f.addContext(i, w.Contexts[0])
		}
		return i
	}
	if w.Stage == "" {
		w.Stage = StageNew
	}
	if w.Added.IsZero() {
		w.Added = time.Now()
	}
	f.Words = append(f.Words, w)
	return len(f.Words) - 1
}

// AddContext records one more place a saved word was met.
func (f *File) AddContext(word string, c Context) bool {
	i := f.Find(word)
	if i < 0 {
		return false
	}
	return f.addContext(i, c)
}

func (f *File) addContext(i int, c Context) bool {
	for _, have := range f.Words[i].Contexts {
		if have.Book == c.Book && have.Chapter == c.Chapter && have.Verse == c.Verse &&
			have.Translation == c.Translation {
			return false
		}
	}
	f.Words[i].Contexts = append(f.Words[i].Contexts, c)
	return true
}

// Remove deletes a word outright and reports whether it was there.
func (f *File) Remove(word string) bool {
	i := f.Find(word)
	if i < 0 {
		return false
	}
	f.Words = append(f.Words[:i], f.Words[i+1:]...)
	return true
}

// SetStage moves a word between the three stages.
func (f *File) SetStage(word, stage string) bool {
	i := f.Find(word)
	if i < 0 {
		return false
	}
	switch stage {
	case StageNew, StageLearning, StageKnown:
		f.Words[i].Stage = stage
		return true
	}
	return false
}

// Known reports whether a word is saved.
func (f *File) Known(word string) bool { return f.Find(word) >= 0 }

// Count is how many words are saved.
func (f *File) Count() int { return len(f.Words) }
