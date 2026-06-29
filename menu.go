package main

import (
	"fmt"
	"math"
	"strings"

	"awesomeProject/internal/game"
	"awesomeProject/internal/input"
	"awesomeProject/internal/render"
)

func (a *app) quitKeyHint() string {
	return strings.ToLower(input.KeyLabel(a.keymap[input.Quit]))
}

var (
	menuItems     = []string{"Play", "Settings", "High Scores", "Quit"}
	pauseItems    = []string{"Resume", "Restart", "Main Menu", "Quit"}
	gameOverItems = []string{"Play Again", "Main Menu", "Quit"}
	stripOrder    = []game.PieceType{game.I, game.O, game.T, game.S, game.Z, game.J, game.L}
)

func selStyle(active bool, anim float64, th theme) (render.Color, string) {
	if active {
		pulse := 0.5 + 0.5*math.Sin(anim*6)
		return render.Lerp(th.text, th.pieces[game.T], pulse), "▶ "
	}
	return th.dim, "  "
}

func fillRect(b *render.Buffer, x, y, w, h int, bg render.Color) {
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			b.Set(x+i, y+j, render.Cell{Ch: ' ', FG: bg, BG: bg})
		}
	}
}

func drawPanel(b *render.Buffer, x, y, w, h int, title string, th theme) render.Color {
	bg := render.Lerp(th.background, render.RGB(0, 0, 0), 0.35)
	fillRect(b, x, y, w, h, bg)
	for i := 1; i < w-1; i++ {
		b.Set(x+i, y, cellOf('─', th.border, bg))
		b.Set(x+i, y+h-1, cellOf('─', th.border, bg))
	}
	for j := 1; j < h-1; j++ {
		b.Set(x, y+j, cellOf('│', th.border, bg))
		b.Set(x+w-1, y+j, cellOf('│', th.border, bg))
	}
	b.Set(x, y, cellOf('┌', th.border, bg))
	b.Set(x+w-1, y, cellOf('┐', th.border, bg))
	b.Set(x, y+h-1, cellOf('└', th.border, bg))
	b.Set(x+w-1, y+h-1, cellOf('┘', th.border, bg))
	if title != "" {
		b.Text(x+(w-len(title))/2, y, " "+title+" ", th.text, bg)
	}
	return bg
}

func drawMenuItems(b *render.Buffer, x, y int, items []string, sel int, anim float64, th theme, bg render.Color) {
	for i, it := range items {
		fg, marker := selStyle(i == sel, anim, th)
		b.Text(x, y+i*2, marker+it, fg, bg)
	}
}

func drawValueRow(b *render.Buffer, x, y int, label, value string, active bool, anim float64, th theme, bg render.Color) {
	fg, marker := selStyle(active, anim, th)
	val := th.text
	if active {
		val = th.pieces[game.S]
	}
	b.Text(x, y, marker+label, fg, bg)
	b.Text(x+16, y, "◂ "+value+" ▸", val, bg)
}

func dimOverlay(b *render.Buffer, th theme) {
	dark := render.Lerp(th.background, render.RGB(0, 0, 0), 0.55)
	for i := range b.Cells {
		c := b.Cells[i]
		c.FG = render.Lerp(c.FG, dark, 0.62)
		c.BG = render.Lerp(c.BG, dark, 0.62)
		b.Cells[i] = c
	}
}

func (a *app) drawHeader(b *render.Buffer, title string, th theme) {
	sub := "C H A O S   B L O C K S"
	b.Text((b.W-len(sub))/2, 1, sub, th.dim, th.background)
	b.Text((b.W-len(title))/2, 3, title, th.pieces[game.T], th.background)
}

func (a *app) drawPieceStrip(b *render.Buffer, x, y int, th theme) {
	for i, t := range stripOrder {
		putBlock(b, x+i*2, y, th.pieces[t])
	}
}

