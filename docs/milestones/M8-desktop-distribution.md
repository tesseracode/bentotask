# Milestone 8: Desktop App & Distribution

**Goal**: Package BentoTask as a native desktop app with cross-platform installers.

---

## Context

BentoTask already runs as a single Go binary (`bt serve`) that serves the web UI via `go:embed`. The desktop milestone wraps this in native packaging so users get a proper app experience — no terminal required.

## Architecture Decision: Wails

[Wails v2](https://wails.io) is the recommended approach:
- **What it is**: Go framework that creates native desktop apps using a webview (not Electron — no bundled Chromium)
- **How it works**: Our Go backend (`app.App`) runs directly in the Wails process. The SvelteKit UI renders in a native OS webview. No HTTP server needed — Wails bridges Go ↔ JS directly.
- **Why not Electron/Tauri**: Electron bundles Chromium (100MB+). Tauri requires Rust. Wails uses Go (which we already have) and the OS-native webview (WebKit on macOS, WebView2 on Windows, WebKitGTK on Linux).
- **Binary size**: ~20-25MB (vs ~150MB for Electron)
- **Alternative (M8.1)**: Before Wails, the simplest step is `bt serve --open` which auto-launches the default browser. Zero dependencies, works immediately.

## Tasks

### M8.1: `bt serve --open` — Browser Auto-Launch
- Add `--open` / `-o` flag to `bt serve`
- After server starts, call `exec.Command("open", url)` (macOS) / `xdg-open` (Linux) / `start` (Windows)
- Detect OS via `runtime.GOOS`
- This is the simplest "desktop-like" experience — no new dependencies

### M8.2: Wails Native Desktop App
- Initialize Wails project alongside existing code
- Create `cmd/bentotask-desktop/` entry point (separate from `cmd/bt/` CLI)
- Wails frontend points at the existing `web/` SvelteKit build
- Bind Go `app.App` methods to the Wails JS bridge
- The API client (`web/src/lib/api.ts`) may need a Wails-compatible fetch wrapper or use `wails.Call()` bindings
- Window: 1200×800 default, resizable, title "BentoTask"
- System tray icon (optional, nice-to-have)
- Dependencies: `github.com/wailsapp/wails/v2`

### M8.3: macOS Distribution
- Wails builds a `.app` bundle natively
- Create `.dmg` installer using `create-dmg` (npm package or shell script)
- Code signing (optional for personal use, required for distribution): Apple Developer ID
- Universal binary (amd64 + arm64) via `wails build -platform darwin/universal`
- Icon: bento box SVG → `.icns` conversion
- Info.plist: bundle ID `com.tesserabox.bentotask`

### M8.4: Windows Distribution
- Wails builds a `.exe` natively
- Create installer using Inno Setup or NSIS (scriptable, free)
- Include: start menu shortcut, optional desktop shortcut, uninstaller
- Cross-compile from macOS: `wails build -platform windows/amd64` (requires Docker or Windows cross-compile toolchain)
- Icon: bento box SVG → `.ico` conversion

### M8.5: Linux Distribution
- AppImage (most universal — single file, no install required):
  - Use `appimagetool` to package the Wails build
  - Include `.desktop` file and icon
- Optional: Flatpak manifest for Flathub distribution
- Optional: `.deb` and `.rpm` packages via `nfpm`
- Binary also works standalone: `./bentotask` with `--data-dir` flag

### M8.6: Cross-Compilation Makefile Targets
```makefile
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-windows-amd64

build-darwin-amd64:
    GOOS=darwin GOARCH=amd64 go build -o dist/bt-darwin-amd64 ./cmd/bt/

build-darwin-arm64:
    GOOS=darwin GOARCH=arm64 go build -o dist/bt-darwin-arm64 ./cmd/bt/

# etc.
```
- For CLI-only builds (no Wails), cross-compilation is trivial with `GOOS`/`GOARCH`
- For Wails builds, platform-specific builds require the target OS SDK (or CI runners)

### M8.7: CI Release Pipeline
- GitHub Actions workflow triggered on tag push (`v*`)
- Build matrix: darwin/amd64, darwin/arm64, linux/amd64, windows/amd64
- CLI binaries: direct `go build` cross-compilation
- Desktop binaries: Wails builds on platform-specific runners
- Create GitHub Release with all binaries + checksums
- Optional: Homebrew tap formula for macOS CLI install

## Implementation Order

1. **M8.1** first (5 minutes of work, huge UX win)
2. **M8.6** next (Makefile targets for CLI cross-compilation)
3. **M8.2** (Wails desktop — the big one)
4. **M8.3 + M8.4 + M8.5** (platform packaging, can be parallelized)
5. **M8.7** last (CI pipeline ties it all together)
