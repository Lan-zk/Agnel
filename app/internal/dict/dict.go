// Package dict is the offline English dictionary that the reading card uses.
//
// The data is built by data/lex/build_dict.py from Open Dictionary, WordNet
// 3.0, the CMU Pronouncing Dictionary, CC-CEDICT and ACAI, and embedded as
// dict.bin: a zlib-compressed JSON index (word list, inflected forms, offsets)
// followed by one zlib-compressed JSON record per word. Looking a word up is a
// string compare and a small inflate, so the window never waits on disk.
package dict

import (
	"bytes"
	"compress/zlib"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"
)

// Entry is one word as the card shows it. A leading form of another word only
// carries Word and Lemma; Lookup resolves it so the card can say "created 是
// create 的过去式" and still show the meaning.
type Entry struct {
	Word         string   `json:"w"`
	Pos          string   `json:"pos"`
	IPA          string   `json:"uk"`
	Senses       []string `json:"s"`
	Zh           []string `json:"zh"`
	Collocations []string `json:"c"`
	Rank         int      `json:"r"`
	Verses       int      `json:"n"`
	Lemma        string   `json:"lemma"`
	Name         string   `json:"name"`
}

//go:embed dict.bin
var blob []byte

type index struct {
	Words   []string          `json:"words"`
	Forms   map[string]string `json:"forms"`
	Entries map[string][2]int `json:"entries"`
}

var (
	data  index
	body  []byte
	ready bool
)

// Load reads the embedded index. It is safe to call more than once.
func Load() error {
	if ready {
		return nil
	}
	if len(blob) < 10 {
		return io.ErrUnexpectedEOF
	}
	n := binary.LittleEndian.Uint32(blob[6:10])
	if int(n) > len(blob)-10 {
		return io.ErrUnexpectedEOF
	}
	var raw []byte
	switch string(blob[:6]) {
	case "LEXI1\n":
		raw = blob[10 : 10+n]
	case "LEXI2\n":
		r, err := zlib.NewReader(bytes.NewReader(blob[10 : 10+n]))
		if err != nil {
			return err
		}
		raw, err = io.ReadAll(r)
		r.Close()
		if err != nil {
			return err
		}
	default:
		return io.ErrUnexpectedEOF
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	body = blob[10+n:]
	ready = true
	return nil
}

// Words is the number of headwords, for the about line.
func Words() int { return len(data.Words) }

// Lookup finds the entry for a word as it appears in the text: exact, then
// lowercased, then through the inflected-form index. An inflected form is
// returned as the entry it belongs to, with Lemma set to that entry's word.
func Lookup(word string) (Entry, bool) {
	if !ready {
		return Entry{}, false
	}
	clean := strings.ToLower(strings.Trim(word, ".,;:!?()[]{}\"'“”‘’—-"))
	// The text uses typographic apostrophes; the dictionary keys use ASCII.
	clean = strings.ReplaceAll(clean, "’", "'")
	clean = strings.Trim(clean, "'")
	if clean == "" {
		return Entry{}, false
	}
	if e, ok := at(clean); ok {
		// An inflected form has its own small record (how it is pronounced,
		// which word it belongs to); the meaning comes from that word.
		if e.Lemma != "" {
			if lemma, ok := at(e.Lemma); ok {
				if e.IPA != "" {
					lemma.IPA = e.IPA
				}
				if e.Verses > 0 {
					lemma.Verses = e.Verses
				}
				lemma.Word, lemma.Lemma = clean, e.Lemma
				return lemma, true
			}
		}
		return e, true
	}
	if lemma, ok := data.Forms[clean]; ok {
		if e, ok := at(lemma); ok {
			form, _ := atPointer(clean)
			e.Word, e.Lemma = clean, lemma
			if form.IPA != "" {
				e.IPA = form.IPA
			}
			return e, true
		}
	}
	// "God's", "David's": the possessive is the base word with a grammatical
	// ending, and the card can show the base meaning without inventing a
	// separate entry.
	if strings.HasSuffix(clean, "'s") {
		if e, ok := Lookup(clean[:len(clean)-2]); ok {
			e.Word, e.Lemma = clean, ""
			return e, true
		}
	}
	// A plural or a participle the form index missed: try the usual trims.
	for _, cut := range []string{"s", "es", "ed", "ing", "d"} {
		if !strings.HasSuffix(clean, cut) || len(clean) <= len(cut)+2 {
			continue
		}
		stem := strings.TrimSuffix(clean, cut)
		for _, cand := range []string{stem, stem + "e", stem + stem[len(stem)-1:]} {
			if e, ok := at(cand); ok {
				e.Lemma = cand
				return e, true
			}
		}
	}
	return Entry{}, false
}

// atPointer reads the small record an inflected form carries on its own: how
// it is pronounced, and which word it belongs to.
func atPointer(word string) (Entry, bool) {
	span, ok := data.Entries[word]
	if !ok || span[1] <= 0 {
		return Entry{}, false
	}
	r, err := zlib.NewReader(bytes.NewReader(body[span[0] : span[0]+span[1]]))
	if err != nil {
		return Entry{}, false
	}
	defer r.Close()
	var e Entry
	if err := json.NewDecoder(r).Decode(&e); err != nil {
		return Entry{}, false
	}
	return e, true
}

func at(word string) (Entry, bool) {
	span, ok := data.Entries[word]
	if !ok {
		return Entry{}, false
	}
	if span[0] < 0 || span[1] <= 0 || span[0]+span[1] > len(body) {
		return Entry{}, false
	}
	r, err := zlib.NewReader(bytes.NewReader(body[span[0] : span[0]+span[1]]))
	if err != nil {
		return Entry{}, false
	}
	defer r.Close()
	var e Entry
	if err := json.NewDecoder(r).Decode(&e); err != nil {
		return Entry{}, false
	}
	if e.Word == "" {
		e.Word = word
	}
	return e, true
}
