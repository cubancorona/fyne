> **STATUS: SUPERSEDED — DO NOT FILE.** The companion issue was filed as
> fyne-io/fyne#6368 and closed (overlay framing rejected). Also, this PR text
> describes only the 2ms timeout; the shipped fix later grew a framePainting
> mid-paint present-guard (without it the short timeout can present half-drawn
> frames). Both halves now live in the cubancorona/fyne fork (bt-main). The
> open upstream path is an on-demand-rendering PR framed as draw-loop
> efficiency. See docs/FYNE_FORK_PLAN.md (bibletext repo), decision point 3.

Title: Fix iOS native-overlay scroll lag: shorten drawloop idle main-thread park (100ms → 2ms)

Fixes #<ISSUE_NUMBER>

### What

On iOS, `drawloop()` runs on the main thread every `CADisplayLink` tick (`render:` → `[glview display]` → `drawInRect` → `drawloop`; `startloop`/`loop()` is macOS-only). On an idle frame its only exit is the `time.After` fallback, so it parks the main run loop for the full 100 ms, back-to-back, starving anything else on that thread. This makes a native `UIScrollView` floated over the Fyne canvas scroll laggily (Fyne's own widgets are fine because they're dirty every frame and return via the publish path). Details and on-device trace in #<ISSUE_NUMBER>.

This replaces the 100 ms fallback with a documented `idleRenderTimeout = 2 * time.Millisecond`, freeing ~14.6 ms of each 60 Hz tick (~6.3 ms of each 120 Hz tick) back to the main thread. The fallback is a watchdog (ensures `drawInRect` returns even when no present is queued), not part of the present protocol, so the smaller value preserves the guarantee.

### Evidence

On-device Instruments `runloop-events` trace (iPhone 16 Pro Max, iOS): during a 19.5 s native scroll the main run loop spent 18.6 s (95%) in ~100 ms iterations (77 × avg 102.8 ms). With 2 ms the lag is gone, with no regression to Fyne's own rendering.

### Correctness / consequences

- **No deadlock, no dropped frame.** `workAvailable` is unbounded; `publish`/`publishResult` are a rendezvous. If the timer fires mid-publish, the pending work/publish is consumed on the next `drawloop` entry — worst case a single idle-frame publish presents one tick (~16 ms @ 60 Hz, ~8 ms @ 120 Hz) later, *better* than the up-to-100 ms recovery window today.
- **Dirty frames unaffected** — they return via the work/publish cases regardless of the timeout (`Publish()` blocks until serviced; the timer never truncates an in-progress paint, since `DoWork()` drains all queued work before the select is re-entered).
- **Honest caveat (idle CPU):** the old 100 ms park *suppressed* `CADisplayLink` ticks (the link can't re-deliver until `drawInRect` returns), so a fully-idle, foregrounded screen re-entered `drawloop` ~10×/s; at 2 ms it rises toward the ~60 Hz ceiling — ~6× more `select` iterations and `time.After` allocations per second. The absolute cost is microseconds/s, dominated by the unconditional 60 Hz `CADisplayLink` that runs regardless, and there is no busy-spin (the link still caps re-entry at the refresh rate). I did not measure idle power; happy to provide a profile, or to gate the value differently, if you'd like.

### Scope

iOS-only and a single logical line. The edited `time.After` is the only one in `internal/driver/mobile`, inside `drawloop`, which is defined only in `darwin_ios.go` (`//go:build darwin && ios`). macOS (`loop()` in `darwin_desktop.go`) and Android (`android.go`) use a background-goroutine renderer with no timeout and are untouched; build constraints exclude `darwin_ios.go` from every non-iOS target.

### Testing

This is cgo / main-thread Objective-C render-loop code behind iOS build tags, not exercised by `go test ./...`, so there is no meaningful unit test — the on-device runloop trace is the evidence. Verified on a physical iPhone 16 Pro Max that native-overlay scrolling goes from laggy to smooth with no regression to Fyne's own rendering.

### Alternative considered

The deeper fix is on-demand rendering (only `[glview display]` on dirty frames via the commented-out `enableSetNeedsDisplay = YES` path) so idle frames never enter `drawloop`. This PR is the minimal, low-risk change; happy to pursue the larger one if preferred.

---
Happy to sign the CLA.
