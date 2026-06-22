package main

import (
	"fmt"

	"awesomeProject/internal/game"
	"awesomeProject/internal/render"
)

const (
	cellW       = 2
	compositeW  = 50
	compositeH  = 22
	boardOffset = 13
)

type theme struct {
	name       string
	background render.Color
	border     render.Color
	text       render.Color
	dim        render.Color
	empty      render.Color
	pieces     map[game.PieceType]render.Color
}

var neon = theme{
	name:       "Neon",
	background: render.RGB(12, 12, 20),
	border:     render.RGB(90, 100, 140),
	text:       render.RGB(225, 230, 245),
	dim:        render.RGB(60, 64, 88),
	empty:      render.RGB(28, 30, 44),
	pieces: map[game.PieceType]render.Color{
		game.I:       render.RGB(60, 220, 230),
		game.O:       render.RGB(240, 215, 70),
		game.T:       render.RGB(200, 90, 230),
		game.S:       render.RGB(90, 220, 110),
		game.Z:       render.RGB(235, 80, 95),
		game.J:       render.RGB(70, 120, 240),
		game.L:       render.RGB(245, 150, 60),
		game.Garbage: render.RGB(120, 120, 130),
	},
}

func block(col render.Color) render.Cell {
	return render.Cell{Ch: '█', FG: col, BG: col}
}

func putBlock(b *render.Buffer, x, y int, col render.Color) {
	top := render.Lerp(col, render.RGB(255, 255, 255), 0.18)
	b.Set(x, y, render.Cell{Ch: '▀', FG: top, BG: col})
	b.Set(x+1, y, render.Cell{Ch: '▀', FG: top, BG: col})
}

func cellOf(r rune, fg, bg render.Color) render.Cell {
	return render.Cell{Ch: r, FG: fg, BG: bg}
}

func drawBox(b *render.Buffer, x, y, w, h int, title string, th theme) {
	for i := 1; i < w-1; i++ {
		b.Set(x+i, y, cellOf('─', th.border, th.background))
		b.Set(x+i, y+h-1, cellOf('─', th.border, th.background))
	}
	for j := 1; j < h-1; j++ {
		b.Set(x, y+j, cellOf('│', th.border, th.background))
		b.Set(x+w-1, y+j, cellOf('│', th.border, th.background))
	}
	b.Set(x, y, cellOf('┌', th.border, th.background))
	b.Set(x+w-1, y, cellOf('┐', th.border, th.background))
	b.Set(x, y+h-1, cellOf('└', th.border, th.background))
	b.Set(x+w-1, y+h-1, cellOf('┘', th.border, th.background))
	if title != "" {
		b.Text(x+2, y, " "+title+" ", th.text, th.background)
	}
}

func drawMini(b *render.Buffer, x, y int, t game.PieceType, th theme) {
	if t == game.Empty {
		return
	}
	p := game.Piece{Type: t, Rotation: 0, X: 0, Y: 0}
	for _, c := range p.Cells() {
		putBlock(b, x+c.X*cellW, y+c.Y, th.pieces[t])
	}
}

func draw(b *render.Buffer, g *game.Game, th theme) {
	b.Reset(th.background)
	ox := (b.W - compositeW) / 2
	oy := (b.H - compositeH) / 2
	if ox < 0 {
		ox = 0
	}
	if oy < 0 {
		oy = 0
	}

	title := "C H A O S   B L O C K S"
	b.Text(ox+(compositeW-len(title))/2, oy-1, title, th.pieces[game.T], th.background)

	drawPlayfield(b, g, th, ox+boardOffset, oy)
	drawLeftPanel(b, g, th, ox, oy)
	drawRightPanel(b, g, th, ox+boardOffset+game.Width*cellW+3, oy)

	if g.Over {
		drawBanner(b, ox+boardOffset, oy, "GAME OVER", "press R to restart", th)
	}
}

