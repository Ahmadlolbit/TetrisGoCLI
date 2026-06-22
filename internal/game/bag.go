package game

import "math/rand"

type Bag struct {
	rng   *rand.Rand
	queue []PieceType
}

func NewBag(seed int64) *Bag {
	b := &Bag{rng: rand.New(rand.NewSource(seed))}
	b.refill()
	return b
}

func (b *Bag) refill() {
	pieces := []PieceType{I, O, T, S, Z, J, L}
	b.rng.Shuffle(len(pieces), func(i, j int) {
		pieces[i], pieces[j] = pieces[j], pieces[i]
	})
	b.queue = append(b.queue, pieces...)
}

func (b *Bag) Next() PieceType {
	if len(b.queue) == 0 {
		b.refill()
	}
	p := b.queue[0]
	b.queue = b.queue[1:]
	return p
}
