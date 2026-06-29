package main

import (
	"time"

	"awesomeProject/internal/chaos"
	"awesomeProject/internal/effects"
	"awesomeProject/internal/game"
	"awesomeProject/internal/input"
	"awesomeProject/internal/render"
	"awesomeProject/internal/store"
)

type appScreen int

const (
	scrMenu appScreen = iota
	scrModeSelect
	scrSettings
	scrKeybinds
	scrScores
	scrPlaying
	scrPaused
	scrGameOver
)

const (
	setTheme = iota
	setStartLevel
	setColorMode
	setKeybinds
	setBack
	settingsRows
)

type session struct {
	g            *game.Game
	eng          *effects.Engine
	ch           *chaos.Engine
	mode         mode
	elapsed      float64
	maxCombo     int
	prevCombo    int
	collapse     float64
	won          bool
	recorded     bool
	chaosToggled bool
	lastRank     int
}

type app struct {
	scr   *render.Screen
	in    *input.Reader
	board *scoreboard

	state appScreen
	anim  float64

	themeIdx   int
	startLevel int
	colorMode  int

	mainSel    int
	modeSel    int
	settingSel int
	scoreTab   int
	pauseSel   int
	overSel    int

	keymap        input.Keymap
	bindSel       int
	capturing     bool
	captureAction input.Event
	bindMsg       string

	recentKind modeKind
	recentRank int

	tooSmall bool

	sess *session
}

func newApp(scr *render.Screen, in *input.Reader) *app {
	a := &app{
		scr:        scr,
		in:         in,
		board:      newScoreboard(),
		state:      scrMenu,
		startLevel: 1,
		recentRank: -1,
	}
	a.tooSmall = scr.W < compositeW || scr.H < compositeH+2
	a.loadState()
	return a
}

func (a *app) loadState() {
	st := store.Load()
	a.themeIdx = wrap(st.Settings.Theme, len(themes))
	a.startLevel = clampLevel(st.Settings.StartLevel)
	a.colorMode = clampColorMode(st.Settings.ColorMode)
	a.scr.SetColorMode(resolveColorMode(a.colorMode))
	a.keymap = input.ImportKeymap(st.Keymap)
	a.in.SetKeymap(a.keymap)
	a.board.load(st.Scores)
}

func (a *app) persist() {
	store.Save(store.State{
		Settings: store.Settings{Theme: a.themeIdx, StartLevel: a.startLevel, ColorMode: a.colorMode},
		Scores:   a.board.export(),
		Keymap:   input.ExportKeymap(a.keymap),
	})
}

func newSession(m mode, seed int64, level int) *session {
	ch := chaos.New(seed, m.chaosEnabled)
	ch.FreqScale = m.chaosFreq
	return &session{
		g:        game.NewGameAtLevel(seed, level),
		eng:      effects.New(seed),
		ch:       ch,
		mode:     m,
		lastRank: -1,
	}
}

func (a *app) run(done <-chan struct{}, resize <-chan [2]int) {
	defer a.persist()
	ticker := time.NewTicker(frameDelay)
	defer ticker.Stop()
	dt := frameDelay.Seconds()
	frame := 0

	a.render()
	a.scr.Flush()

	for {
		select {
		case <-done:
			return
		case sz := <-resize:
			a.onResize(sz[0], sz[1])
			a.render()
			a.scr.Flush()
		case ev := <-a.in.Events():
			if !a.handle(ev) {
				return
			}
			a.render()
			a.scr.Flush()
		case k := <-a.in.Captures():
			a.onCapture(k)
			a.render()
			a.scr.Flush()
		case <-ticker.C:
			frame++
			a.update(dt)
			if a.animating() || frame%3 == 0 {
				a.render()
				a.scr.Flush()
			}
		}
	}
}

func (a *app) animating() bool {
	return !a.tooSmall && (a.state == scrPlaying || a.state == scrGameOver)
}

func (a *app) onResize(cols, rows int) {
	a.scr.Resize(cols, rows)
	a.tooSmall = cols < compositeW || rows < compositeH+2
}

func (a *app) update(dt float64) {
	if a.tooSmall {
		return
	}
	a.anim += dt
	switch a.state {
	case scrPlaying:
		a.updatePlaying(dt)
	case scrGameOver:
		a.updateGameOver(dt)
	}
	if a.sess != nil {
		a.sess.eng.Update(dt)
	}
}

