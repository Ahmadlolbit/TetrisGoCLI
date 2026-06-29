package game

import "testing"

func TestBagPermutation(t *testing.T) {
	b := NewBag(42)
	for round := 0; round < 200; round++ {
		seen := map[PieceType]int{}
		for i := 0; i < 7; i++ {
			seen[b.Next()]++
		}
		if len(seen) != 7 {
			t.Fatalf("round %d not a permutation: %v", round, seen)
		}
		for p, c := range seen {
			if c != 1 {
				t.Fatalf("round %d piece %d appeared %d times", round, p, c)
			}
		}
	}
}

func TestBagDeterministic(t *testing.T) {
	a := NewBag(7)
	b := NewBag(7)
	for i := 0; i < 100; i++ {
		if a.Next() != b.Next() {
			t.Fatalf("same seed diverged at draw %d", i)
		}
	}
}

func TestClearSingleLine(t *testing.T) {
	b := NewBoard()
	row := Height - 1
	for x := 0; x < Width; x++ {
		b.Set(x, row, L)
	}
	b.Set(0, row-1, I)
	if n := b.ClearLines(); n != 1 {
		t.Fatalf("expected 1 cleared, got %d", n)
	}
	if b.At(0, row) != I {
		t.Fatalf("block above did not fall: got %d", b.At(0, row))
	}
}

func TestClearTetris(t *testing.T) {
	b := NewBoard()
	for y := Height - 4; y < Height; y++ {
		for x := 1; x < Width; x++ {
			b.Set(x, y, S)
		}
	}
	if n := b.ClearLines(); n != 0 {
		t.Fatalf("expected 0 with empty column, got %d", n)
	}
	for y := Height - 4; y < Height; y++ {
		b.Set(0, y, S)
	}
	if n := b.ClearLines(); n != 4 {
		t.Fatalf("expected 4 cleared, got %d", n)
	}
}

func TestRotateOpenSpace(t *testing.T) {
	g := NewGame(1)
	g.Current = Piece{Type: T, Rotation: 0, X: 4, Y: 10}
	if !g.Rotate(CW) {
		t.Fatal("rotation in open space should succeed")
	}
	if g.Current.Rotation != 1 {
		t.Fatalf("rotation = %d, want 1", g.Current.Rotation)
	}
}

func TestWallKick(t *testing.T) {
	g := NewGame(1)
	g.Current = Piece{Type: J, Rotation: 1, X: -1, Y: 10}
	if g.Board.Collides(g.Current) {
		t.Fatal("precondition failed: start position should be valid")
	}
	if !g.Rotate(CW) {
		t.Fatal("expected wall kick to succeed")
	}
	if g.Current.Rotation != 2 {
		t.Fatalf("rotation = %d, want 2", g.Current.Rotation)
	}
	if g.Current.X != 0 {
		t.Fatalf("expected kick to shift X to 0, got %d", g.Current.X)
	}
}

func TestTSpinDetection(t *testing.T) {
	g := NewGame(1)
	g.Current = Piece{Type: T, Rotation: 2, X: 3, Y: 20}
	g.Board.Set(3, 20, S)
	g.Board.Set(5, 20, S)
	g.Board.Set(3, 22, S)
	g.lastWasRotation = true
	g.lastKickIndex = 0
	if got := g.detectTSpin(); got != TSpinMini {
		t.Fatalf("expected mini t-spin, got %d", got)
	}
	g.Board.Set(5, 22, S)
	if got := g.detectTSpin(); got != TSpinFull {
		t.Fatalf("expected full t-spin, got %d", got)
	}
	g.lastWasRotation = false
	if got := g.detectTSpin(); got != TSpinNone {
		t.Fatalf("expected none without rotation, got %d", got)
	}
}

func TestTetrisScoringAndBackToBack(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	fillForTetris := func() {
		for y := Height - 4; y < Height; y++ {
			for x := 1; x < Width; x++ {
				g.Board.Set(x, y, S)
			}
		}
	}
	fillForTetris()
	g.Current = Piece{Type: I, Rotation: 1, X: -2, Y: Height - 4}
	res := g.HardDrop()
	if res.Lines != 4 {
		t.Fatalf("expected 4 lines, got %d", res.Lines)
	}
	if !res.Difficult {
		t.Fatal("tetris should be a difficult clear")
	}
	if g.Lines != 4 {
		t.Fatalf("total lines = %d, want 4", g.Lines)
	}
	if res.Score != 800 {
		t.Fatalf("first tetris should score 800 with no bonus, got %d", res.Score)
	}

	fillForTetris()
	g.Current = Piece{Type: I, Rotation: 1, X: -2, Y: Height - 4}
	res2 := g.HardDrop()
	if !res2.BackToBack {
		t.Fatal("second consecutive tetris should be back-to-back")
	}
	if res2.Score <= 800 {
		t.Fatalf("back-to-back tetris should beat a plain tetris, got %d", res2.Score)
	}
}

