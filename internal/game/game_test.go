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
