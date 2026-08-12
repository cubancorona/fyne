//go:build !no_emoji

package theme

// [bt-only] BibleText fork invariant: the bundled emoji font must cover
// modern emoji. The stock EmojiOneColor.otf is a ~2016 set — anything from
// Emoji 11.0 (2018) on (🥺 🤏 🤌 🫶) has no glyph and drew as a notdef box in
// Entry widgets. A rebase that reverts the fork's Noto Color Emoji commit
// fails here on the cmap lookups.

import (
	"bytes"
	"testing"

	"github.com/go-text/typesetting/font"
)

func TestBTEmojiFontCoversModernEmoji(t *testing.T) {
	if emoji.StaticName != "NotoColorEmoji.ttf" {
		t.Errorf("bundled emoji resource is %q, want NotoColorEmoji.ttf", emoji.StaticName)
	}
	face, err := font.ParseTTF(bytes.NewReader(emojiFontData))
	if err != nil {
		t.Fatalf("bundled emoji font does not parse: %v", err)
	}
	for _, r := range []rune{
		0x1F90F, // 🤏 pinching hand, Emoji 11.0
		0x1F97A, // 🥺 pleading face, Emoji 11.0
		0x1F90C, // 🤌 pinched fingers, Emoji 13.0
		0x1FAF6, // 🫶 heart hands, Emoji 14.0
	} {
		if gid, ok := face.Cmap.Lookup(r); !ok || gid == 0 {
			t.Errorf("emoji %c (U+%X) has no glyph in the bundled font — modern emoji would draw as a notdef box", r, r)
		}
	}
}
