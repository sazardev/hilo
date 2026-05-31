# hilo — AGENTS.md

## Build & Run
- `go run .` — start the TUI
- `go build .` — produces binary `./hilo`
- No tests, no CI, no lint config. Use `go vet .` for static checks.

## Architecture
- Single flat `main` package, 5 source files:
  - `main.go` — entrypoint, Bubble Tea program with alt screen + mouse
  - `model.go` — model, tabs, all update/view logic
  - `styles.go` — lipgloss style factory parameterized by theme
  - `themes.go` — 9 full themes + 20 color accent schemes
  - `config.go` — JSON config read/write to `~/.config/hilo/config.json`
- Config saves theme index, color index, and mode (dark/light). No other persistent state.

## Framework quirks
- Bubble Tea TUI with `tea.WithAltScreen()` — must run in interactive terminal, exits with `q`/`ctrl+c`/`esc`
- All config I/O errors are silently swallowed by design (TUI resilience)

## Conventions
- No os.Exit outside `main.go`
- Styles rebuilt from scratch on any theme/color/mode change via `rebuildStyles()`