func (a *app) renderTooSmall(b *render.Buffer) {
	th := themes[a.themeIdx]
	b.Reset(th.background)
	msg := fmt.Sprintf("Resize terminal to at least %dx%d", compositeW, compositeH+2)
	cur := fmt.Sprintf("currently %dx%d", b.W, b.H)
	b.Text((b.W-len(msg))/2, b.H/2, msg, th.pieces[game.Z], th.background)
	b.Text((b.W-len(cur))/2, b.H/2+1, cur, th.dim, th.background)
}

func (a *app) renderMenu(b *render.Buffer) {
	th := themes[a.themeIdx]
	b.Reset(th.background)
	cx := b.W / 2
	cy := b.H / 2
	title := "C H A O S   B L O C K S"
	b.Text(cx-len(title)/2, cy-6, title, th.pieces[game.T], th.background)
	a.drawPieceStrip(b, cx-7, cy-4, th)
	drawMenuItems(b, cx-7, cy-1, menuItems, a.mainSel, a.anim, th, th.background)
	hint := "↑/↓ move   ⏎ select   " + a.quitKeyHint() + " quit"
	b.Text(cx-len(hint)/2, b.H-2, hint, th.dim, th.background)
}

func (a *app) renderModeSelect(b *render.Buffer) {
	th := themes[a.themeIdx]
	b.Reset(th.background)
	cx := b.W / 2
	a.drawHeader(b, "SELECT MODE", th)
	top := b.H/2 - 5
	for i, m := range modes {
		fg, marker := selStyle(i == a.modeSel, a.anim, th)
		b.Text(cx-13, top+i*2, marker+m.name, fg, th.background)
	}
	sel := modes[a.modeSel]
	b.Text(cx-len(sel.tagline)/2, top+len(modes)*2, sel.tagline, th.text, th.background)
	drawValueRow(b, cx-13, top+len(modes)*2+2, "Start Level", fmt.Sprintf("%d", a.startLevel), false, a.anim, th, th.background)
	hint := "↑/↓ mode   ◂/▸ level   ⏎ start   esc back"
	b.Text(cx-len(hint)/2, b.H-2, hint, th.dim, th.background)
}

func (a *app) renderSettings(b *render.Buffer) {
	th := themes[a.themeIdx]
	b.Reset(th.background)
	cx := b.W / 2
	a.drawHeader(b, "SETTINGS", th)
	top := b.H/2 - 4
	drawValueRow(b, cx-13, top, "Theme", th.name, a.settingSel == setTheme, a.anim, th, th.background)
	drawValueRow(b, cx-13, top+2, "Start Level", fmt.Sprintf("%d", a.startLevel), a.settingSel == setStartLevel, a.anim, th, th.background)
	drawValueRow(b, cx-13, top+4, "Color Mode", a.colorModeLabel(), a.settingSel == setColorMode, a.anim, th, th.background)
	kbFg, kbMarker := selStyle(a.settingSel == setKeybinds, a.anim, th)
	b.Text(cx-13, top+6, kbMarker+"Key Bindings", kbFg, th.background)
	fg, marker := selStyle(a.settingSel == setBack, a.anim, th)
	b.Text(cx-13, top+8, marker+"Back", fg, th.background)
	hint := "↑/↓ row   ◂/▸ change   ⏎ open   esc back"
	b.Text(cx-len(hint)/2, b.H-2, hint, th.dim, th.background)
}

