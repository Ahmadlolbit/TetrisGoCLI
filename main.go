package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"awesomeProject/internal/chaos"
	"awesomeProject/internal/effects"
	"awesomeProject/internal/game"
	"awesomeProject/internal/input"
	"awesomeProject/internal/render"
)

const frameDelay = 16 * time.Millisecond

func main() {
	cols, rows := termSize()
	if cols < compositeW || rows < compositeH {
		fmt.Printf("Terminal too small. Need at least %dx%d, got %dx%d.\n", compositeW, compositeH+1, cols, rows)
		os.Exit(1)
	}

	in, restore := input.Start()
	scr := render.NewScreen(cols, rows)
	scr.Enter()

	cleanup := func() {
		scr.Leave()
		restore()
	}
	defer func() {
		if r := recover(); r != nil {
			cleanup()
			panic(r)
		}
		cleanup()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cleanup()
		os.Exit(0)
	}()

	loop(scr, in)
}

func loop(scr *render.Screen, in *input.Reader) {
	seed := time.Now().UnixNano()
	g := game.NewGame(seed)
	eng := effects.New(seed)
	ch := chaos.New(seed, true)
	themeIdx := 0
	paused := false

	ticker := time.NewTicker(frameDelay)
	defer ticker.Stop()
	dt := frameDelay.Seconds()

	for {
		th := themes[themeIdx]
		select {
		case ev := <-in.Events():
			switch ev {
			case input.Quit:
				return
			case input.ThemeNext:
				themeIdx = (themeIdx + 1) % len(themes)
			case input.ToggleChaos:
				ch.Toggle()
			case input.Restart:
				seed = time.Now().UnixNano()
				g = game.NewGame(seed)
				eng.Clear()
				ch.Reset()
				paused = false
			case input.Pause:
				if !g.Over {
					paused = !paused
				}
			default:
				if !paused && !g.Over {
					apply(g, eng, ch, ev, scr.W, scr.H, th)
				}
			}
		case <-ticker.C:
			ox, oy := origin(scr.W, scr.H)
			if !paused && !g.Over {
				if fired, ok := ch.Update(dt, g); ok {
					spawnChaos(eng, fired, ox, oy, th)
				}
				g.GravityScale = ch.GravityScale()
				g.ScoreScale = ch.ScoreScale()
				before := g.Level
				for _, res := range g.Tick(dt) {
					spawnLockEffects(eng, res, ox, oy, th)
					ch.OnPiece()
				}
				if g.Level > before {
					spawnLevelUp(eng, ox, oy, th)
				}
			}
			eng.Update(dt)
			sx, sy := eng.ShakeOffset()
			draw(scr.Back(), g, ch, th, sx, sy)
			eng.Apply(scr.Back(), sx, sy)
			if paused && !g.Over {
				drawBanner(scr.Back(), (scr.W-compositeW)/2+boardOffset, (scr.H-compositeH)/2, "PAUSED", "press P to resume", th)
			}
			scr.Flush()
		}
	}
}

func apply(g *game.Game, eng *effects.Engine, ch *chaos.Engine, ev input.Event, w, h int, th theme) {
	ox, oy := origin(w, h)
	switch ev {
	case input.MoveLeft:
		g.TryMove(-1, 0)
	case input.MoveRight:
		g.TryMove(1, 0)
	case input.SoftDrop:
		g.SoftDrop()
	case input.HardDrop:
		landing := g.Ghost()
		before := g.Level
		res := g.HardDrop()
		spawnHardDrop(eng, landing, ox, oy, th)
		spawnLockEffects(eng, res, ox, oy, th)
		ch.OnPiece()
		if g.Level > before {
			spawnLevelUp(eng, ox, oy, th)
		}
	case input.RotateCW:
		g.Rotate(game.CW)
	case input.RotateCCW:
		g.Rotate(game.CCW)
	case input.Hold:
		g.Hold()
	}
}

func termSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80, 24
	}
	var rows, cols int
	if _, err := fmt.Sscanf(string(out), "%d %d", &rows, &cols); err != nil || cols == 0 {
		return 80, 24
	}
	return cols, rows
}
