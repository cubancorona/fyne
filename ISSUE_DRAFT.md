> **STATUS: FILED AND CLOSED.** This draft was filed as
> https://github.com/fyne-io/fyne/issues/6368 (2026-06-20) and closed the same
> day after maintainer feedback ("we do not support multiplexing with other
> toolkits"; on-demand rendering "may run parallel to" PR #5422). Do not
> re-file. The fix now lives in the cubancorona/fyne fork (bt-main); the open
> upstream path is an on-demand-rendering PR framed as draw-loop efficiency.
> See docs/FYNE_FORK_PLAN.md (bibletext repo), decision point 3.

Title: iOS: native UIKit views overlaid on the canvas scroll laggily — `drawloop` parks the main thread ~100ms per idle frame

### Describe the bug

On iOS, an app that floats a native UIKit scroll view (e.g. a `UITextView` or `UITableView`) above the Fyne canvas — a common pattern for native text selection/editing or large virtualized lists — scrolls visibly laggily, while Fyne's own scroll widgets (`widget.List`, `container.Scroll`) stay smooth. The native view's frame rate collapses even though it is doing almost no work. (Found while building an iOS reader that floats a `UITextView` over the canvas for native text selection.)

### Root cause

On iOS, Fyne renders on the **main thread**. `GoAppAppController` (`internal/driver/mobile/app/darwin_ios.m`, `viewDidLoad`) sets `glview.enableSetNeedsDisplay = NO` and installs an unconditional `CADisplayLink` targeting `render:`, so every display tick runs `render:` → `[glview display]` → `glkView:drawInRect:` → `drawloop()` (the `//export drawloop` in `internal/driver/mobile/app/darwin_ios.go`), synchronously on the main thread. (`startloop`/`loop()`, the background-goroutine renderer, is only started on macOS via `darwin_desktop.m`; on iOS `drawloop` is the sole renderer.)

`drawloop`'s select:
```go
select {
case <-workAvailable:    theApp.worker.DoWork()
case <-theApp.publish:   theApp.publishResult <- PublishResult{}; return
case <-time.After(100 * time.Millisecond): return   // darwin_ios.go:233
}
```

The Go-side driver is dirty-gated (`internal/driver/mobile/driver.go`): a 60 Hz ticker sends `paint.Event`, and `handlePaint` only calls `Publish()` when `canvasNeedRefresh`. So:

- On a **dirty** frame, `drawloop` returns immediately via the work/publish cases.
- On an **idle** frame (Fyne has nothing to draw — exactly the case while a *native* overlay scrolls over a static canvas), neither fires, so `drawloop` blocks the **main thread for the full 100 ms** fallback, then returns; the next `CADisplayLink` tick re-enters and parks another 100 ms, back-to-back. The native `UIScrollView`'s pan gesture handling and CoreAnimation commit, which run on that same main thread, are starved.

Fyne's own scroll widgets are smooth because Fyne is dirty every animating frame, so `drawloop` returns fast via the publish path.

### Evidence (on device)

iPhone 16 Pro Max, iOS, Fyne v2.7.4. An Instruments `runloop-events` capture during a ~19.5 s native-view scroll: the main run loop spent **18.6 s (95%) inside ~100 ms `individual_iteration` passes** — 77 iterations averaging **102.8 ms** — i.e. the 100 ms `drawloop` park, observed directly. Reducing the fallback to 2 ms eliminates the lag (verified on device).

### To reproduce

Float a native `UIScrollView` (e.g. a `UITextView`) as a subview of the window's root view controller and scroll it; compare against `container.Scroll`. Happy to share a minimal repro if useful.

### Proposed fix

Shorten the idle fallback so the main thread is freed between ticks. The fallback is a watchdog so `drawInRect` always returns even when no present is queued — it is **not** part of the present protocol — so a smaller value preserves that guarantee; it only needs to stay well under one refresh (16.6 ms @ 60 Hz / 8.3 ms @ 120 Hz). I have a PR ready changing it to a documented `idleRenderTimeout = 2 * time.Millisecond`. Worst case for a dirty publish that loses the race is a one-tick (~16 ms) deferral — strictly better than the current up-to-100 ms.

The deeper fix is **on-demand rendering** (only `[glview display]` on dirty frames, e.g. via the commented-out `enableSetNeedsDisplay = YES` path), so idle frames never enter `drawloop` at all — happy to discuss if you'd prefer that.

### Notes

Present unchanged on `develop` and `master`; long-standing inherited gomobile (`golang.org/x/mobile`) code. Searched existing issues — none describe this mechanism (closest are about Fyne's own scroll widgets / idle CPU, not the iOS main-thread `drawloop`).

- Fyne version: v2.7.4 (and `develop`)
- Platform: iOS (iPhone 16 Pro Max)
