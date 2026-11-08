package main

import (
	"testing"

	"github.com/vivaui/viva/app/iconpng"

	"app/assets"
	"app/internal/desk"
)

func TestTheBibleLoads(t *testing.T) {
	if err := desk.Load(); err != nil {
		t.Fatal(err)
	}
	if desk.Config(desk.New()).Paint == nil {
		t.Fatal("the window has no painter")
	}
}

func TestTheAppIconIsUsable(t *testing.T) {
	icon, err := iconpng.Decode(assets.IconPNG)
	if err != nil {
		t.Fatal(err)
	}
	if icon.Size < 512 {
		t.Fatalf("app icon is %d px wide; want at least 512", icon.Size)
	}
}
