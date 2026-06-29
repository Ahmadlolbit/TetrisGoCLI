package input

import "testing"

func TestTokenize(t *testing.T) {
	cases := []struct {
		b    byte
		want Key
	}{
		{'A', "a"},
		{'z', "z"},
		{' ', "space"},
		{'\r', "enter"},
		{'\n', "enter"},
		{3, "c-c"},
		{'5', "5"},
		{9, ""},
		{8, ""},
		{127, ""},
	}
	for _, tc := range cases {
		if got := tokenize(tc.b); got != tc.want {
			t.Errorf("tokenize(%d) = %q, want %q", tc.b, got, tc.want)
		}
	}
}

func TestReserved(t *testing.T) {
	for _, k := range []Key{"up", "down", "left", "right", "enter", "esc", "c-c"} {
		if !Reserved(k) {
			t.Errorf("%q should be reserved", k)
		}
	}
	for _, k := range []Key{"h", "x", "space", "q", "a"} {
		if Reserved(k) {
			t.Errorf("%q should not be reserved", k)
		}
	}
}

func TestCompileFixedWins(t *testing.T) {
	m := compile(DefaultKeymap())
	want := map[Key]Event{
		"h": MoveLeft, "l": MoveRight, "j": SoftDrop, "space": HardDrop,
		"x": RotateCW, "z": RotateCCW, "c": Hold, "q": Quit,
		"up": RotateCW, "k": RotateCW, "down": SoftDrop, "left": MoveLeft, "right": MoveRight,
		"enter": Select, "esc": Pause,
	}
	for k, e := range want {
		if m[k] != e {
			t.Errorf("compiled[%q] = %v, want %v", k, m[k], e)
		}
	}
}

func TestImportExportRoundTrip(t *testing.T) {
	m := DefaultKeymap()
	m[MoveLeft] = "a"
	m[HardDrop] = "f"
	got := ImportKeymap(ExportKeymap(m))
	for e, k := range m {
		if got[e] != k {
			t.Errorf("roundtrip[%v] = %q, want %q", e, got[e], k)
		}
	}
}

func TestImportTolerant(t *testing.T) {
	got := ImportKeymap(nil)
	if got[MoveLeft] != "h" {
		t.Errorf("nil import should yield defaults, MoveLeft = %q", got[MoveLeft])
	}
	got = ImportKeymap(map[string]string{
		"moveLeft": "a",
		"bogus":    "b",
		"hold":     "up",
		"quit":     "",
	})
	if got[MoveLeft] != "a" {
		t.Errorf("valid override dropped, MoveLeft = %q", got[MoveLeft])
	}
	if got[Hold] != "c" {
		t.Errorf("reserved token should be ignored, Hold = %q (want default c)", got[Hold])
	}
	if got[Quit] != "q" {
		t.Errorf("empty token should be ignored, Quit = %q (want default q)", got[Quit])
	}
}

func TestImportKeymapDeterministicUnique(t *testing.T) {
	in := map[string]string{"moveLeft": "Z", "rotateCCW": "z", "hold": "X", "rotateCW": "x"}
	first := ImportKeymap(in)
	if first[MoveLeft] != "z" {
		t.Errorf("uppercase token not lowercased: MoveLeft = %q", first[MoveLeft])
	}
	for i := 0; i < 20; i++ {
		got := ImportKeymap(in)
		for _, e := range Bindable {
			if got[e] != first[e] {
				t.Fatalf("nondeterministic import for %v: %q vs %q", e, got[e], first[e])
			}
		}
	}
	seen := map[Key]Event{}
	for _, e := range Bindable {
		k := first[e]
		if k == "" {
			continue
		}
		if prev, dup := seen[k]; dup {
			t.Errorf("duplicate key %q bound to %v and %v", k, prev, e)
		}
		seen[k] = e
	}
}

func TestCaptureIgnoresControlBytes(t *testing.T) {
	r := &Reader{events: make(chan Event, 16), captures: make(chan Key, 16)}
	r.SetKeymap(DefaultKeymap())
	r.BeginCapture()
	r.parse([]byte{9})
	r.parse([]byte{127})
	select {
	case k := <-r.Captures():
		t.Errorf("control byte should not be captured, got %q", k)
	default:
	}
	r.parse([]byte("g"))
	if got := <-r.Captures(); got != "g" {
		t.Errorf("captured %q, want g", got)
	}
	r.EndCapture()
}

func TestParseSplitEscapeSequence(t *testing.T) {
	r := &Reader{events: make(chan Event, 16), captures: make(chan Key, 16)}
	r.SetKeymap(DefaultKeymap())
	r.pending = append([]byte(nil), r.parse([]byte{0x1b, '['})...)
	select {
	case e := <-r.Events():
		t.Fatalf("incomplete escape should not emit, got %v", e)
	default:
	}
	if len(r.pending) != 2 {
		t.Fatalf("incomplete escape should be carried over, pending=%v", r.pending)
	}
	data := append(r.pending, 'C')
	r.pending = append([]byte(nil), r.parse(data)...)
	if got := <-r.Events(); got != MoveRight {
		t.Fatalf("split arrow should resolve to MoveRight, got %v", got)
	}
}

func TestParseModifiedArrowConsumed(t *testing.T) {
	r := &Reader{events: make(chan Event, 16), captures: make(chan Key, 16)}
	r.SetKeymap(DefaultKeymap())
	if leftover := r.parse([]byte{0x1b, '[', '1', ';', '5', 'A'}); leftover != nil {
		t.Fatalf("complete CSI should leave no leftover, got %v", leftover)
	}
	select {
	case e := <-r.Events():
		t.Fatalf("modified arrow should emit nothing, got %v", e)
	default:
	}
}

func TestImportKeymapNeverLeavesActionUnbound(t *testing.T) {
	m := ImportKeymap(map[string]string{"hold": "x"})
	seen := map[Key]bool{}
	for _, e := range Bindable {
		if m[e] == "" {
			t.Fatalf("action %v left unbound", e)
		}
		if seen[m[e]] {
			t.Fatalf("duplicate key %q across actions", m[e])
		}
		seen[m[e]] = true
	}
}

func TestReaderDispatchAndCapture(t *testing.T) {
	r := &Reader{events: make(chan Event, 16), captures: make(chan Key, 16)}
	r.SetKeymap(DefaultKeymap())

	r.parse([]byte("h"))
	if got := <-r.Events(); got != MoveLeft {
		t.Errorf("h => %v, want MoveLeft", got)
	}
	r.parse([]byte{0x1b, '[', 'D'})
	if got := <-r.Events(); got != MoveLeft {
		t.Errorf("left arrow => %v, want MoveLeft", got)
	}

	r.BeginCapture()
	r.parse([]byte("a"))
	if got := <-r.Captures(); got != "a" {
		t.Errorf("captured %q, want a", got)
	}
	select {
	case e := <-r.Events():
		t.Errorf("no event expected during capture, got %v", e)
	default:
	}

	r.parse([]byte{3})
	select {
	case e := <-r.Events():
		if e != Quit {
			t.Errorf("ctrl-c => %v, want Quit", e)
		}
	default:
		t.Error("ctrl-c during capture should still emit Quit")
	}
	r.EndCapture()
}
