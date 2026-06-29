package render

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ColorMode int

const (
	ColorAuto ColorMode = iota
	TrueColor
	Color256
	Color16
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
	mode  ColorMode
}

func NewScreen(w, h int) *Screen {
	return &Screen{
		W:     w,
		H:     h,
		back:  NewBuffer(w, h),
		front: NewBuffer(w, h),
		out:   bufio.NewWriter(os.Stdout),
		mode:  TrueColor,
	}
}

func (s *Screen) SetColorMode(m ColorMode) {
	s.mode = m
	for i := range s.front.Cells {
		s.front.Cells[i] = Cell{Ch: 0}
	}
	fmt.Fprint(s.out, "\x1b[2J")
	s.out.Flush()
}

func (s *Screen) Back() *Buffer {
	return s.back
}

func (s *Screen) Resize(w, h int) {
	s.W = w
	s.H = h
	s.back = NewBuffer(w, h)
	s.front = &Buffer{W: w, H: h, Cells: make([]Cell, w*h)}
	fmt.Fprint(s.out, "\x1b[2J")
	s.out.Flush()
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
				s.emitColor(next.FG, next.BG)
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

func (s *Screen) emitColor(fg, bg Color) {
	switch s.mode {
	case Color256:
		fmt.Fprintf(s.out, "\x1b[38;5;%dm\x1b[48;5;%dm", rgbTo256(fg), rgbTo256(bg))
	case Color16:
		fmt.Fprintf(s.out, "\x1b[%dm\x1b[%dm", fgSeq(rgbTo16(fg)), bgSeq(rgbTo16(bg)))
	default:
		fmt.Fprintf(s.out, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm",
			fg.R, fg.G, fg.B, bg.R, bg.G, bg.B)
	}
}

func fgSeq(i int) int {
	if i < 8 {
		return 30 + i
	}
	return 90 + i - 8
}

func bgSeq(i int) int {
	if i < 8 {
		return 40 + i
	}
	return 100 + i - 8
}

func sq(x int) int {
	return x * x
}

var cubeLevels = [6]int{0, 95, 135, 175, 215, 255}

func rgbTo256(c Color) int {
	nearest := func(v int) int {
		best, bd := 0, 1<<30
		for i, lv := range cubeLevels {
			d := v - lv
			if d < 0 {
				d = -d
			}
			if d < bd {
				bd, best = d, i
			}
		}
		return best
	}
	r, g, b := int(c.R), int(c.G), int(c.B)
	ri, gi, bi := nearest(r), nearest(g), nearest(b)
	cubeDist := sq(r-cubeLevels[ri]) + sq(g-cubeLevels[gi]) + sq(b-cubeLevels[bi])

	gray := (r + g + b) / 3
	gi2 := (gray - 3) / 10
	if gi2 < 0 {
		gi2 = 0
	}
	if gi2 > 23 {
		gi2 = 23
	}
	grayVal := 8 + gi2*10
	grayDist := sq(r-grayVal) + sq(g-grayVal) + sq(b-grayVal)

	if grayDist < cubeDist {
		return 232 + gi2
	}
	return 16 + 36*ri + 6*gi + bi
}

var ansi16 = [16]Color{
	{0, 0, 0}, {170, 0, 0}, {0, 170, 0}, {170, 85, 0},
	{0, 0, 170}, {170, 0, 170}, {0, 170, 170}, {170, 170, 170},
	{85, 85, 85}, {255, 85, 85}, {85, 255, 85}, {255, 255, 85},
	{85, 85, 255}, {255, 85, 255}, {85, 255, 255}, {255, 255, 255},
}

func rgbTo16(c Color) int {
	best, bd := 0, 1<<30
	for i, p := range ansi16 {
		d := sq(int(c.R)-int(p.R)) + sq(int(c.G)-int(p.G)) + sq(int(c.B)-int(p.B))
		if d < bd {
			bd, best = d, i
		}
	}
	return best
}

func DetectColorMode() ColorMode {
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	if strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit") {
		return TrueColor
	}
	if strings.Contains(os.Getenv("TERM"), "256") {
		return Color256
	}
	return Color16
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
