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
	cursorInterruptDuration = 300 * time.Millisecond
	cursorFadeAlpha         = uint8(0x16)
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
func (a *entryCursorAnimation) createAnim() *fyne.Animation {
	r, g, b, opaqueAlpha := col.ToNRGBA(theme.Color(theme.ColorNamePrimary))
	opaqueColor := color.NRGBA{R: r, G: g, B: b, A: opaqueAlpha}
	endColor := color.NRGBA{R: r, G: g, B: b, A: cursorFadeAlpha}
	startColor := opaqueColor
	a.cursor.FillColor = startColor

	interrupted := false
	anim := fyne.NewAnimation(time.Second/2, func(f float32) {
		if timeNow().Sub(a.lastInterruptTime) < cursorInterruptDuration {
			if !interrupted {
				a.cursor.FillColor = opaqueColor
				a.cursor.Refresh()
				interrupted = true
			}
			return
		}

		if interrupted {
			interrupted = false
			// stop and start effectively restarts animation from the beginning
			a.anim.Stop()
			a.anim.Start()
			return
		}

		// BibleText patch: discrete caret blink. The smooth fade calls
		// cursor.Refresh() — a FULL-canvas repaint (tree walk + complete GL
		// re-stream on mobile) — on every animation frame inside its fade band,
		// for as long as any Entry has focus. On iOS that burned 30-60% CPU
		// with an idle focused field. Snapping at the half-cycle keeps the
		// blink cadence and the typing-interrupt behaviour but refreshes only
		// twice per second.
		if f < 0.5 {
			if a.cursor.FillColor == startColor {
				return
			}

			a.cursor.FillColor = startColor
		} else {
			if a.cursor.FillColor == endColor {
				return
			}

			a.cursor.FillColor = endColor
		}

		a.cursor.Refresh()
	})

	anim.RepeatCount = fyne.AnimationRepeatForever
	anim.AutoReverse = true
	return anim
}

// starts cursor animation.
func (a *entryCursorAnimation) start() {
	if a.anim == nil {
		a.anim = a.createAnim()
		a.anim.Start()
	}
}

// Stops the animation for cursorInterruptDuration.
// This is used to keep the cursor visible while typing.
func (a *entryCursorAnimation) interrupt() {
	a.lastInterruptTime = timeNow()
}

func (a *entryCursorAnimation) stop() {
	if a.anim != nil {
		a.anim.Stop()
		a.anim = nil
	}
}
