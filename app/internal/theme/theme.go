// Package theme holds the colours and type tokens of the reading app.
package theme

import (
	"github.com/vivaui/viva/ui"

	"app/internal/font"
)

var (
	Bg        = ui.RGB24(0xdde2e6)
	Surface   = ui.RGB24(0xeef1f4)
	Surface2  = ui.RGB24(0xe2e7ea)
	DeskTrack = ui.RGB24(0xd2d8dc)
	DeskEdge  = ui.RGB24(0xc3cbd0)
	Card      = ui.RGB24(0xfbf9f3)
	CardRule  = ui.RGB24(0xd9d2c5)
	CardInset = ui.RGB24(0xf2ede2)
	Ink       = ui.RGB24(0x181c1f)
	InkSoft   = ui.RGB24(0x43484c)
	InkFaint  = ui.RGB24(0x63696e)
	Rule      = ui.RGB24(0xc9d2d6)
	RuleSoft  = ui.RGB24(0xd5dce0)
	Accent    = ui.RGB24(0x1a6560)
	AccentBg  = ui.RGB24(0xdce8e6)
	OnAccent  = ui.RGB24(0xffffff)
	Danger    = ui.RGB24(0xb83228)
	Secondary = ui.RGB24(0x552a30)

	Paper       = ui.RGB24(0xf4f1e8)
	PaperWarm   = ui.RGB24(0xede8dc)
	PaperStack  = ui.RGB24(0xe4decf)
	PaperBoard  = ui.RGB24(0xc8c1b2)
	PaperInk    = ui.RGB24(0x1f221e)
	PaperSoft   = ui.RGB24(0x5f615b)
	PaperSubtle = ui.RGB24(0x837f75)
	VerseNum    = ui.RGB24(0x8a7963)
	PaperRule   = ui.RGB24(0xd8d2c5)
	PaperAcc    = ui.RGB24(0x1a6560)
	PaperMark   = ui.RGB24(0xd3e2de)
	PaperHover  = ui.RGBA(26, 101, 96, 0.10)
	PaperShade  = ui.RGBA(18, 22, 20, 0.14)

	Ribbon     = ui.RGB24(0x9e3b33)
	RibbonDark = ui.RGB24(0x7a2a23)
)

// Type sizes, in logical pixels.
const (
	SizeMicro = 11
	SizeSmall = 12
	SizeBody  = 13.5
	SizeRead  = 15
	SizeLead  = 26.5 // 中文与英文对开页行高
	SizeUI    = 14
	SizeTitle = 21
	SizeBig   = 28
)

// FamilyBody names the embedded hybrid literary typeface ("Bible Literata":
// Literata + STIX Two IPA + Source Han Serif SC subset) registered at startup.
const (
	FamilyBody = font.Family
	FamilyMono = "Menlo"
)