func (a *app) render() {
	if a.tooSmall {
		a.renderTooSmall(a.scr.Back())
		return
	}
	switch a.state {
	case scrMenu:
		a.renderMenu(a.scr.Back())
	case scrModeSelect:
		a.renderModeSelect(a.scr.Back())
	case scrSettings:
		a.renderSettings(a.scr.Back())
	case scrKeybinds:
		a.renderKeybinds(a.scr.Back())
	case scrScores:
		a.renderScores(a.scr.Back())
	case scrPlaying:
		a.renderPlaying(a.scr.Back())
	case scrPaused:
		a.renderPaused(a.scr.Back())
	case scrGameOver:
		a.renderGameOver(a.scr.Back())
	}
}

func (a *app) updatePlaying(dt float64) {
	s := a.sess
	th := themes[a.themeIdx]
	ox, oy := origin(a.scr.W, a.scr.H)
	if fired, ok := s.ch.Update(dt, s.g); ok {
		spawnChaos(s.eng, fired, ox, oy, th)
	}
	s.g.GravityScale = s.mode.gravityScale * s.ch.GravityScale()
	s.g.ScoreScale = s.ch.ScoreScale()
	before := s.g.Level
	for _, res := range s.g.Tick(dt) {
		spawnLockEffects(s.eng, res, ox, oy, th)
		s.ch.OnPiece()
		s.observe(res)
		a.comboFeedback(res, ox, oy, th)
	}
	if s.g.Level > before {
		spawnLevelUp(s.eng, ox, oy, th)
	}
	s.elapsed += dt
	a.checkEnd()
}

func (a *app) updateGameOver(dt float64) {
	s := a.sess
	if int(s.collapse) >= game.VisibleRows {
		return
	}
	th := themes[a.themeIdx]
	ox, oy := origin(a.scr.W, a.scr.H)
	bx, by := boardInterior(ox, oy)
	prev := int(s.collapse)
	s.collapse += dt * 60
	now := int(s.collapse)
	if now > game.VisibleRows {
		now = game.VisibleRows
		s.collapse = float64(game.VisibleRows)
	}
	for r := prev; r < now; r++ {
		ry := game.VisibleRows - 1 - r
		gy := game.VisibleTop + ry
		for gx := 0; gx < game.Width; gx++ {
			if c := s.g.Board.At(gx, gy); c != game.Empty {
				s.eng.Burst(float64(bx+gx*cellW), float64(by+ry), th.pieces[c], 3, 7)
			}
		}
	}
}

func (s *session) observe(res game.LockResult) {
	if res.Combo > s.maxCombo {
		s.maxCombo = res.Combo
	}
}

func (a *app) comboFeedback(res game.LockResult, ox, oy int, th theme) {
	s := a.sess
	if s.prevCombo >= 2 && res.Combo == 0 {
		bx := ox + boardOffset + game.Width*cellW + 5
		by := oy + 16
		s.eng.Flash(bx, by, 8, 1, th.pieces[game.Z], 0.3)
		for i := 0; i < 8; i++ {
			s.eng.Burst(float64(bx+i), float64(by), th.dim, 1, 4)
		}
	}
	s.prevCombo = res.Combo
}

func (a *app) checkEnd() {
	s := a.sess
	if s.mode.timed() && s.g.Lines >= s.mode.sprintLines {
		s.won = true
		s.g.Over = true
	}
	if s.g.Over && a.state == scrPlaying {
		a.state = scrGameOver
		a.overSel = 0
		s.collapse = 0
		s.eng.Shake(2.4, 0.5)
		a.recordScore()
	}
}

func (a *app) recordScore() {
	s := a.sess
	if s.recorded {
		return
	}
	s.recorded = true
	if s.chaosToggled {
		return
	}
	if s.mode.timed() && !s.won {
		return
	}
	e := scoreEntry{
		score: s.g.Score,
		level: s.g.Level,
		lines: s.g.Lines,
		combo: comboChain(s.maxCombo),
		time:  s.elapsed,
	}
	s.lastRank = a.board.record(s.mode, e)
	a.recentKind = s.mode.kind
	a.recentRank = s.lastRank
	a.persist()
}

func (a *app) startGame(m mode) {
	a.recentRank = -1
	a.sess = newSession(m, time.Now().UnixNano(), a.startLevel)
	a.state = scrPlaying
}

func (a *app) restart() {
	if a.sess == nil {
		return
	}
	a.recentRank = -1
	a.sess = newSession(a.sess.mode, time.Now().UnixNano(), a.startLevel)
	a.state = scrPlaying
}

func (a *app) toMenu() {
	a.sess = nil
	a.state = scrMenu
}

