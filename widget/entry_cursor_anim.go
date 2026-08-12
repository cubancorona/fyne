package widget

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	col "fyne.io/fyne/v2/internal/color"
	"fyne.io/fyne/v2/theme"
)

var timeNow = time.Now // used in tests

const (
	cursorInterruptTime = 300 * time.Millisecond
	cursorFadeAlpha     = uint8(0x16)
	cursorFadeRatio     = 0.1
)

type entryCursorAnimation struct {
	cursor            *canvas.Rectangle
	anim              *fyne.Animation
	lastInterruptTime time.Time
}

func newEntryCursorAnimation(cursor *canvas.Rectangle) *entryCursorAnimation {
	return &entryCursorAnimation{cursor: cursor}
}

// creates fyne animation
func (a *entryCursorAnimation) createAnim(inverted bool) *fyne.Animation {
	cursorOpaque := theme.Color(theme.ColorNamePrimary)
	ri, gi, bi, _ := col.ToNRGBA(cursorOpaque)
	r := uint8(ri >> 8)
	g := uint8(gi >> 8)
	b := uint8(bi >> 8)
	cursorDim := color.NRGBA{R: r, G: g, B: b, A: cursorFadeAlpha}
	if inverted {
		a.cursor.FillColor = cursorOpaque
	} else {
		a.cursor.FillColor = cursorDim
	}

	interrupted := false
	anim := fyne.NewAnimation(time.Second/2, func(f float32) {
		shouldInterrupt := timeNow().Sub(a.lastInterruptTime) <= cursorInterruptTime
		if shouldInterrupt {
			if !interrupted {
				a.cursor.FillColor = cursorOpaque
				a.cursor.Refresh()
				interrupted = true
			}
			return
		}
		if interrupted {
			a.anim.Stop()
			if !inverted {
				a.anim = a.createAnim(true)
			}
			interrupted = false
			canStart := a.anim != nil
			if canStart {
				a.anim.Start()
			}
			return
		}

		// BibleText patch: discrete caret blink. The stock smooth fade calls
		// cursor.Refresh() — a FULL-canvas repaint (tree walk + complete GL
		// re-stream on mobile) — on every animation frame inside its fade band,
		// ~8 repaints/s for as long as any Entry has focus. On iOS that burned
		// 30-60% CPU with an idle focused field. Snapping between dim and opaque
		// at the half-cycle keeps the blink cadence and the typing-interrupt
		// behaviour but refreshes only twice per second.
		opaque := f >= 0.5
		if inverted {
			opaque = f < 0.5
		}
		if opaque {
			if _, _, _, al := a.cursor.FillColor.RGBA(); al == 0xffff {
				return
			}

			a.cursor.FillColor = cursorOpaque
		} else {
			if _, _, _, al := a.cursor.FillColor.RGBA(); uint8(al>>8) == cursorFadeAlpha {
				return
			}

			a.cursor.FillColor = cursorDim
		}

		a.cursor.Refresh()
	})

	anim.RepeatCount = fyne.AnimationRepeatForever
	anim.AutoReverse = true
	return anim
}

// starts cursor animation.
func (a *entryCursorAnimation) start() {
	isStopped := a.anim == nil
	if isStopped {
		a.anim = a.createAnim(false)
		a.anim.Start()
	}
}

// temporarily stops the animation by "cursorInterruptTime".
func (a *entryCursorAnimation) interrupt() {
	a.lastInterruptTime = timeNow()
}

// stops cursor animation.
func (a *entryCursorAnimation) stop() {
	if a.anim != nil {
		a.anim.Stop()
		a.anim = nil
	}
}
