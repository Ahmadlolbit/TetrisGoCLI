package game

const (
	CW  = 1
	CCW = -1

	spawnX       = 3
	spawnY       = 2
	previewCount = 5

	lockDelaySeconds = 0.5
	maxLockResets    = 15
)

type TSpinType int

const (
	TSpinNone TSpinType = iota
	TSpinMini
	TSpinFull
)

type LockResult struct {
	Lines      int
	TSpin      TSpinType
	Score      int
	Combo      int
	BackToBack bool
	Difficult  bool
}

var gravityTable = map[int]float64{
	1: 1.5, 2: 2.2, 3: 3.0, 4: 4.0, 5: 5.5,
	6: 7.5, 7: 10, 8: 14, 9: 18, 10: 24,
	11: 30, 12: 36, 13: 42, 14: 48, 15: 55,
	16: 62, 17: 70, 18: 80, 19: 90, 20: 100,
}

type Game struct {
	Board      *Board
	Current    Piece
	HoldType   PieceType
	NextQueue  []PieceType
	Score      int
	Level      int
	Lines      int
	Combo      int
	BackToBack bool
	Over       bool

	bag             *Bag
	holdUsed        bool
	startLevel      int
	lastWasRotation bool
	lastKickIndex   int
	gravityAccum    float64
	lockTimer       float64
	lockResets      int
}

func NewGame(seed int64) *Game {
	return NewGameAtLevel(seed, 1)
}

func NewGameAtLevel(seed int64, level int) *Game {
	if level < 1 {
		level = 1
	}
	g := &Game{
		Board:      NewBoard(),
		bag:        NewBag(seed),
		HoldType:   Empty,
		Level:      level,
		startLevel: level,
	}
	g.refillQueue()
	g.spawn()
	return g
}

func (g *Game) refillQueue() {
	for len(g.NextQueue) < previewCount {
		g.NextQueue = append(g.NextQueue, g.bag.Next())
	}
}

func (g *Game) spawn() {
	g.refillQueue()
	t := g.NextQueue[0]
	g.NextQueue = g.NextQueue[1:]
	g.refillQueue()
	g.spawnPiece(t)
}

func (g *Game) spawnPiece(t PieceType) {
	g.Current = Piece{Type: t, Rotation: 0, X: spawnX, Y: spawnY}
	g.lastWasRotation = false
	g.lockTimer = 0
	g.lockResets = 0
	g.gravityAccum = 0
	if g.Board.Collides(g.Current) {
		g.Over = true
	}
}

func (g *Game) grounded() bool {
	test := g.Current
	test.Y++
	return g.Board.Collides(test)
}

func (g *Game) resetLock() {
	if g.lockResets < maxLockResets && g.grounded() {
		g.lockTimer = 0
		g.lockResets++
	}
}

func (g *Game) TryMove(dx, dy int) bool {
	test := g.Current
	test.X += dx
	test.Y += dy
	if g.Board.Collides(test) {
		return false
	}
	g.Current = test
	g.lastWasRotation = false
	g.resetLock()
	return true
}

func (g *Game) Rotate(dir int) bool {
	if g.Current.Type == O {
		return false
	}
	from := g.Current.Rotation
	to := (from + dir + 4) % 4
	for i, k := range kicksFor(g.Current.Type, from, to) {
		test := g.Current
		test.Rotation = to
		test.X += k.X
		test.Y += k.Y
		if !g.Board.Collides(test) {
			g.Current = test
			g.lastWasRotation = true
			g.lastKickIndex = i
			g.resetLock()
			return true
		}
	}
	return false
}

func (g *Game) SoftDrop() bool {
	if g.TryMove(0, 1) {
		g.Score++
		return true
	}
	return false
}

func (g *Game) HardDrop() LockResult {
	dist := 0
	for !g.grounded() {
		g.Current.Y++
		dist++
	}
	g.Score += 2 * dist
	return g.lockPiece()
}

