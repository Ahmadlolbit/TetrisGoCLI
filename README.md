# Chaos Blocks

A fast, beautiful terminal Tetris written in Go — truecolor blocks, particle
effects, combo scoring, and a controlled "chaos" system that bends the rules in
surprising but fair ways.

```
                 C H A O S   B L O C K S
        ┌────────────────────┐ ┌──────────┐
 HOLD   │ · · · · · · · · · ·│ │ NEXT     │
 ┌────┐ │ · · · · ▀▀ · · · · │ │  ▀▀▀▀    │
 │    │ │ · · · ▀▀▀▀▀▀ · · · │ │          │
 └────┘ │ · · · · · · · · · ·│ │ COMBO x3 │
 SCORE  │ ▒▒ · · · · · · · · │ │ ▮▮▮▮▯▯▯▯  │
 12 400 │ ▀▀▀▀▀▀ · · ▀▀▀▀▀▀ │ │ CHAOS    │
        └────────────────────┘ │ ▮▮▮▮▮▯▯▯  │
                               └──────────┘
```

## Run

You need a Go toolchain (1.17+) and a real ANSI terminal.

```sh
go run .
```

Or build a binary:

```sh
go build -o chaosblocks . && ./chaosblocks
```

The terminal must be at least **50×24**. The game detects undersized terminals
and prompts you to resize; it always restores the terminal cleanly on exit,
Ctrl-C, or panic.

## Controls

| Action | Default keys |
|--------|--------------|
| Move left / right | `←` `→` · `h` `l` |
| Soft drop | `↓` · `j` |
| Hard drop | `Space` |
| Rotate CW / CCW | `↑` `x` `k` / `z` |
| Hold | `c` |
| Pause | `Esc` · `p` |
| Restart | `r` |
| Cycle theme | `t` |
| Toggle chaos | `m` |
| Quit | `q` · `Ctrl-C` |

In menus: `↑`/`↓` navigate, `←`/`→` adjust values, `Enter`/`Space` select,
`Esc` goes back.

All gameplay keys are **rebindable** in *Settings → Key Bindings*. The arrow
keys, `Enter`, `Esc`, and `Ctrl-C` stay fixed so you can never lock yourself out.

## Modes

- **Marathon** — endless climb through the levels.
- **Sprint** — clear 40 lines as fast as you can; ranked by time.
- **Chaos** — the signature mode, chaos events in overdrive.
- **Classic** — chaos off, gentler gravity, for purists.

## Features

- Super Rotation System with wall kicks and T-spin detection, fed by a 7-bag
  randomizer; hold slot, next queue, and a ghost piece.
- Combo and back-to-back scoring with an escalating on-screen meter.
- Visual effects: line-clear flashes and particle bursts, hard-drop trails,
  screen shake, and a level-up wash — capped so they never hurt readability.
- **Chaos events**, always telegraphed and bounded: Garbage Surge, Gravity
  Spike, Color Scramble, Lights Dim, and the rewarding Bonus Frenzy.
- Three themes — **Neon**, **Synthwave**, **Mono-glow**.
- **Color modes**: truecolor with graceful degradation to 256-color and
  16-color, auto-detected from your terminal (overridable in Settings).
- Persistent per-mode high scores and settings.

## Settings & persistence

Settings (theme, starting level, color mode, key bindings) and per-mode high
scores are saved to your user config directory, e.g.
`~/.config/chaosblocks/state.json`. The file is human-readable JSON and
corruption-tolerant — a missing or malformed file falls back to sane defaults
rather than crashing. Set `CHAOSBLOCKS_CONFIG_DIR` to override the location.

## Development

```sh
go test ./...      # run the test suite
go vet ./...       # static checks
gofmt -l .         # formatting (should print nothing)
```

The codebase is intentionally **comment-free** — it documents itself through
naming and structure. Core game logic in `internal/game` is deterministic and
unit-tested independently of rendering.

### Layout

```
main.go                 entry point, loop, signal/resize handling
app.go, menu.go         app state machine, screens, settings
view.go, mode.go        playfield rendering, themes, game modes
scores.go               high-score board
internal/game           board, pieces, SRS, scoring (headless, tested)
internal/render         framebuffer, cell diffing, color quantization
internal/input          raw keyboard, configurable keymap
internal/effects        particles, flashes, screen shake
internal/chaos          chaos meter and events
internal/store          JSON persistence
```
