//go:build !no_emoji

package theme

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

// BibleText patch: current emoji. EmojiOneColor is a ~2016 set; anything from
// Emoji 11.0 (2018) on — 🥺 🤏 🤌 🫶 — has no glyph and drew as a notdef box.
// Noto Color Emoji is current; OFL 1.1, licence shipped beside it as
// LICENSE_NotoColorEmoji.txt.
//
//go:embed font/NotoColorEmoji.ttf
var emojiFontData []byte

var emoji = &fyne.StaticResource{
	StaticName:    "NotoColorEmoji.ttf",
	StaticContent: emojiFontData,
}
