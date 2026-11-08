// Package store keeps the reader's place, its bookmarks and the words it saved.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Bookmark is a spot in the text: a chapter and the spread inside it.
type Bookmark struct {
	Book    string `json:"book"` // OSIS id
	Chapter int    `json:"chapter"`
	Spread  int    `json:"spread"`
}

// State is everything worth keeping between runs.
type State struct {
	Book      string     `json:"book"`
	Chapter   int        `json:"chapter"`
	Spread    int        `json:"spread"`
	Mode      int        `json:"mode"`
	Lang      int        `json:"lang"`
	Left      string     `json:"left"`
	Right     string     `json:"right"`
	Bookmarks []Bookmark `json:"bookmarks"`
	Recent    []string   `json:"recent"` // "Gen:1"
	Saved     []string   `json:"saved"`
}

// Dir is where state.json lives; tests and the screenshot tool point it at a
// temporary directory.
var Dir = defaultDir()

func defaultDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(base, "bible-en-study")
}

func Path() string { return filepath.Join(Dir, "state.json") }

// Load reads the state; a missing or damaged file is not an error, it is a
// first run.
func Load() (State, error) {
	var s State
	b, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, nil
	}
	return s, nil
}

func Save(s State) error {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), b, 0o644)
}