func TestBonusFrenzyScalesSoftDrop(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	g.ScoreScale = 2
	if !g.SoftDrop() {
		t.Fatal("soft drop on empty board should succeed")
	}
	if g.Score != 2 {
		t.Fatalf("frenzy soft drop scored %d, want 2", g.Score)
	}
}

func TestBonusFrenzyDoublesHardDrop(t *testing.T) {
	plain := NewGameAtLevel(5, 1)
	plain.HardDrop()
	frenzy := NewGameAtLevel(5, 1)
	frenzy.ScoreScale = 2
	frenzy.HardDrop()
	if plain.Score == 0 {
		t.Fatal("hard drop should award distance points")
	}
	if frenzy.Score != plain.Score*2 {
		t.Fatalf("frenzy hard drop = %d, want double of %d", frenzy.Score, plain.Score)
	}
}

func TestAddGarbageLiftsPieceAndNeverTopsOut(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	for y := VisibleTop + 1; y < Height; y++ {
		for x := 0; x < Width; x++ {
			g.Board.Set(x, y, Garbage)
		}
	}
	g.Current = Piece{Type: I, Rotation: 0, X: spawnX, Y: spawnY}
	beforeY := g.Current.Y
	g.AddGarbage(2, 3)
	if g.Over {
		t.Fatal("garbage surge must never top out the player")
	}
	if g.Current.Y != beforeY-1 {
		t.Fatalf("piece should lift by the rows added (1), got %d want %d", g.Current.Y, beforeY-1)
	}
	if top := g.stackTop(); top < VisibleTop {
		t.Fatalf("garbage pushed the stack into the hidden buffer, top=%d", top)
	}
}

func TestAddGarbageSkippedWhenNoHeadroom(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	for y := VisibleTop; y < Height; y++ {
		for x := 0; x < Width; x++ {
			g.Board.Set(x, y, Garbage)
		}
	}
	g.Current = Piece{Type: I, Rotation: 0, X: spawnX, Y: spawnY}
	beforeY := g.Current.Y
	g.AddGarbage(2, 3)
	if g.Current.Y != beforeY {
		t.Fatalf("no garbage should be added without headroom; Y moved %d->%d", beforeY, g.Current.Y)
	}
	if g.Over {
		t.Fatal("must not top out")
	}
}

func TestComboCounter(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	clearOne := func() {
		g.Board = NewBoard()
		for x := 1; x < Width; x++ {
			g.Board.Set(x, Height-1, S)
		}
		g.Current = Piece{Type: I, Rotation: 1, X: -2, Y: Height - 4}
		g.HardDrop()
	}
	clearOne()
	if g.Combo != 1 {
		t.Fatalf("combo after first clear = %d, want 1", g.Combo)
	}
	clearOne()
	if g.Combo != 2 {
		t.Fatalf("combo after second clear = %d, want 2", g.Combo)
	}
}

func TestTSpinScoring(t *testing.T) {
	cases := []struct {
		tspin   TSpinType
		cleared int
		level   int
		base    int
		diff    bool
	}{
		{TSpinFull, 1, 1, 800, true},
		{TSpinFull, 2, 1, 1200, true},
		{TSpinFull, 3, 1, 1600, true},
		{TSpinFull, 0, 3, 400, false},
		{TSpinMini, 1, 1, 200, true},
		{TSpinMini, 0, 2, 100, false},
	}
	for _, tc := range cases {
		g := NewGameAtLevel(1, tc.level)
		res := g.scoreClear(tc.cleared, tc.tspin)
		if want := tc.base * tc.level; res.Score != want {
			t.Errorf("tspin=%d cleared=%d level=%d: score=%d want=%d", tc.tspin, tc.cleared, tc.level, res.Score, want)
		}
		if res.Difficult != tc.diff {
			t.Errorf("tspin=%d cleared=%d: difficult=%v want=%v", tc.tspin, tc.cleared, res.Difficult, tc.diff)
		}
	}
}

