# Measuring visual bugs in a headless browser

When a visual defect (animation, rendering, timing) survives its first fix, stop fixing by reasoning and measure it (rule in [`spec/constitution.md`](../spec/constitution.md), workflow step 5). This page is how, distilled from the probe that found milestone 04's edit transition flicker: four reasoning-only fixes had each uncovered the next problem, one measured pass found the cause.

## Setup (once per machine)

A standalone headless Chrome from Google's [Chrome for Testing](https://googlechromelabs.github.io/chrome-for-testing/) builds, not a snap (snaps are unreliable in WSL).

```bash
# Latest stable URL: the "chrome-headless-shell" / "linux64" entry in
# https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json
V=154.0.8037.57
D=$HOME/.local/share/chrome-headless-shell/$V
curl -fsSL "https://storage.googleapis.com/chrome-for-testing-public/$V/linux64/chrome-headless-shell-linux64.zip" -o /tmp/chs.zip
mkdir -p "$D" && unzip -qo /tmp/chs.zip -d "$D"
ln -sf "$D/chrome-headless-shell-linux64/chrome-headless-shell" ~/.local/bin/chrome-headless-shell
sudo apt install -y libnss3 libasound2t64 fonts-liberation   # libraries it needs; real fonts for realistic text
```

## Quick check: one screenshot

No protocol needed for a still picture at phone size and density:

```bash
chrome-headless-shell --no-sandbox --hide-scrollbars --window-size=390,844 \
  --force-device-scale-factor=3 --screenshot=/tmp/shot.png http://127.0.0.1:8080/trips/1
```

To keep a running `make dev` and its database untouched, run a separate copy: `go build -o /tmp/rt ./cmd/web && DB_PATH=/tmp/check.db ADDR=127.0.0.1:18080 /tmp/rt`.

## Run

```bash
make run &                                   # the app on 127.0.0.1:8080
chrome-headless-shell --remote-debugging-port=9222 --no-sandbox about:blank &
curl -s http://127.0.0.1:9222/json/list      # "webSocketDebuggerUrl" of the page target
```

A small Go program (in the scratchpad, not the repo: it needs a websocket dependency the app does not) talks to that URL with the [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/): JSON messages `{"id", "method", "params"}` over one websocket, replies matched by `id`. [`github.com/coder/websocket`](https://github.com/coder/websocket) was enough; no CDP library needed.

## The technique

1. **Emulate the phone, not the desktop.** `Emulation.setDeviceMetricsOverride` with `{"width": 390, "height": 844, "deviceScaleFactor": 3, "mobile": true}`. Pixel density matters: in 04 the bug was invisible at density 1 and large at 3.
2. **Load the page:** `Page.navigate`, then wait about 1.5s.
3. **Freeze the transition at a chosen moment** by evaluating this (`Runtime.evaluate` with `awaitPromise: true`, `returnByValue: true`), with `T` the time in ms and `SEL` the element to click:

   ```js
   new Promise(res => {
     const orig = document.startViewTransition.bind(document);
     document.startViewTransition = cb => {
       const vt = orig(cb);
       vt.ready.then(() => requestAnimationFrame(() => {
         for (const a of document.getAnimations()) { a.pause(); a.currentTime = T; }
         requestAnimationFrame(() => requestAnimationFrame(() => res("paused")));
       }));
       return vt;
     };
     document.querySelector(SEL).click();
   })
   ```

   Wrapping `startViewTransition` catches the transition htmx starts (`hx-swap="... transition:true"`). For a plain CSS animation, skip the wrapper and pause `document.getAnimations()` directly.
4. **Screenshot mid-transition:** `Page.captureScreenshot` (`format: png`, base64 in `data`).
5. **Finish and screenshot the end state:** evaluate `for (const a of document.getAnimations()) a.finish()`, wait about 400ms, screenshot again.
6. **Diff the region you care about.** Get its box in CSS pixels with `getBoundingClientRect()` before clicking, multiply by the device scale factor, and compare the two images pixel by pixel: report the maximum channel difference (0 to 255) and how many pixels differ by more than 8. Repeat for T = 0, 20, 40, 60, 80, 100ms.

Reading the numbers: unchanged content should be 0 (or 1, rounding) at every T. In 04 the header measured 233 at 0ms falling to 4 at 100ms: the old view-transition snapshot was rasterised blurrier than the live page, so text that did not change still visibly sharpened during a 120ms fade.

Save the mid-transition screenshots too and look at them (the Read tool shows images): a number says something differs, a picture says what.

## Security

- **Localhost only.** Point this browser at the app on 127.0.0.1 and nothing else. It is a manually downloaded copy that does not update itself, and `--no-sandbox` switches off Chrome's main defence against malicious pages (often needed in WSL); both are acceptable for our own pages only.
- **Stop it when done.** `--remote-debugging-port` lets anything that can reach the port fully control the browser. It listens on this machine only, but with WSL mirrored networking Windows programs can reach it too. Never add a firewall rule for 9222.
- **Refresh it before reuse** if it has been a while: repeat the setup with the current version from the JSON listing.
- The system libraries come from Ubuntu's signed repositories and are patched by the normal `apt upgrade`.

## Limits

- Headless Chrome is Chromium only. Safari/WebKit behaviour is not covered; the phone check is still the final word.
- Emulated pixel density is close to a real phone, not identical. Treat a measured cause as strong evidence and confirm the fix on the phone.
