// Package bible reads the canon from a compact store built out of the source
// database by cmd/bibledata. One chapter is decoded on demand, so the whole
// text never sits in memory and the app stays a single file.
package bible

import (
	"bytes"
	"compress/zlib"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

//go:embed store.bin
var raw []byte

// Store is the decoded index plus the chapter payload.
type Store struct {
	books    []Book
	byOSIS   map[string]Book
	byID     map[int]Book
	trans    []Translation
	titles   map[string]string
	chapters map[string][2]int
	payload  []byte
	cache    map[string]Chapter
}

type Book struct {
	ID        int    `json:"id"`
	OSIS      string `json:"osis"`
	Name      string `json:"name"`
	Testament string `json:"testament"`
	Section   string `json:"section"`
	Chapters  int    `json:"chapters"`
}

type Translation struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbr         string `json:"abbr"`
	Language     string `json:"language"`
	LanguageName string `json:"language_name"`
	Edition      string `json:"edition"`
	License      string `json:"license"`
	LicenseURL   string `json:"license_url"`
	Books        int    `json:"books"`
	Chapters     int    `json:"chapters"`
	Verses       int    `json:"verses"`
}

// InLanguage lists the translations of one language code ("en", "cmn"), in the
// order the dataset stores them.
func (s *Store) InLanguage(code string) []Translation {
	var out []Translation
	for _, t := range s.trans {
		if t.Language == code {
			out = append(out, t)
		}
	}
	return out
}

// IsChinese reports whether the translation's text is Chinese.
func (t Translation) IsChinese() bool { return t.Language == "cmn" }

// Verse keeps the translation's own empty verses as empty text rather than
// dropping them, so verse numbering stays comparable across translations.
type Verse struct {
	N    int    `json:"n"`
	Text string `json:"t"`
	Para bool   `json:"p"`
}

// Heading is a section heading printed above a verse.
type Heading struct {
	Verse int    `json:"v"`
	Level int    `json:"l"`
	Kind  string `json:"k"`
	Text  string `json:"x"`
}

type Chapter struct {
	Verses   []Verse
	Headings []Heading
}

type index struct {
	Books        []Book            `json:"books"`
	Translations []Translation     `json:"translations"`
	Titles       map[string]string `json:"titles"`
	Chapters     map[string][2]int `json:"chapters"`
}

var magic = []byte("BIBL1\n")

// Open decodes the embedded index.
func Open() (*Store, error) {
	if len(raw) < len(magic)+4 || !bytes.HasPrefix(raw, magic) {
		return nil, errors.New("bible: store.bin has no header")
	}
	n := int(binary.LittleEndian.Uint32(raw[len(magic):]))
	start := len(magic) + 4
	if n <= 0 || start+n > len(raw) {
		return nil, errors.New("bible: store.bin is truncated")
	}
	var ix index
	if err := json.Unmarshal(raw[start:start+n], &ix); err != nil {
		return nil, fmt.Errorf("bible: index: %w", err)
	}
	s := &Store{
		books:    ix.Books,
		byOSIS:   make(map[string]Book, len(ix.Books)),
		byID:     make(map[int]Book, len(ix.Books)),
		trans:    ix.Translations,
		titles:   ix.Titles,
		chapters: ix.Chapters,
		payload:  raw[start+n:],
		cache:    make(map[string]Chapter, 8),
	}
	for _, b := range s.books {
		s.byOSIS[b.OSIS] = b
		s.byID[b.ID] = b
	}
	return s, nil
}

func (s *Store) Books() []Book               { return s.books }
func (s *Store) Translations() []Translation { return s.trans }

func (s *Store) Book(osis string) (Book, bool) {
	b, ok := s.byOSIS[osis]
	return b, ok
}

func (s *Store) Translation(id string) (Translation, bool) {
	for _, t := range s.trans {
		if t.ID == id {
			return t, true
		}
	}
	return Translation{}, false
}

// Title is the book's own running head in that translation ("Psalm", "Psalms").
func (s *Store) Title(tr, osis string) string {
	if v, ok := s.titles[tr+"|"+osis]; ok && v != "" {
		return v
	}
	if b, ok := s.byOSIS[osis]; ok {
		return b.Name
	}
	return osis
}

// Chapter decodes one chapter, or reports that the store has no such entry.
func (s *Store) Chapter(tr, osis string, ch int) (Chapter, bool) {
	if ch < 1 {
		return Chapter{}, false
	}
	key := fmt.Sprintf("%s|%s|%d", tr, osis, ch)
	if c, ok := s.cache[key]; ok {
		return c, true
	}
	sp, ok := s.chapters[key]
	if !ok || sp[0] < 0 || sp[0]+sp[1] > len(s.payload) {
		return Chapter{}, false
	}
	var wire struct {
		V [][3]any `json:"v"`
		H [][4]any `json:"h"`
	}
	r, zerr := zlib.NewReader(bytes.NewReader(s.payload[sp[0] : sp[0]+sp[1]]))
	if zerr != nil {
		return Chapter{}, false
	}
	defer r.Close()
	if err := json.NewDecoder(r).Decode(&wire); err != nil {
		return Chapter{}, false
	}
	c := Chapter{Verses: make([]Verse, 0, len(wire.V)), Headings: make([]Heading, 0, len(wire.H))}
	for _, row := range wire.V {
		text, _ := row[1].(string)
		c.Verses = append(c.Verses, Verse{N: num(row[0]), Text: text, Para: num(row[2]) == 1})
	}
	for _, row := range wire.H {
		kind, _ := row[2].(string)
		text, _ := row[3].(string)
		c.Headings = append(c.Headings, Heading{Verse: num(row[0]), Level: num(row[1]), Kind: kind, Text: text})
	}
	if len(s.cache) > 16 {
		s.cache = make(map[string]Chapter, 8)
	}
	s.cache[key] = c
	return c, true
}

// Sections lists the section keys in canonical order, with the Chinese label
// the interface shows and the testaments they sit in.
type Section struct {
	Key   string
	ZH    string
	Books []Book
}

var sectionZH = map[string]string{
	"Law":              "律法书",
	"History":          "历史书",
	"Wisdom":           "诗歌智慧书",
	"Major Prophets":   "大先知书",
	"Minor Prophets":   "小先知书",
	"Gospels":          "福音书",
	"Acts":             "使徒行传",
	"Pauline Epistles": "保罗书信",
	"General Epistles": "普通书信",
	"Apocalyptic":      "启示录",
}

var sectionOrder = []string{
	"Law", "History", "Wisdom", "Major Prophets", "Minor Prophets",
	"Gospels", "Acts", "Pauline Epistles", "General Epistles", "Apocalyptic",
}

// Sections groups the books the way a printed Bible does.
func (s *Store) Sections() []Section {
	byKey := make(map[string][]Book, len(sectionOrder))
	for _, b := range s.books {
		byKey[b.Section] = append(byKey[b.Section], b)
	}
	out := make([]Section, 0, len(sectionOrder))
	for _, k := range sectionOrder {
		if books, ok := byKey[k]; ok {
			out = append(out, Section{Key: k, ZH: label(k), Books: books})
		}
	}
	return out
}

func label(key string) string {
	if zh, ok := sectionZH[key]; ok {
		return zh
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

func num(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