func TestBackToBackBreakAndComboReset(t *testing.T) {
	g := NewGameAtLevel(1, 5)
	r1 := g.scoreClear(4, TSpinNone)
	if !r1.BackToBack || !r1.Difficult {
		t.Fatalf("tetris should arm back-to-back: %+v", r1)
	}
	r2 := g.scoreClear(1, TSpinNone)
	if r2.BackToBack {
		t.Fatal("a single after a tetris should break back-to-back")
	}
	if want := 100*5 + 50*1*5; r2.Score != want {
		t.Fatalf("single after tetris score=%d want=%d (no b2b bonus, +combo)", r2.Score, want)
	}
	r3 := g.scoreClear(0, TSpinNone)
	if r3.Combo != 0 {
		t.Fatalf("a no-line lock should reset combo, got %d", r3.Combo)
	}
}

func TestHoldMechanic(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	first := g.Current.Type
	if !g.Hold() {
		t.Fatal("first hold should succeed")
	}
	if g.HoldType != first {
		t.Fatalf("hold should store the current piece, got %d want %d", g.HoldType, first)
	}
	if g.Hold() {
		t.Fatal("a second hold before locking must fail")
	}
	g.HardDrop()
	if !g.Hold() {
		t.Fatal("hold should be allowed again after a lock")
	}
	if g.Current.Type != first {
		t.Fatalf("hold swap should bring back the stored piece, got %d want %d", g.Current.Type, first)
	}
}

func TestTopOut(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	for y := 0; y <= VisibleTop+2; y++ {
		for x := 0; x < Width-1; x++ {
			g.Board.Set(x, y, Garbage)
		}
	}
	g.spawnPiece(O)
	if !g.Over {
		t.Fatal("spawning into a filled board should set Over")
	}
	if got := g.Tick(0.016); got != nil {
		t.Fatalf("Tick after Over must be a no-op, got %v", got)
	}
}

func TestGravityStepsDown(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	startY := g.Current.Y
	g.Tick(1.0)
	if g.Current.Y != startY+1 {
		t.Fatalf("level 1 over 1s (1.5 cells) should step once, Y %d->%d", startY, g.Current.Y)
	}
}

func TestGravityScaleDoublesDescent(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	g.GravityScale = 2
	startY := g.Current.Y
	g.Tick(1.0)
	if g.Current.Y != startY+3 {
		t.Fatalf("gravity scale 2 at level 1 over 1s (3 cells) should step 3, Y %d->%d", startY, g.Current.Y)
	}
}

func TestLockResetCap(t *testing.T) {
	g := NewGameAtLevel(1, 1)
	for !g.grounded() {
		g.Current.Y++
	}
	for i := 0; i < maxLockResets; i++ {
		g.lockTimer = lockDelaySeconds * 0.5
		g.resetLock()
		if g.lockTimer != 0 {
			t.Fatalf("reset %d should zero the lock timer", i)
		}
	}
	g.lockTimer = lockDelaySeconds * 0.5
	g.resetLock()
	if g.lockTimer == 0 {
		t.Fatalf("after %d resets the cap must prevent further reset", maxLockResets)
	}
}

func TestKicksForDispatch(t *testing.T) {
	got := kicksFor(I, 0, 1)
	want := iKicks[kickKey{0, 1}]
	if len(got) != len(want) {
		t.Fatalf("I kicks 0->1: len %d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("I kick %d = %v want %v", i, got[i], want[i])
		}
	}
	if kicksFor(T, 0, 1)[1] == kicksFor(I, 0, 1)[1] {
		t.Fatal("I and JLSTZ kick tables should differ")
	}
	if k := kicksFor(O, 0, 1); len(k) != 1 || k[0] != (Point{0, 0}) {
		t.Fatalf("O should have no real kicks, got %v", k)
	}
}

func TestIPieceWallKick(t *testing.T) {
	g := NewGame(1)
	g.Current = Piece{Type: I, Rotation: 1, X: 7, Y: 10}
	if g.Board.Collides(g.Current) {
		t.Fatal("precondition: vertical I at X=7 should be valid")
	}
	if !g.Rotate(CCW) {
		t.Fatal("I should wall-kick when rotating off the right wall")
	}
	if g.Current.Rotation != 0 {
		t.Fatalf("rotation = %d, want 0", g.Current.Rotation)
	}
	if g.Current.X != 6 {
		t.Fatalf("expected kick to shift X to 6, got %d", g.Current.X)
	}
}
