package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

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
	g := game.NewGame(time.Now().UnixNano())
	th := neon
	paused := false

	ticker := time.NewTicker(frameDelay)
	defer ticker.Stop()
	dt := frameDelay.Seconds()

	for {
		select {
		case ev := <-in.Events():
			switch ev {
			case input.Quit:
				return
			case input.Restart:
				g = game.NewGame(time.Now().UnixNano())
				paused = false
			case input.Pause:
				if !g.Over {
					paused = !paused
				}
			default:
				if !paused && !g.Over {
					apply(g, ev)
				}
			}
		case <-ticker.C:
			if !paused && !g.Over {
				g.Tick(dt)
			}
			draw(scr.Back(), g, th)
			if paused && !g.Over {
				drawBanner(scr.Back(), (scr.W-compositeW)/2+boardOffset, (scr.H-compositeH)/2, "PAUSED", "press P to resume", th)
			}
			scr.Flush()
		}
	}
}

func apply(g *game.Game, ev input.Event) {
	switch ev {
	case input.MoveLeft:
		g.TryMove(-1, 0)
	case input.MoveRight:
		g.TryMove(1, 0)
	case input.SoftDrop:
		g.SoftDrop()
	case input.HardDrop:
		g.HardDrop()
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
