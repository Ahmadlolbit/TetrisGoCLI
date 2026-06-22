package input

import (
	"os"
	"os/exec"
	"strings"
)

type Event int

const (
	MoveLeft Event = iota
	MoveRight
	SoftDrop
	HardDrop
	RotateCW
	RotateCCW
	Hold
	Pause
	Restart
	ThemeNext
	ToggleChaos
	Quit
)

type Reader struct {
	events chan Event
}

func (r *Reader) Events() <-chan Event {
	return r.events
}

func Start() (*Reader, func()) {
	saved := captureState()
	rawMode()
	r := &Reader{events: make(chan Event, 64)}
	go r.loop()
	restore := func() {
		if saved != "" {
			run("stty", saved)
		} else {
			run("stty", "sane")
		}
	}
	return r, restore
}

func (r *Reader) loop() {
	buf := make([]byte, 16)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		r.parse(buf[:n])
	}
}

func (r *Reader) parse(b []byte) {
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case 3, 'q':
			r.emit(Quit)
		case ' ':
			r.emit(HardDrop)
		case 'z', 'Z':
			r.emit(RotateCCW)
		case 'x', 'X':
			r.emit(RotateCW)
		case 'c', 'C':
			r.emit(Hold)
		case 'h':
			r.emit(MoveLeft)
		case 'l':
			r.emit(MoveRight)
		case 'j':
			r.emit(SoftDrop)
		case 'k':
			r.emit(RotateCW)
		case 'p', 'P':
			r.emit(Pause)
		case 'r', 'R':
			r.emit(Restart)
		case 't', 'T':
			r.emit(ThemeNext)
		case 'm', 'M':
			r.emit(ToggleChaos)
		case 0x1b:
			if i+2 < len(b) && b[i+1] == '[' {
				switch b[i+2] {
				case 'A':
					r.emit(RotateCW)
				case 'B':
					r.emit(SoftDrop)
				case 'C':
					r.emit(MoveRight)
				case 'D':
					r.emit(MoveLeft)
				}
				i += 2
			} else {
				r.emit(Pause)
			}
		}
	}
}

func (r *Reader) emit(e Event) {
	select {
	case r.events <- e:
	default:
	}
}

func captureState() string {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func rawMode() {
	run("stty", "-echo", "-icanon", "min", "1", "time", "0")
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Run()
}