func (g *Game) Hold() bool {
	if g.holdUsed {
		return false
	}
	g.holdUsed = true
	stored := g.HoldType
	g.HoldType = g.Current.Type
	if stored == Empty {
		g.spawn()
	} else {
		g.spawnPiece(stored)
	}
	return true
}

func (g *Game) Tick(dt float64) []LockResult {
	if g.Over {
		return nil
	}
	g.gravityAccum += dt * g.gravitySpeed()
	for g.gravityAccum >= 1 {
		g.gravityAccum -= 1
		g.stepDown()
	}
	var out []LockResult
	if g.grounded() {
		g.lockTimer += dt
		if g.lockTimer >= lockDelaySeconds {
			out = append(out, g.lockPiece())
		}
	} else {
		g.lockTimer = 0
	}
	return out
}

func (g *Game) stepDown() bool {
	if g.grounded() {
		return false
	}
	g.Current.Y++
	g.lastWasRotation = false
	return true
}

func (g *Game) gravitySpeed() float64 {
	l := g.Level
	if l < 1 {
		l = 1
	}
	if l > 20 {
		l = 20
	}
	return gravityTable[l]
}

func (g *Game) lockPiece() LockResult {
	tspin := g.detectTSpin()
	g.Board.LockPiece(g.Current)
	cleared := g.Board.ClearLines()
	res := g.scoreClear(cleared, tspin)
	g.holdUsed = false
	g.spawn()
	return res
}

func (g *Game) detectTSpin() TSpinType {
	if g.Current.Type != T || !g.lastWasRotation {
		return TSpinNone
	}
	cx := g.Current.X + 1
	cy := g.Current.Y + 1
	corners := [4]Point{
		{cx - 1, cy - 1},
		{cx + 1, cy - 1},
		{cx - 1, cy + 1},
		{cx + 1, cy + 1},
	}
	var occ [4]bool
	count := 0
	for i, c := range corners {
		if g.Board.Occupied(c.X, c.Y) {
			occ[i] = true
			count++
		}
	}
	if count < 3 {
		return TSpinNone
	}
	front := frontCorners(g.Current.Rotation)
	frontCount := 0
	for _, idx := range front {
		if occ[idx] {
			frontCount++
		}
	}
	if frontCount == 2 || g.lastKickIndex == 4 {
		return TSpinFull
	}
	return TSpinMini
}

func frontCorners(rot int) [2]int {
	switch rot {
	case 0:
		return [2]int{0, 1}
	case 1:
		return [2]int{1, 3}
	case 2:
		return [2]int{2, 3}
	default:
		return [2]int{0, 2}
	}
}

func (g *Game) scoreClear(cleared int, tspin TSpinType) LockResult {
	base := 0
	difficult := false
	switch tspin {
	case TSpinFull:
		switch cleared {
		case 0:
			base = 400
		case 1:
			base, difficult = 800, true
		case 2:
			base, difficult = 1200, true
		case 3:
			base, difficult = 1600, true
		}
	case TSpinMini:
		switch cleared {
		case 0:
			base = 100
		case 1:
			base, difficult = 200, true
		case 2:
			base, difficult = 400, true
		}
	default:
		switch cleared {
		case 1:
			base = 100
		case 2:
			base = 300
		case 3:
			base = 500
		case 4:
			base, difficult = 800, true
		}
	}

	level := g.Level
	gain := base * level
	if difficult && g.BackToBack {
		gain += gain / 2
	}
	if cleared > 0 {
		g.Combo++
		if g.Combo > 1 {
			gain += 50 * (g.Combo - 1) * level
		}
		g.BackToBack = difficult
	} else {
		g.Combo = 0
	}

	g.Score += gain
	g.Lines += cleared
	g.Level = g.startLevel + g.Lines/10

	return LockResult{
		Lines:      cleared,
		TSpin:      tspin,
		Score:      gain,
		Combo:      g.Combo,
		BackToBack: g.BackToBack,
		Difficult:  difficult,
	}
}

func (g *Game) Ghost() Piece {
	p := g.Current
	for {
		test := p
		test.Y++
		if g.Board.Collides(test) {
			return p
		}
		p = test
	}
}
