// Command shot paints the app headlessly and writes a screenshot, so a change
// can be looked at without opening a window. Steps after the first path drive
// the app first:
//
//	go run -tags vivahead ./cmd/shot /tmp/a.png "label:英文" "shot:/tmp/b.png" "locate:created" "shot:/tmp/c.png"
//
// Steps: click:X,Y  label:文字  locate:文字  type:text
// size:w,h  keys:spec  wait:ms  shot:path
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vivaui/viva/headless"
	"github.com/vivaui/viva/ui"

	"app/internal/desk"
	"app/internal/store"
)

func main() {
	out := "/tmp/viva.png"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	// A screenshot must not move the reader's own place on disk.
	store.Dir = filepath.Join(os.TempDir(), "bible-shot")
	if err := desk.Load(); err != nil {
		log.Fatal(err)
	}
	session, err := headless.Open(desk.Config(desk.New()), nil, headless.Options{})
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	if err := session.Screenshot(out); err != nil {
		log.Fatal(err)
	}
	fmt.Println(out)

	for _, step := range os.Args[2:] {
		kind, arg, _ := strings.Cut(step, ":")
		switch kind {
		case "click":
			x, y, ok := strings.Cut(arg, ",")
			if !ok {
				log.Fatalf("click needs x,y: %q", step)
			}
			fx, _ := strconv.ParseFloat(x, 32)
			fy, _ := strconv.ParseFloat(y, 32)
			if err := session.Click(float32(fx), float32(fy)); err != nil {
				log.Fatal(err)
			}
		case "label":
			if err := session.ClickLabel(arg); err != nil {
				log.Fatal(err)
			}
		case "locate":
			hit, err := session.Locate(arg)
			if err != nil {
				log.Fatalf("locate %q: %v", arg, err)
			}
			b := hit.Bounds
			if err := session.Click(b.X+b.Width/2, b.Y+b.Height/2); err != nil {
				log.Fatal(err)
			}
		case "size":
			w, h, ok := strings.Cut(arg, ",")
			if !ok {
				log.Fatalf("size needs w,h: %q", step)
			}
			fw, _ := strconv.Atoi(w)
			fh, _ := strconv.Atoi(h)
			if err := session.Driver().Resize(fw, fh); err != nil {
				log.Fatal(err)
			}
		case "type":
			if err := session.Type(arg); err != nil {
				log.Fatal(err)
			}
		case "keys":
			if err := session.Keys(arg); err != nil {
				log.Fatal(err)
			}
		case "wait":
			ms, err := strconv.Atoi(arg)
			if err != nil {
				log.Fatalf("wait needs milliseconds: %q", step)
			}
			if err := session.Wait(time.Duration(ms) * time.Millisecond); err != nil {
				log.Fatal(err)
			}
		case "key":
			if err := session.Driver().Key(ui.KeyRight, 0); err != nil {
				log.Fatal(err)
			}
		case "advance":
			ms, err := strconv.Atoi(arg)
			if err != nil {
				log.Fatalf("advance needs milliseconds: %q", step)
			}
			if err := session.Driver().Advance(time.Duration(ms) * time.Millisecond); err != nil {
				log.Fatal(err)
			}
		case "shot":
			if err := session.Screenshot(arg); err != nil {
				log.Fatal(err)
			}
			fmt.Println(arg)
		default:
			log.Fatalf("unknown step %q", step)
		}
	}
}
