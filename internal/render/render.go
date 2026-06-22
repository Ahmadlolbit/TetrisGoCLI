package render

import (
	"bufio"
	"fmt"
	"os"
)

type Color struct {
	R uint8
	G uint8
	B uint8
}

func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b}
}

type Cell struct {
	Ch rune
	FG Color
	BG Color
}

type Buffer struct {
	W     int
	H     int
	Cells []Cell
}

func NewBuffer(w, h int) *Buffer {
	b := &Buffer{W: w, H: h, Cells: make([]Cell, w*h)}
	b.Reset(Color{0, 0, 0})
	return b
}

func (b *Buffer) Reset(bg Color) {
	blank := Cell{Ch: ' ', FG: Color{180, 180, 180}, BG: bg}
	for i := range b.Cells {
		b.Cells[i] = blank
	}
}

func (b *Buffer) Set(x, y int, c Cell) {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return
	}
	b.Cells[y*b.W+x] = c
}

func (b *Buffer) Text(x, y int, s string, fg, bg Color) {
	col := x
	for _, r := range s {
		b.Set(col, y, Cell{Ch: r, FG: fg, BG: bg})
		col++
	}
}

type Screen struct {
	W     int
	H     int
	back  *Buffer
	front *Buffer
	out   *bufio.Writer
}

func NewScreen(w, h int) *Screen {
	return &Screen{
		W:     w,
		H:     h,
		back:  NewBuffer(w, h),
		front: NewBuffer(w, h),
		out:   bufio.NewWriter(os.Stdout),
	}
}

func (s *Screen) Back() *Buffer {
	return s.back
}

func (s *Screen) Enter() {
	fmt.Fprint(s.out, "\x1b[?1049h\x1b[?25l\x1b[2J")
	s.out.Flush()
	for i := range s.front.Cells {
		s.front.Cells[i] = Cell{Ch: 0}
	}
}

func (s *Screen) Leave() {
	fmt.Fprint(s.out, "\x1b[0m\x1b[?25h\x1b[?1049l")
	s.out.Flush()
}

func (s *Screen) Flush() {
	var curFG, curBG Color
	hasColor := false
	lastX, lastY := -2, -2
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			i := y*s.W + x
			next := s.back.Cells[i]
			if next == s.front.Cells[i] {
				continue
			}
			if x != lastX+1 || y != lastY {
				fmt.Fprintf(s.out, "\x1b[%d;%dH", y+1, x+1)
			}
			if !hasColor || next.FG != curFG || next.BG != curBG {
				fmt.Fprintf(s.out, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm",
					next.FG.R, next.FG.G, next.FG.B, next.BG.R, next.BG.G, next.BG.B)
				curFG, curBG = next.FG, next.BG
				hasColor = true
			}
			ch := next.Ch
			if ch == 0 {
				ch = ' '
			}
			s.out.WriteRune(ch)
			s.front.Cells[i] = next
			lastX, lastY = x, y
		}
	}
	s.out.Flush()
}

func Lerp(a, b Color, t float64) Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	mix := func(x, y uint8) uint8 {
		return uint8(float64(x) + (float64(y)-float64(x))*t)
	}
	return Color{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B)}
}
