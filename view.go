package main

import (
	"fmt"
	"strconv"

	"awesomeProject/internal/chaos"
	"awesomeProject/internal/effects"
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

func origin(w, h int) (int, int) {
	ox := (w - compositeW) / 2
	oy := (h - compositeH) / 2
	if ox < 0 {
		ox = 0
	}
	if oy < 0 {
		oy = 0
	}
	return ox, oy
}

func boardInterior(ox, oy int) (int, int) {
	return ox + boardOffset + 1, oy + 1
}

func pieceColor(th theme, ch *chaos.Engine, t game.PieceType) render.Color {
	return th.pieces[ch.Remap(t)]
}

func draw(b *render.Buffer, g *game.Game, ch *chaos.Engine, th theme, shakeX, shakeY int) {
	b.Reset(th.background)
	ox, oy := origin(b.W, b.H)
	ox += shakeX
	oy += shakeY

	title := "C H A O S   B L O C K S"
	b.Text(ox+(compositeW-len(title))/2, oy-1, title, th.pieces[game.T], th.background)

	drawPlayfield(b, g, ch, th, ox+boardOffset, oy)
	drawLeftPanel(b, g, th, ox, oy)
	drawRightPanel(b, g, ch, th, ox+boardOffset+game.Width*cellW+3, oy)

	if g.Over {
		drawBanner(b, ox+boardOffset, oy, "GAME OVER", "press R to restart", th)
	}
}

func drawPlayfield(b *render.Buffer, g *game.Game, ch *chaos.Engine, th theme, x, y int) {
	w := game.Width*cellW + 2
	drawBox(b, x, y, w, game.VisibleRows+2, "", th)
	bx := x + 1
	by := y + 1
	dimmed := ch.Active() == chaos.LightsDim
	for ry := 0; ry < game.VisibleRows; ry++ {
		gy := game.VisibleTop + ry
		for gx := 0; gx < game.Width; gx++ {
			px := bx + gx*cellW
			py := by + ry
			c := g.Board.At(gx, gy)
			if c == game.Empty {
				empty := th.empty
				if dimmed {
					empty = render.Lerp(empty, th.background, 0.6)
				}
				b.Set(px, py, cellOf('·', th.dim, empty))
				b.Set(px+1, py, cellOf(' ', th.dim, empty))
			} else {
				col := pieceColor(th, ch, c)
				if dimmed {
					col = render.Lerp(col, th.background, 0.72)
				}
				putBlock(b, px, py, col)
			}
		}
	}

	gp := g.Ghost()
	ghostCol := render.Lerp(pieceColor(th, ch, gp.Type), th.background, 0.55)
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
		putBlock(b, bx+c.X*cellW, by+(c.Y-game.VisibleTop), pieceColor(th, ch, g.Current.Type))
	}
}

func drawLeftPanel(b *render.Buffer, g *game.Game, th theme, x, y int) {
	drawBox(b, x, y, 11, 5, "HOLD", th)
	drawMini(b, x+2, y+1, g.HoldType, th)

	stats := y + 6
	drawBox(b, x, stats, 11, compositeH-6, "", th)
	b.Text(x+2, stats+1, "SCORE", th.dim, th.background)
	b.Text(x+2, stats+2, strconv.Itoa(g.Score), th.text, th.background)
	b.Text(x+2, stats+4, "LEVEL", th.dim, th.background)
	b.Text(x+2, stats+5, strconv.Itoa(g.Level), th.text, th.background)
	b.Text(x+2, stats+7, "LINES", th.dim, th.background)
	b.Text(x+2, stats+8, strconv.Itoa(g.Lines), th.text, th.background)
}

func drawRightPanel(b *render.Buffer, g *game.Game, ch *chaos.Engine, th theme, x, y int) {
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
	drawComboMeter(b, g.Combo, th, x+2, info+2)
	if g.BackToBack {
		b.Text(x+2, info+3, "BACK2BACK", th.pieces[game.I], th.background)
	} else {
		b.Text(x+2, info+3, "B2B -", th.dim, th.background)
	}

	drawChaosMeter(b, ch, th, x+2, info+4)
}

func drawComboMeter(b *render.Buffer, combo int, th theme, x, y int) {
	count := combo - 1
	if count < 0 {
		count = 0
	}
	barW := 8
	if count > barW {
		count = barW
	}
	for i := 0; i < barW; i++ {
		col := th.dim
		if i < count {
			col = render.Lerp(th.pieces[game.I], th.pieces[game.Z], float64(i)/float64(barW-1))
		}
		b.Set(x+i, y, render.Cell{Ch: '▮', FG: col, BG: th.background})
	}
}

