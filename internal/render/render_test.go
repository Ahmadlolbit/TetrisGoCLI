package render

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRgbTo256Cube(t *testing.T) {
	cases := []struct {
		c    Color
		want int
	}{
		{Color{0, 0, 0}, 16},
		{Color{255, 255, 255}, 231},
		{Color{255, 0, 0}, 196},
		{Color{0, 255, 0}, 46},
		{Color{0, 0, 255}, 21},
	}
	for _, tc := range cases {
		if got := rgbTo256(tc.c); got != tc.want {
			t.Errorf("rgbTo256(%v) = %d, want %d", tc.c, got, tc.want)
		}
	}
}

func TestRgbTo256Grayscale(t *testing.T) {
	cases := []struct {
		c    Color
		want int
	}{
		{Color{128, 128, 128}, 244},
		{Color{18, 18, 18}, 233},
		{Color{17, 17, 17}, 233},
	}
	for _, tc := range cases {
		if got := rgbTo256(tc.c); got != tc.want {
			t.Errorf("rgbTo256(%v) = %d, want %d", tc.c, got, tc.want)
		}
	}
}

func TestRgbTo256NearestGray(t *testing.T) {
	step := func(i int) int { return 8 + i*10 }
	for v := 0; v < 256; v++ {
		got := rgbTo256(Color{uint8(v), uint8(v), uint8(v)})
		if got < 232 {
			continue
		}
		ramp := step(got - 232)
		for i := 0; i < 24; i++ {
			cand := step(i)
			if sq(v-cand) < sq(v-ramp) {
				t.Errorf("gray %d picked ramp %d (val %d), but %d (val %d) is nearer", v, got, ramp, 232+i, cand)
				break
			}
		}
	}
}

func TestRgbTo16(t *testing.T) {
	cases := []struct {
		c    Color
		want int
	}{
		{Color{0, 0, 0}, 0},
		{Color{170, 0, 0}, 1},
		{Color{0, 170, 0}, 2},
		{Color{0, 0, 170}, 4},
		{Color{85, 85, 85}, 8},
		{Color{255, 85, 85}, 9},
		{Color{255, 255, 85}, 11},
		{Color{255, 255, 255}, 15},
	}
	for _, tc := range cases {
		if got := rgbTo16(tc.c); got != tc.want {
			t.Errorf("rgbTo16(%v) = %d, want %d", tc.c, got, tc.want)
		}
	}
}

func TestDetectColorMode(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM", "xterm")
	if got := DetectColorMode(); got != TrueColor {
		t.Errorf("COLORTERM=truecolor => %d, want TrueColor", got)
	}
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm-256color")
	if got := DetectColorMode(); got != Color256 {
		t.Errorf("TERM=xterm-256color => %d, want Color256", got)
	}
	t.Setenv("TERM", "vt100")
	if got := DetectColorMode(); got != Color16 {
		t.Errorf("TERM=vt100 => %d, want Color16", got)
	}
}

func TestFlushEmitsModeEscapes(t *testing.T) {
	cases := []struct {
		mode   ColorMode
		expect string
	}{
		{TrueColor, "\x1b[38;2;200;50;50m\x1b[48;2;10;20;30m"},
		{Color256, "\x1b[38;5;"},
		{Color16, "\x1b[33m\x1b[40mA"},
	}
	for _, tc := range cases {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdout
		os.Stdout = w
		s := NewScreen(2, 1)
		os.Stdout = old
		s.SetColorMode(tc.mode)
		s.Back().Set(0, 0, Cell{Ch: 'A', FG: Color{200, 50, 50}, BG: Color{10, 20, 30}})
		s.Flush()
		w.Close()
		data, _ := io.ReadAll(r)
		if !strings.Contains(string(data), tc.expect) {
			t.Errorf("mode %d output missing %q; got %q", tc.mode, tc.expect, string(data))
		}
	}
}

func TestSetColorModeForcesRepaint(t *testing.T) {
	s := NewScreen(4, 2)
	s.front.Cells[0] = Cell{Ch: 'x', FG: Color{1, 2, 3}, BG: Color{4, 5, 6}}
	s.SetColorMode(Color256)
	if s.mode != Color256 {
		t.Fatalf("mode not set, got %d", s.mode)
	}
	if s.front.Cells[0].Ch != 0 {
		t.Errorf("front buffer not invalidated, cell0 = %+v", s.front.Cells[0])
	}
}