func (a *app) renderKeybinds(b *render.Buffer) {
	th := themes[a.themeIdx]
	b.Reset(th.background)
	cx := b.W / 2
	a.drawHeader(b, "KEY BINDINGS", th)
	top := 5
	for i, e := range input.Bindable {
		fg, marker := selStyle(i == a.bindSel, a.anim, th)
		b.Text(cx-16, top+i, marker+input.Label(e), fg, th.background)
		key := input.KeyLabel(a.keymap[e])
		val := th.text
		if i == a.bindSel {
			val = th.pieces[game.S]
			if a.capturing {
				key = "press a key…"
				val = th.pieces[game.O]
			}
		}
		b.Text(cx+2, top+i, key, val, th.background)
	}
	resetIdx := len(input.Bindable)
	rFg, rMarker := selStyle(a.bindSel == resetIdx, a.anim, th)
	b.Text(cx-16, top+resetIdx+1, rMarker+"Reset to Defaults", rFg, th.background)
	bFg, bMarker := selStyle(a.bindSel == resetIdx+1, a.anim, th)
	b.Text(cx-16, top+resetIdx+2, bMarker+"Back", bFg, th.background)
	if a.bindMsg != "" {
		b.Text(cx-16, top+resetIdx+4, a.bindMsg, th.dim, th.background)
	}
	hint := "↑/↓ row   ⏎ rebind   esc back"
	b.Text(cx-len(hint)/2, b.H-2, hint, th.dim, th.background)
}

func (a *app) renderScores(b *render.Buffer) {
	th := themes[a.themeIdx]
	b.Reset(th.background)
	cx := b.W / 2
	a.drawHeader(b, "HIGH SCORES", th)
	tabY := b.H/2 - 6
	tx := cx - 22
	for i, m := range modes {
		fg := th.dim
		if i == a.scoreTab {
			fg = th.pieces[game.T]
		}
		b.Text(tx, tabY, m.name, fg, th.background)
		tx += len(m.name) + 3
	}
	m := modes[a.scoreTab]
	list := a.board.entries(m.kind)
	valHead := "SCORE"
	if m.timed() {
		valHead = "TIME"
	}
	hy := tabY + 2
	b.Text(cx-18, hy, fmt.Sprintf("%-4s %-11s %-6s %-6s %-6s", "#", valHead, "LEVEL", "LINES", "COMBO"), th.dim, th.background)
	if len(list) == 0 {
		msg := "No scores yet — go play!"
		b.Text(cx-len(msg)/2, hy+2, msg, th.dim, th.background)
	}
	for i, e := range list {
		val := fmt.Sprintf("%d", e.score)
		if m.timed() {
			val = formatTime(e.time)
		}
		fg := th.text
		if m.kind == a.recentKind && i == a.recentRank {
			fg = th.pieces[game.O]
		}
		row := fmt.Sprintf("%-4d %-11s %-6d %-6d %-6d", i+1, val, e.level, e.lines, e.combo)
		b.Text(cx-18, hy+2+i, row, fg, th.background)
	}
	hint := "◂/▸ mode   esc back"
	b.Text(cx-len(hint)/2, b.H-2, hint, th.dim, th.background)
}

func (a *app) renderPlaying(b *render.Buffer) {
	th := themes[a.themeIdx]
	s := a.sess
	sx, sy := s.eng.ShakeOffset()
	draw(b, s.g, s.ch, th, sx, sy)
	a.drawPlayHUD(b, sx, sy, th)
	s.eng.Apply(b, sx, sy)
}

func (a *app) drawPlayHUD(b *render.Buffer, sx, sy int, th theme) {
	s := a.sess
	ox, oy := origin(b.W, b.H)
	ox += sx
	oy += sy
	b.Text(ox, oy-1, s.mode.name, th.dim, th.background)
	var info string
	col := th.dim
	if s.mode.timed() {
		info = fmt.Sprintf("%d/%d %s", s.g.Lines, s.mode.sprintLines, formatTime(s.elapsed))
		col = th.text
	} else {
		info = formatTime(s.elapsed)
	}
	b.Text(ox+compositeW-len(info), oy-1, info, col, th.background)
}

func (a *app) renderPaused(b *render.Buffer) {
	th := themes[a.themeIdx]
	s := a.sess
	draw(b, s.g, s.ch, th, 0, 0)
	a.drawPlayHUD(b, 0, 0, th)
	dimOverlay(b, th)
	a.drawCenterMenu(b, "PAUSED", pauseItems, a.pauseSel, th)
	hint := "↑/↓ move   ⏎ select   esc resume"
	b.Text((b.W-len(hint))/2, b.H-2, hint, th.dim, th.background)
}

