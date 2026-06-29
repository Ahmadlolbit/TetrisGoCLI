package main

import (
	"testing"

	"awesomeProject/internal/game"
)

func TestPieceGlyphCoverage(t *testing.T) {
	for _, pt := range []game.PieceType{game.I, game.O, game.T, game.S, game.Z, game.J, game.L, game.Garbage} {
		if pieceGlyph[pt] == 0 {
			t.Errorf("piece type %d has no texture glyph", pt)
		}
	}
}
