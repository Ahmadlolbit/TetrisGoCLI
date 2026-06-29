package input

import (
	"os"
	"os/exec"
	"strings"
	"sync"
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
	Select
	Quit
)

type Key string

type Keymap map[Event]Key

var Bindable = []Event{
	MoveLeft, MoveRight, SoftDrop, HardDrop,
	RotateCW, RotateCCW, Hold,
	Pause, Restart, ThemeNext, ToggleChaos, Quit,
}

func DefaultKeymap() Keymap {
	return Keymap{
		MoveLeft:    "h",
		MoveRight:   "l",
		SoftDrop:    "j",
		HardDrop:    "space",
		RotateCW:    "x",
		RotateCCW:   "z",
		Hold:        "c",
		Pause:       "p",
		Restart:     "r",
		ThemeNext:   "t",
		ToggleChaos: "m",
		Quit:        "q",
	}
}

var actionIDs = map[Event]string{
	MoveLeft:    "moveLeft",
	MoveRight:   "moveRight",
	SoftDrop:    "softDrop",
	HardDrop:    "hardDrop",
	RotateCW:    "rotateCW",
	RotateCCW:   "rotateCCW",
	Hold:        "hold",
	Pause:       "pause",
	Restart:     "restart",
	ThemeNext:   "theme",
	ToggleChaos: "chaos",
	Quit:        "quit",
}

var actionLabels = map[Event]string{
	MoveLeft:    "Move Left",
	MoveRight:   "Move Right",
	SoftDrop:    "Soft Drop",
	HardDrop:    "Hard Drop",
	RotateCW:    "Rotate CW",
	RotateCCW:   "Rotate CCW",
	Hold:        "Hold",
	Pause:       "Pause",
	Restart:     "Restart",
	ThemeNext:   "Cycle Theme",
	ToggleChaos: "Toggle Chaos",
	Quit:        "Quit",
}

func ActionID(e Event) string { return actionIDs[e] }

func Label(e Event) string { return actionLabels[e] }

var fixed = map[Key]Event{
	"up":    RotateCW,
	"k":     RotateCW,
	"down":  SoftDrop,
	"left":  MoveLeft,
	"right": MoveRight,
	"enter": Select,
	"esc":   Pause,
}

func Reserved(k Key) bool {
	if k == "c-c" {
		return true
	}
	_, ok := fixed[k]
	return ok
}

func KeyLabel(k Key) string {
	switch k {
	case "":
		return "—"
	case "space":
		return "Space"
	case "enter":
		return "Enter"
	case "esc":
		return "Esc"
	case "up":
		return "Up"
	case "down":
		return "Down"
	case "left":
		return "Left"
	case "right":
		return "Right"
	}
	return strings.ToUpper(string(k))
}

func compile(m Keymap) map[Key]Event {
	out := make(map[Key]Event, len(m)+len(fixed))
	for e, k := range m {
		if k == "" || Reserved(k) {
			continue
		}
		out[k] = e
	}
	for k, e := range fixed {
		out[k] = e
	}
	return out
}

func ExportKeymap(m Keymap) map[string]string {
	out := make(map[string]string, len(m))
	for e, k := range m {
		if id, ok := actionIDs[e]; ok {
			out[id] = string(k)
		}
	}
	return out
}

func ImportKeymap(p map[string]string) Keymap {
	def := DefaultKeymap()
	m := make(Keymap, len(def))
	used := make(map[Key]bool)
	assign := func(e Event, k Key) {
		if k == "" || Reserved(k) || used[k] {
			return
		}
		m[e] = k
		used[k] = true
	}
	for _, e := range Bindable {
		if ks, ok := p[ActionID(e)]; ok {
			assign(e, Key(strings.ToLower(ks)))
		}
	}
	for _, e := range Bindable {
		if _, ok := m[e]; !ok {
			assign(e, def[e])
		}
	}
	for _, e := range Bindable {
		if _, ok := m[e]; ok {
			continue
		}
		for _, alt := range Bindable {
			if !used[def[alt]] {
				assign(e, def[alt])
				break
			}
		}
	}
	return m
}

type Reader struct {
	events    chan Event
	captures  chan Key
	mu        sync.RWMutex
	lookup    map[Key]Event
	capturing bool
	pending   []byte
}

func (r *Reader) Events() <-chan Event {
	return r.events
}

func (r *Reader) Captures() <-chan Key {
	return r.captures
}

func (r *Reader) SetKeymap(m Keymap) {
	c := compile(m)
	r.mu.Lock()
	r.lookup = c
	r.mu.Unlock()
}

func (r *Reader) BeginCapture() {
	r.mu.Lock()
	r.capturing = true
	r.mu.Unlock()
}

func (r *Reader) EndCapture() {
	r.mu.Lock()
	r.capturing = false
	r.mu.Unlock()
}

func Start() (*Reader, func()) {
	saved := captureState()
	rawMode()
	r := &Reader{
		events:   make(chan Event, 64),
		captures: make(chan Key, 8),
	}
	r.SetKeymap(DefaultKeymap())
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
		data := buf[:n]
		if len(r.pending) > 0 {
			data = append(r.pending, data...)
		}
		r.pending = append([]byte(nil), r.parse(data)...)
	}
}

func (r *Reader) parse(b []byte) []byte {
	i := 0
	for i < len(b) {
		c := b[i]
		if c == 0x1b {
			if i+1 < len(b) && b[i+1] == '[' {
				j := i + 2
				for j < len(b) && (b[j] < 0x40 || b[j] > 0x7e) {
					j++
				}
				if j >= len(b) {
					return b[i:]
				}
				if j == i+2 {
					switch b[j] {
					case 'A':
						r.dispatch("up")
					case 'B':
						r.dispatch("down")
					case 'C':
						r.dispatch("right")
					case 'D':
						r.dispatch("left")
					}
				}
				i = j + 1
				continue
			}
			r.dispatch("esc")
			i++
			continue
		}
		r.dispatch(tokenize(c))
		i++
	}
	return nil
}

func tokenize(c byte) Key {
	switch c {
	case 3:
		return "c-c"
	case '\r', '\n':
		return "enter"
	case ' ':
		return "space"
	}
	if c < 0x20 || c == 0x7f {
		return ""
	}
	if c >= 'A' && c <= 'Z' {
		return Key(string(rune(c + 32)))
	}
	return Key(string(rune(c)))
}

func (r *Reader) dispatch(k Key) {
	if k == "c-c" {
		r.emit(Quit)
		return
	}
	if k == "" {
		return
	}
	r.mu.RLock()
	capturing := r.capturing
	e, ok := r.lookup[k]
	r.mu.RUnlock()
	if capturing {
		select {
		case r.captures <- k:
		default:
		}
		return
	}
	if ok {
		r.emit(e)
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
