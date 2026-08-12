package widget

// [bt-only] BibleText fork invariant: the Entry caret must blink DISCRETELY —
// snapping between exactly two fill colours (dim, opaque) — never through a
// smooth alpha fade. The stock fade calls cursor.Refresh() (a full-canvas
// repaint; a complete GL re-stream on mobile) on every animation frame inside
// its fade band, ~8 repaints/s while any Entry has focus, which burned 30-60%
// CPU on an idle focused field on iOS. A rebase that reverts the fork's
// discrete-blink commit reintroduces intermediate alphas and fails here.

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/canvas"
	_ "fyne.io/fyne/v2/test"
)

func TestBTCaretBlinkIsDiscrete(t *testing.T) {
	cursor := canvas.NewRectangle(color.Black)
	a := newEntryCursorAnimation(cursor)
	a.start()
	defer a.stop()

	// 16-bit premultiplied alpha via RGBA() — the same call the production
	// callback gates on. (internal/color.ToNRGBA has an 8-bit fast path for
	// NRGBA values, so the >>8 idiom from the upstream test loses the alpha.)
	seen := map[uint32]bool{}
	transitions := 0
	var prev uint32
	havePrev := false
	for i := 0; i <= 100; i++ {
		a.anim.Tick(float32(i) / 100)
		_, _, _, al := a.cursor.FillColor.RGBA()
		seen[al] = true
		if havePrev && prev != al {
			transitions++
		}
		prev, havePrev = al, true
	}

	if len(seen) > 2 {
		t.Fatalf("caret swept %d distinct alphas %v — the smooth fade (and its ~8 full-canvas repaints/s) is back", len(seen), keys(seen))
	}
	if len(seen) < 2 {
		t.Fatalf("caret never changed alpha (stuck at %v) — no blink at all", keys(seen))
	}
	if transitions != 1 {
		t.Errorf("caret changed alpha %d times across one half-cycle, want exactly 1 (a single dim→opaque snap)", transitions)
	}
}

func keys(m map[uint32]bool) []uint32 {
	out := make([]uint32, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
