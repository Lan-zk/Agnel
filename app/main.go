package main

import (
	"log"

	"github.com/vivaui/viva/app/iconpng"

	"app/assets"
	"app/internal/desk"
)

func main() {
	icon, err := iconpng.Decode(assets.IconPNG)
	if err != nil {
		log.Fatal(err)
	}
	desk.Run(icon)
}
