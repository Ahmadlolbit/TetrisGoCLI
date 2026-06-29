package chaos

import (
	"testing"

	"awesomeProject/internal/game"
)

func TestGarbageAddsRowsWithGap(t *testing.T) {
	b := game.NewBoard()
	b.AddGarbage(2, 3)
	for y := game.Height - 2; y < game.Height; y++ {
		if b.At(3, y) != game.Empty {
			t.Fatalf("gap column should stay empty at row %d", y)
		}
		filled := 0
		for x := 0; x < game.Width; x++ {
			if b.At(x, y) != game.Empty {
				filled++
			}
		}
		if filled != game.Width-1 {
			t.Fatalf("row %d filled=%d, want %d", y, filled, game.Width-1)
		}
	}
}

func TestFiresWhenEnabled(t *testing.T) {
	g := game.NewGame(1)
	e := New(1, true)
	for i := 0; i < 4000; i++ {
		if k, ok := e.Update(0.1, g); ok {
			if k == None || Name(k) == "" {
				t.Fatalf("fired with invalid kind %d", k)
			}
			return
		}
	}
	t.Fatal("chaos never fired while enabled")
}

func TestDisabledNeverFires(t *testing.T) {
	g := game.NewGame(1)
	e := New(1, false)
	for i := 0; i < 4000; i++ {
		if _, ok := e.Update(0.1, g); ok {
			t.Fatal("disabled chaos should never fire")
		}
	}
}

func TestRemapIdentityWithoutScramble(t *testing.T) {
	e := New(1, true)
	for _, p := range []game.PieceType{game.I, game.O, game.T, game.S, game.Z, game.J, game.L} {
		if e.Remap(p) != p {
			t.Fatalf("Remap should be identity with no scramble active, %d -> %d", p, e.Remap(p))
		}
	}
}

func TestRemapScramblePermutes(t *testing.T) {
	e := New(1, true)
	e.buildRemap()
	e.active = ColorScramble
	seen := map[game.PieceType]bool{}
	for _, p := range []game.PieceType{game.I, game.O, game.T, game.S, game.Z, game.J, game.L} {
		r := e.Remap(p)
		if seen[r] {
			t.Fatalf("scramble remap is not a bijection: %d repeated", r)
		}
		seen[r] = true
	}
}

func TestToggleAndReset(t *testing.T) {
	e := New(1, true)
	e.active = GravitySpike
	e.remaining = 5
	e.meter = 0.5
	e.Toggle()
	if e.Enabled || e.active != None || e.remaining != 0 || e.meter != 0 {
		t.Fatalf("toggle should disable and clear active state: %+v", e)
	}
	e.active = LightsDim
	e.meter = 0.9
	e.Reset()
	if e.active != None || e.meter != 0 {
		t.Fatalf("reset should clear active state and meter: %+v", e)
	}
}

func TestActiveEventBlocksNewAndCooldown(t *testing.T) {
	g := game.NewGame(1)
	e := New(1, true)
	e.active = GravitySpike
	e.remaining = 0.5
	e.meter = 5
	if _, ok := e.Update(0.1, g); ok {
		t.Fatal("no new event should fire while one is active")
	}
	if _, ok := e.Update(0.5, g); ok {
		t.Fatal("expiring an event should not itself fire a new one")
	}
	if e.active != None {
		t.Fatal("event should have expired")
	}
	if _, ok := e.Update(0.1, g); ok {
		t.Fatal("cooldown should block a new event right after one expires")
	}
}

func TestScalesTrackActiveEvent(t *testing.T) {
	g := game.NewGame(1)
	e := New(1, true)
	for i := 0; i < 8000; i++ {
		e.Update(0.1, g)
		switch e.Active() {
		case GravitySpike:
			if e.GravityScale() != 2 {
				t.Fatal("gravity spike should double gravity")
			}
		case BonusFrenzy:
			if e.ScoreScale() != 2 {
				t.Fatal("bonus frenzy should double score")
			}
		}
	}
}
