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