func (a *app) handle(ev input.Event) bool {
	if ev == input.Quit {
		return false
	}
	if a.tooSmall {
		return true
	}
	if ev == input.ThemeNext {
		a.themeIdx = wrap(a.themeIdx+1, len(themes))
		return true
	}
	switch a.state {
	case scrMenu:
		return a.handleMenu(ev)
	case scrModeSelect:
		return a.handleModeSelect(ev)
	case scrSettings:
		return a.handleSettings(ev)
	case scrKeybinds:
		return a.handleKeybinds(ev)
	case scrScores:
		return a.handleScores(ev)
	case scrPlaying:
		return a.handlePlaying(ev)
	case scrPaused:
		return a.handlePaused(ev)
	case scrGameOver:
		return a.handleGameOver(ev)
	}
	return true
}

func (a *app) handleMenu(ev input.Event) bool {
	switch {
	case isUp(ev):
		a.mainSel = wrap(a.mainSel-1, len(menuItems))
	case isDown(ev):
		a.mainSel = wrap(a.mainSel+1, len(menuItems))
	case isConfirm(ev):
		switch a.mainSel {
		case 0:
			a.state = scrModeSelect
		case 1:
			a.state = scrSettings
		case 2:
			a.state = scrScores
		case 3:
			return false
		}
	}
	return true
}

func (a *app) handleModeSelect(ev input.Event) bool {
	switch {
	case isUp(ev):
		a.modeSel = wrap(a.modeSel-1, len(modes))
	case isDown(ev):
		a.modeSel = wrap(a.modeSel+1, len(modes))
	case isLeft(ev):
		a.startLevel = clampLevel(a.startLevel - 1)
	case isRight(ev):
		a.startLevel = clampLevel(a.startLevel + 1)
	case isConfirm(ev):
		a.startGame(modes[a.modeSel])
	case isBack(ev):
		a.state = scrMenu
	}
	return true
}

func (a *app) handleSettings(ev input.Event) bool {
	switch {
	case isUp(ev):
		a.settingSel = wrap(a.settingSel-1, settingsRows)
	case isDown(ev):
		a.settingSel = wrap(a.settingSel+1, settingsRows)
	case isLeft(ev):
		a.adjustSetting(-1)
	case isRight(ev):
		a.adjustSetting(1)
	case isConfirm(ev):
		switch a.settingSel {
		case setBack:
			a.state = scrMenu
		case setKeybinds:
			a.openKeybinds()
		default:
			a.adjustSetting(1)
		}
	case isBack(ev):
		a.state = scrMenu
	}
	return true
}

func (a *app) openKeybinds() {
	a.bindSel = 0
	a.bindMsg = ""
	a.capturing = false
	a.state = scrKeybinds
}

func (a *app) handleKeybinds(ev input.Event) bool {
	n := len(input.Bindable) + 2
	resetIdx := len(input.Bindable)
	switch {
	case isUp(ev):
		a.bindSel = wrap(a.bindSel-1, n)
		a.bindMsg = ""
	case isDown(ev):
		a.bindSel = wrap(a.bindSel+1, n)
		a.bindMsg = ""
	case isConfirm(ev):
		switch a.bindSel {
		case resetIdx:
			a.keymap = input.DefaultKeymap()
			a.in.SetKeymap(a.keymap)
			a.persist()
			a.bindMsg = "reset to defaults"
		case resetIdx + 1:
			a.state = scrSettings
		default:
			a.captureAction = input.Bindable[a.bindSel]
			a.capturing = true
			a.bindMsg = ""
			a.in.BeginCapture()
		}
	case isBack(ev):
		a.state = scrSettings
	}
	return true
}

func (a *app) onCapture(k input.Key) {
	if !a.capturing {
		return
	}
	a.capturing = false
	a.in.EndCapture()
	switch {
	case k == "esc":
		a.bindMsg = "cancelled"
	case input.Reserved(k):
		a.bindMsg = input.KeyLabel(k) + " is reserved"
	default:
		a.rebind(a.captureAction, k)
		a.in.SetKeymap(a.keymap)
		a.persist()
		a.bindMsg = input.Label(a.captureAction) + " = " + input.KeyLabel(k)
	}
}

func (a *app) rebind(action input.Event, k input.Key) {
	old := a.keymap[action]
	for e, ek := range a.keymap {
		if e != action && ek == k {
			a.keymap[e] = old
		}
	}
	a.keymap[action] = k
}

func (a *app) adjustSetting(d int) {
	switch a.settingSel {
	case setTheme:
		a.themeIdx = wrap(a.themeIdx+d, len(themes))
	case setStartLevel:
		a.startLevel = clampLevel(a.startLevel + d)
	case setColorMode:
		a.colorMode = wrap(a.colorMode+d, len(colorModeNames))
		a.scr.SetColorMode(resolveColorMode(a.colorMode))
	}
}

func (a *app) handleScores(ev input.Event) bool {
	switch {
	case isLeft(ev), isUp(ev):
		a.scoreTab = wrap(a.scoreTab-1, len(modes))
	case isRight(ev), isDown(ev):
		a.scoreTab = wrap(a.scoreTab+1, len(modes))
	case isBack(ev), isConfirm(ev):
		a.state = scrMenu
	}
	return true
}

