package game

const (
	Width       = 10
	Height      = 24
	HiddenRows  = 4
	VisibleTop  = HiddenRows
	VisibleRows = Height - HiddenRows
)

type Board struct {
	cells [Height][Width]PieceType
}

func NewBoard() *Board {
	return &Board{}
}

func (b *Board) At(x, y int) PieceType {
	return b.cells[y][x]
}

func (b *Board) Set(x, y int, t PieceType) {
	b.cells[y][x] = t
}

func (b *Board) Occupied(x, y int) bool {
	if x < 0 || x >= Width || y >= Height {
		return true
	}
	if y < 0 {
		return false
	}
	return b.cells[y][x] != Empty
}

func (b *Board) Collides(p Piece) bool {
	for _, c := range p.Cells() {
		if b.Occupied(c.X, c.Y) {
			return true
		}
	}
	return false
}

func (b *Board) LockPiece(p Piece) {
	for _, c := range p.Cells() {
		if c.X >= 0 && c.X < Width && c.Y >= 0 && c.Y < Height {
			b.cells[c.Y][c.X] = p.Type
		}
	}
}

func (b *Board) FullRows() []int {
	var rows []int
	for y := 0; y < Height; y++ {
		full := true
		for x := 0; x < Width; x++ {
			if b.cells[y][x] == Empty {
				full = false
				break
			}
		}
		if full {
			rows = append(rows, y)
		}
	}
	return rows
}

func (b *Board) ClearLines() int {
	cleared := 0
	for y := Height - 1; y >= 0; y-- {
		full := true
		for x := 0; x < Width; x++ {
			if b.cells[y][x] == Empty {
				full = false
				break
			}
		}
		if full {
			cleared++
			for yy := y; yy > 0; yy-- {
				b.cells[yy] = b.cells[yy-1]
			}
			b.cells[0] = [Width]PieceType{}
			y++
		}
	}
	return cleared
}
