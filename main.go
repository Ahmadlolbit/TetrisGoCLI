package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"awesomeProject/internal/input"
	"awesomeProject/internal/render"
)

const frameDelay = 16 * time.Millisecond

func main() {
	cols, rows := termSize()
	if cols < compositeW || rows < compositeH+2 {
		fmt.Printf("Terminal too small. Need at least %dx%d, got %dx%d.\n", compositeW, compositeH+2, cols, rows)
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
	done := make(chan struct{})
	go func() {
		<-sig
		close(done)
	}()

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	resize := make(chan [2]int, 1)
	go func() {
		for range winch {
			cols, rows := termSize()
			select {
			case <-resize:
			default:
			}
			resize <- [2]int{cols, rows}
		}
	}()

	newApp(scr, in).run(done, resize)
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