func drawChaosMeter(b *render.Buffer, ch *chaos.Engine, th theme, x, y int) {
	if !ch.Enabled {
		b.Text(x, y, "CLASSIC", th.dim, th.background)
		return
	}
	b.Text(x, y, "CHAOS", th.pieces[game.T], th.background)
	barW := 8
	fill := int(ch.Meter() * float64(barW))
	for i := 0; i < barW; i++ {
		col := th.dim
		if i < fill {
			col = render.Lerp(th.pieces[game.S], th.pieces[game.Z], float64(i)/float64(barW))
		}
		b.Set(x+i, y+1, render.Cell{Ch: '▮', FG: col, BG: th.background})
	}
	if name, firing := ch.Status(); name != "" {
		col := th.text
		if firing {
			col = th.pieces[game.Z]
		}
		b.Text(x, y+2, name, col, th.background)
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

func spawnLockEffects(e *effects.Engine, res game.LockResult, ox, oy int, th theme) {
	bx, by := boardInterior(ox, oy)
	w := game.Width * cellW
	for _, gy := range res.ClearedRows {
		py := by + (gy - game.VisibleTop)
		e.Flash(bx, py, w, 1, render.RGB(255, 255, 255), 0.35)
		for gx := 0; gx < game.Width; gx++ {
			e.Burst(float64(bx+gx*cellW), float64(py), th.text, 2, 6)
		}
	}
	switch {
	case res.TSpin != game.TSpinNone && res.Lines > 0:
		e.Flash(bx, by, w, game.VisibleRows, th.pieces[game.T], 0.4)
		e.Shake(1.6, 0.35)
	case res.Lines >= 4:
		e.Flash(bx, by, w, game.VisibleRows, th.pieces[game.I], 0.4)
		e.Shake(1.9, 0.4)
	case res.Lines > 0:
		e.Shake(0.5+float64(res.Lines)*0.25, 0.18)
	}
	if res.Combo > 2 {
		mag := float64(res.Combo-1) * 0.2
		if mag > 1.4 {
			mag = 1.4
		}
		e.Shake(mag, 0.16)
	}
}

func spawnHardDrop(e *effects.Engine, landing game.Piece, ox, oy int, th theme) {
	bx, by := boardInterior(ox, oy)
	for _, c := range landing.Cells() {
		if c.Y < game.VisibleTop {
			continue
		}
		e.Burst(float64(bx+c.X*cellW), float64(by+(c.Y-game.VisibleTop)), th.pieces[landing.Type], 3, 5)
	}
	e.Shake(0.7, 0.12)
}

func spawnLevelUp(e *effects.Engine, ox, oy int, th theme) {
	bx, by := boardInterior(ox, oy)
	e.Flash(bx, by, game.Width*cellW, game.VisibleRows, th.pieces[game.S], 0.6)
	e.Shake(1.2, 0.3)
}

func spawnChaos(e *effects.Engine, k chaos.Kind, ox, oy int, th theme) {
	bx, by := boardInterior(ox, oy)
	w := game.Width * cellW
	col := th.pieces[game.T]
	if k == chaos.BonusFrenzy {
		col = th.pieces[game.O]
	}
	e.Flash(bx, by, w, game.VisibleRows, col, 0.7)
	e.Shake(1.5, 0.35)
	for gx := 0; gx < game.Width; gx++ {
		e.Burst(float64(bx+gx*cellW), float64(by), col, 2, 7)
	}
}

var themes = []theme{neon, synthwave, monoglow}

var synthwave = theme{
	name:       "Synthwave",
	background: render.RGB(20, 8, 34),
	border:     render.RGB(120, 70, 160),
	text:       render.RGB(245, 220, 250),
	dim:        render.RGB(90, 55, 110),
	empty:      render.RGB(34, 16, 52),
	pieces: map[game.PieceType]render.Color{
		game.I:       render.RGB(80, 230, 240),
		game.O:       render.RGB(255, 200, 90),
		game.T:       render.RGB(235, 90, 200),
		game.S:       render.RGB(120, 240, 150),
		game.Z:       render.RGB(255, 70, 130),
		game.J:       render.RGB(120, 110, 255),
		game.L:       render.RGB(255, 140, 90),
		game.Garbage: render.RGB(110, 90, 130),
	},
}

var monoglow = theme{
	name:       "Mono-glow",
	background: render.RGB(8, 14, 12),
	border:     render.RGB(70, 120, 90),
	text:       render.RGB(190, 245, 210),
	dim:        render.RGB(45, 80, 60),
	empty:      render.RGB(16, 26, 20),
	pieces: map[game.PieceType]render.Color{
		game.I:       render.RGB(120, 240, 170),
		game.O:       render.RGB(170, 250, 150),
		game.T:       render.RGB(90, 220, 140),
		game.S:       render.RGB(140, 250, 180),
		game.Z:       render.RGB(80, 200, 130),
		game.J:       render.RGB(100, 230, 160),
		game.L:       render.RGB(150, 245, 175),
		game.Garbage: render.RGB(70, 110, 85),
	},
}
