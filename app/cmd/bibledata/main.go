// Command bibledata turns the source SQLite database into the compact store the
// app embeds. It never ships the database itself: the app reads store.bin.
//
//	go run ./cmd/bibledata -db ../../data/db/bible-en-study.db -out ../bible/store.bin
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
)

type bookRow struct {
	ID        int    `json:"book_id"`
	OSIS      string `json:"osis_id"`
	Name      string `json:"name"`
	Testament string `json:"testament"`
	Section   string `json:"section"`
	Chapters  int    `json:"chapter_count"`
}

type transRow struct {
	ID        string `json:"translation_id"`
	Name      string `json:"name"`
	Abbr      string `json:"abbreviation"`
	Language  string `json:"language_code"`
	LangName  string `json:"language_name"`
	Edition   string `json:"edition"`
	License   string `json:"license"`
	LicenseUL string `json:"license_url"`
	Books     int    `json:"book_count"`
	Chapters  int    `json:"chapter_count"`
	Verses    int    `json:"verse_count"`
}

type titleRow struct {
	Translation string `json:"translation_id"`
	Book        int    `json:"book_id"`
	Title       string `json:"title"`
}

type verseRow struct {
	Translation string `json:"translation_id"`
	Book        int    `json:"book_id"`
	Chapter     int    `json:"chapter"`
	Verse       int    `json:"verse"`
	Text        string `json:"text"`
	Para        int    `json:"paragraph_start"`
}

type sectionRow struct {
	Translation string `json:"translation_id"`
	Book        int    `json:"book_id"`
	Chapter     int    `json:"chapter"`
	Verse       int    `json:"verse"`
	Ordinal     int    `json:"ordinal"`
	Kind        string `json:"heading_type"`
	Level       int    `json:"level"`
	Text        string `json:"text"`
}

func main() {
	db := flag.String("db", "../../data/db/bible-en-study.db", "source SQLite database")
	out := flag.String("out", "internal/bible/store.bin", "store to write")
	flag.Parse()

	books := query[bookRow](*db, `select book_id, osis_id, name, testament, section, chapter_count
		from books order by book_id`)
	trans := query[transRow](*db, `select translation_id, name, abbreviation, language_code, language_name,
		edition, license, coalesce(license_url,'') as license_url, book_count, chapter_count, verse_count
		from translations order by translation_id`)
	titles := query[titleRow](*db, `select translation_id, book_id, coalesce(title,'') as title
		from translation_books order by translation_id, book_id`)
	verses := query[verseRow](*db, `select translation_id, book_id, chapter, verse, text, paragraph_start
		from verses order by translation_id, book_id, chapter, verse`)
	sections := query[sectionRow](*db, `select translation_id, book_id, chapter, verse, ordinal,
		heading_type, level, text from sections order by translation_id, book_id, chapter, verse, ordinal`)

	byOSIS := make(map[int]string, len(books))
	for _, b := range books {
		byOSIS[b.ID] = b.OSIS
	}

	type chapterKey struct {
		tr, osis string
		ch       int
	}
	chapters := make(map[chapterKey]*wireChapter)
	for _, v := range verses {
		k := chapterKey{v.Translation, byOSIS[v.Book], v.Chapter}
		c := chapters[k]
		if c == nil {
			c = &wireChapter{}
			chapters[k] = c
		}
		p := 0
		if v.Para == 1 {
			p = 1
		}
		c.V = append(c.V, []any{v.Verse, v.Text, p})
	}
	for _, h := range sections {
		k := chapterKey{h.Translation, byOSIS[h.Book], h.Chapter}
		c := chapters[k]
		if c == nil {
			continue // a heading without a verse row in this chapter
		}
		c.H = append(c.H, []any{h.Verse, h.Level, h.Kind, h.Text})
	}

	keys := make([]chapterKey, 0, len(chapters))
	for k := range chapters {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].tr != keys[j].tr {
			return keys[i].tr < keys[j].tr
		}
		if keys[i].osis != keys[j].osis {
			return keys[i].osis < keys[j].osis
		}
		return keys[i].ch < keys[j].ch
	})

	var payload bytes.Buffer
	index := map[string]any{}
	spans := make(map[string][2]int, len(keys))
	for _, k := range keys {
		var buf bytes.Buffer
		zw, _ := zlib.NewWriterLevel(&buf, zlib.BestCompression)
		if err := json.NewEncoder(zw).Encode(chapters[k]); err != nil {
			log.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			log.Fatal(err)
		}
		off := payload.Len()
		payload.Write(buf.Bytes())
		spans[fmt.Sprintf("%s|%s|%d", k.tr, k.osis, k.ch)] = [2]int{off, buf.Len()}
	}

	titlesOut := make(map[string]string, len(titles))
	for _, t := range titles {
		if t.Title != "" {
			titlesOut[t.Translation+"|"+byOSIS[t.Book]] = t.Title
		}
	}
	index["books"] = toBookJSON(books)
	index["translations"] = toTransJSON(trans)
	index["titles"] = titlesOut
	index["chapters"] = spans

	ix, err := json.Marshal(index)
	if err != nil {
		log.Fatal(err)
	}
	var file bytes.Buffer
	file.WriteString("BIBL1\n")
	if err := binary.Write(&file, binary.LittleEndian, uint32(len(ix))); err != nil {
		log.Fatal(err)
	}
	file.Write(ix)
	file.Write(payload.Bytes())

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, file.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("books %d · translations %d · chapters %d · verses %d · sections %d\n",
		len(books), len(trans), len(keys), len(verses), len(sections))
	fmt.Printf("index %s · payload %s · store %s\n",
		human(len(ix)), human(payload.Len()), human(file.Len()))
}

type wireChapter struct {
	V [][]any `json:"v"`
	H [][]any `json:"h"`
}

func toBookJSON(rows []bookRow) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, b := range rows {
		out = append(out, map[string]any{
			"id": b.ID, "osis": b.OSIS, "name": b.Name,
			"testament": b.Testament, "section": b.Section, "chapters": b.Chapters,
		})
	}
	return out
}

func toTransJSON(rows []transRow) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, t := range rows {
		out = append(out, map[string]any{
			"id": t.ID, "name": t.Name, "abbr": t.Abbr,
			"language": t.Language, "language_name": t.LangName,
			"edition": t.Edition, "license": t.License, "license_url": t.LicenseUL,
			"books": t.Books, "chapters": t.Chapters, "verses": t.Verses,
		})
	}
	return out
}

// query runs a read-only sqlite3 query and decodes its JSON rows. The generator
// is a build-time tool, so the CLI is enough and the app needs no driver.
func query[T any](db, sql string) []T {
	cmd := exec.Command("sqlite3", "-readonly", "-json", db, sql)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("sqlite3: %v: %s", err, stderr.String())
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil
	}
	var rows []T
	if err := json.Unmarshal(out, &rows); err != nil {
		log.Fatalf("decode rows: %v", err)
	}
	return rows
}

func human(n int) string {
	switch {
	case n > 1<<20:
		return strconv.FormatFloat(float64(n)/(1<<20), 'f', 1, 64) + " MB"
	case n > 1<<10:
		return strconv.FormatInt(int64(n)/(1<<10), 10) + " KB"
	}
	return strconv.Itoa(n) + " B"
}