func (a *app) handlePlaying(ev input.Event) bool {
	switch ev {
	case input.Pause:
		a.pauseSel = 0
		a.state = scrPaused
	case input.ToggleChaos:
		a.sess.ch.Toggle()
		a.sess.chaosToggled = true
	case input.Restart:
		a.restart()
	case input.MoveLeft, input.MoveRight, input.SoftDrop, input.HardDrop,
		input.RotateCW, input.RotateCCW, input.Hold:
		a.applyGame(ev)
		a.checkEnd()
	}
	return true
}

func (a *app) handlePaused(ev input.Event) bool {
	switch {
	case isUp(ev):
		a.pauseSel = wrap(a.pauseSel-1, len(pauseItems))
	case isDown(ev):
		a.pauseSel = wrap(a.pauseSel+1, len(pauseItems))
	case ev == input.Restart:
		a.restart()
	case isBack(ev):
		a.state = scrPlaying
	case isConfirm(ev):
		switch a.pauseSel {
		case 0:
			a.state = scrPlaying
		case 1:
			a.restart()
		case 2:
			a.toMenu()
		case 3:
			return false
		}
	}
	return true
}

func (a *app) handleGameOver(ev input.Event) bool {
	s := a.sess
	if int(s.collapse) < game.VisibleRows {
		if isConfirm(ev) || isBack(ev) {
			s.collapse = float64(game.VisibleRows)
		}
		return true
	}
	switch {
	case isUp(ev):
		a.overSel = wrap(a.overSel-1, len(gameOverItems))
	case isDown(ev):
		a.overSel = wrap(a.overSel+1, len(gameOverItems))
	case ev == input.Restart:
		a.restart()
	case isConfirm(ev):
		switch a.overSel {
		case 0:
			a.restart()
		case 1:
			a.toMenu()
		case 2:
			return false
		}
	case isBack(ev):
		a.toMenu()
	}
	return true
}

func (a *app) applyGame(ev input.Event) {
	s := a.sess
	th := themes[a.themeIdx]
	ox, oy := origin(a.scr.W, a.scr.H)
	switch ev {
	case input.MoveLeft:
		s.g.TryMove(-1, 0)
	case input.MoveRight:
		s.g.TryMove(1, 0)
	case input.SoftDrop:
		s.g.SoftDrop()
	case input.HardDrop:
		landing := s.g.Ghost()
		before := s.g.Level
		res := s.g.HardDrop()
		spawnHardDrop(s.eng, landing, ox, oy, th)
		spawnLockEffects(s.eng, res, ox, oy, th)
		s.ch.OnPiece()
		s.observe(res)
		a.comboFeedback(res, ox, oy, th)
		if s.g.Level > before {
			spawnLevelUp(s.eng, ox, oy, th)
		}
	case input.RotateCW:
		s.g.Rotate(game.CW)
	case input.RotateCCW:
		s.g.Rotate(game.CCW)
	case input.Hold:
		s.g.Hold()
	}
}

func isUp(ev input.Event) bool      { return ev == input.RotateCW }
func isDown(ev input.Event) bool    { return ev == input.SoftDrop }
func isLeft(ev input.Event) bool    { return ev == input.MoveLeft }
func isRight(ev input.Event) bool   { return ev == input.MoveRight }
func isConfirm(ev input.Event) bool { return ev == input.Select || ev == input.HardDrop }
func isBack(ev input.Event) bool    { return ev == input.Pause }

func wrap(i, n int) int {
	if n <= 0 {
		return 0
	}
	return (i%n + n) % n
}

func clampLevel(l int) int {
	if l < 1 {
		return 1
	}
	if l > 15 {
		return 15
	}
	return l
}

var colorModeNames = []string{"Auto", "Truecolor", "256-color", "16-color"}

func clampColorMode(m int) int {
	if m < 0 || m >= len(colorModeNames) {
		return 0
	}
	return m
}

func resolveColorMode(m int) render.ColorMode {
	switch m {
	case 1:
		return render.TrueColor
	case 2:
		return render.Color256
	case 3:
		return render.Color16
	default:
		return render.DetectColorMode()
	}
}

func resolvedColorName() string {
	switch render.DetectColorMode() {
	case render.Color256:
		return "256-color"
	case render.Color16:
		return "16-color"
	default:
		return "Truecolor"
	}
}

func (a *app) colorModeLabel() string {
	m := clampColorMode(a.colorMode)
	if m == 0 {
		return "Auto (" + resolvedColorName() + ")"
	}
	return colorModeNames[m]
}
