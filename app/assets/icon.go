// Package assets holds the application's pictures.
package assets

import _ "embed"

// IconPNG is the application icon, the same file viva.toml names as [app] icon.
// A program run without an .app bundle decodes it for its Dock or taskbar icon;
// a bundle keeps the icon it was packaged with.
//
//go:embed icon.png
var IconPNG []byte