func drawPlayfield(b *render.Buffer, g *game.Game, th theme, x, y int) {
	w := game.Width*cellW + 2
	drawBox(b, x, y, w, game.VisibleRows+2, "", th)
	bx := x + 1
	by := y + 1
	for ry := 0; ry < game.VisibleRows; ry++ {
		gy := game.VisibleTop + ry
		for gx := 0; gx < game.Width; gx++ {
			px := bx + gx*cellW
			py := by + ry
			c := g.Board.At(gx, gy)
			if c == game.Empty {
				b.Set(px, py, cellOf('·', th.dim, th.empty))
				b.Set(px+1, py, cellOf(' ', th.dim, th.empty))
			} else {
				putBlock(b, px, py, th.pieces[c])
			}
		}
	}

	gp := g.Ghost()
	ghostCol := render.Lerp(th.pieces[gp.Type], th.background, 0.55)
	for _, c := range gp.Cells() {
		if c.Y < game.VisibleTop {
			continue
		}
		px := bx + c.X*cellW
		py := by + (c.Y - game.VisibleTop)
		b.Set(px, py, cellOf('▒', ghostCol, th.empty))
		b.Set(px+1, py, cellOf('▒', ghostCol, th.empty))
	}

	for _, c := range g.Current.Cells() {
		if c.Y < game.VisibleTop {
			continue
		}
		putBlock(b, bx+c.X*cellW, by+(c.Y-game.VisibleTop), th.pieces[g.Current.Type])
	}
}

func drawLeftPanel(b *render.Buffer, g *game.Game, th theme, x, y int) {
	drawBox(b, x, y, 11, 5, "HOLD", th)
	drawMini(b, x+2, y+1, g.HoldType, th)

	stats := y + 6
	drawBox(b, x, stats, 11, compositeH-6, "", th)
	b.Text(x+2, stats+1, "SCORE", th.dim, th.background)
	b.Text(x+2, stats+2, fmt.Sprintf("%d", g.Score), th.text, th.background)
	b.Text(x+2, stats+4, "LEVEL", th.dim, th.background)
	b.Text(x+2, stats+5, fmt.Sprintf("%d", g.Level), th.text, th.background)
	b.Text(x+2, stats+7, "LINES", th.dim, th.background)
	b.Text(x+2, stats+8, fmt.Sprintf("%d", g.Lines), th.text, th.background)
}

func drawRightPanel(b *render.Buffer, g *game.Game, th theme, x, y int) {
	drawBox(b, x, y, 12, 13, "NEXT", th)
	for i, t := range g.NextQueue {
		if i >= 4 {
			break
		}
		drawMini(b, x+2, y+1+i*3, t, th)
	}

	info := y + 14
	drawBox(b, x, info, 12, compositeH-14, "", th)
	if g.Combo > 1 {
		hot := render.Lerp(th.text, th.pieces[game.Z], float64(g.Combo)/12)
		b.Text(x+2, info+1, fmt.Sprintf("COMBO x%d", g.Combo-1), hot, th.background)
	} else {
		b.Text(x+2, info+1, "COMBO -", th.dim, th.background)
	}
	if g.BackToBack {
		b.Text(x+2, info+3, "BACK2BACK", th.pieces[game.I], th.background)
	} else {
		b.Text(x+2, info+3, "B2B -", th.dim, th.background)
	}
}

func drawBanner(b *render.Buffer, boardX, boardY int, line1, line2 string, th theme) {
	w := game.Width*cellW + 2
	cy := boardY + game.VisibleRows/2
	bg := render.RGB(20, 10, 16)
	for i := 0; i < w; i++ {
		b.Set(boardX+i, cy-1, cellOf(' ', th.text, bg))
		b.Set(boardX+i, cy, cellOf(' ', th.text, bg))
		b.Set(boardX+i, cy+1, cellOf(' ', th.text, bg))
	}
	b.Text(boardX+(w-len(line1))/2, cy, line1, th.pieces[game.Z], bg)
	b.Text(boardX+(w-len(line2))/2, cy+1, line2, th.dim, bg)
}