func (a *app) drawCenterMenu(b *render.Buffer, title string, items []string, sel int, th theme) {
	w := 24
	h := len(items)*2 + 4
	x := (b.W - w) / 2
	y := (b.H - h) / 2
	bg := drawPanel(b, x, y, w, h, title, th)
	drawMenuItems(b, x+5, y+2, items, sel, a.anim, th, bg)
}

func (a *app) renderGameOver(b *render.Buffer) {
	th := themes[a.themeIdx]
	s := a.sess
	b.Reset(th.background)
	sx, sy := s.eng.ShakeOffset()
	ox, oy := origin(b.W, b.H)
	dox, doy := ox+sx, oy+sy
	title := "C H A O S   B L O C K S"
	b.Text(dox+(compositeW-len(title))/2, doy-1, title, th.pieces[game.T], th.background)
	a.drawCollapseBoard(b, s, dox+boardOffset, doy, th)
	drawLeftPanel(b, s.g, th, dox, doy)
	drawRightPanel(b, s.g, s.ch, th, dox+boardOffset+game.Width*cellW+3, doy)
	s.eng.Apply(b, sx, sy)
	if int(s.collapse) >= game.VisibleRows {
		a.drawGameOverPanel(b, th)
		hint := "↑/↓ move   ⏎ select   esc menu   " + a.quitKeyHint() + " quit"
		b.Text((b.W-len(hint))/2, b.H-2, hint, th.dim, th.background)
	}
}

func (a *app) drawCollapseBoard(b *render.Buffer, s *session, x, y int, th theme) {
	w := game.Width*cellW + 2
	drawBox(b, x, y, w, game.VisibleRows+2, "", th)
	bx := x + 1
	by := y + 1
	gone := int(s.collapse)
	for ry := 0; ry < game.VisibleRows; ry++ {
		survived := ry < game.VisibleRows-gone
		for gx := 0; gx < game.Width; gx++ {
			px := bx + gx*cellW
			py := by + ry
			c := s.g.Board.At(gx, game.VisibleTop+ry)
			if c == game.Empty || !survived {
				b.Set(px, py, cellOf('·', th.dim, th.empty))
				b.Set(px+1, py, cellOf(' ', th.dim, th.empty))
			} else {
				putBlock(b, px, py, th.pieces[c])
			}
		}
	}
}

func (a *app) drawGameOverPanel(b *render.Buffer, th theme) {
	s := a.sess
	w := 26
	h := 16
	x := (b.W - w) / 2
	y := (b.H - h) / 2
	bg := drawPanel(b, x, y, w, h, "", th)
	head := "GAME OVER"
	headCol := th.pieces[game.Z]
	if s.won {
		head = "SPRINT CLEAR!"
		headCol = th.pieces[game.S]
	}
	b.Text(x+(w-len(head))/2, y+1, head, headCol, bg)

	stat := func(row int, label, val string) {
		b.Text(x+3, y+row, label, th.dim, bg)
		b.Text(x+w-3-len(val), y+row, val, th.text, bg)
	}
	stat(3, "Score", fmt.Sprintf("%d", s.g.Score))
	stat(4, "Level", fmt.Sprintf("%d", s.g.Level))
	stat(5, "Lines", fmt.Sprintf("%d", s.g.Lines))
	stat(6, "Max Combo", fmt.Sprintf("%d", comboChain(s.maxCombo)))
	stat(7, "Time", formatTime(s.elapsed))
	if s.lastRank >= 0 {
		badge := fmt.Sprintf("New #%d!", s.lastRank+1)
		b.Text(x+(w-len(badge))/2, y+8, badge, th.pieces[game.O], bg)
	}
	drawMenuItems(b, x+5, y+9, gameOverItems, a.overSel, a.anim, th, bg)
}
