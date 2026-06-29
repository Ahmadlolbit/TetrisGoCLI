package chaos

import (
	"math/rand"

	"awesomeProject/internal/game"
)

type Kind int

const (
	None Kind = iota
	GarbageSurge
	GravitySpike
	ColorScramble
	LightsDim
	BonusFrenzy
)

type Engine struct {
	Enabled   bool
	FreqScale float64
	meter     float64
	pieces    int
	active    Kind
	lastFired Kind
	remaining float64
	cooldown  float64
	announce  float64
	remap     map[game.PieceType]game.PieceType
	rng       *rand.Rand
}

func New(seed int64, enabled bool) *Engine {
	return &Engine{Enabled: enabled, FreqScale: 1, rng: rand.New(rand.NewSource(seed))}
}

func (e *Engine) Toggle() {
	e.Enabled = !e.Enabled
	e.active = None
	e.remaining = 0
	e.meter = 0
}

func (e *Engine) Reset() {
	e.meter = 0
	e.pieces = 0
	e.active = None
	e.remaining = 0
	e.cooldown = 0
	e.announce = 0
}

func (e *Engine) OnPiece() {
	e.pieces++
}

func (e *Engine) Active() Kind {
	return e.active
}

func (e *Engine) Meter() float64 {
	if e.meter > 1 {
		return 1
	}
	return e.meter
}

func (e *Engine) GravityScale() float64 {
	if e.active == GravitySpike {
		return 2
	}
	return 1
}

func (e *Engine) ScoreScale() float64 {
	if e.active == BonusFrenzy {
		return 2
	}
	return 1
}

func (e *Engine) Status() (string, bool) {
	if e.announce > 0 {
		return ShortName(e.lastFired), true
	}
	if e.active != None {
		return ShortName(e.active), false
	}
	return "", false
}

func (e *Engine) Remap(t game.PieceType) game.PieceType {
	if e.active != ColorScramble || e.remap == nil {
		return t
	}
	if r, ok := e.remap[t]; ok {
		return r
	}
	return t
}

func (e *Engine) Update(dt float64, g *game.Game) (Kind, bool) {
	if e.announce > 0 {
		e.announce -= dt
		if e.announce < 0 {
			e.announce = 0
		}
	}
	if !e.Enabled {
		return None, false
	}
	if e.active != None {
		e.remaining -= dt
		if e.remaining <= 0 {
			e.active = None
			e.cooldown = 4
		}
		return None, false
	}
	if e.cooldown > 0 {
		e.cooldown -= dt
		return None, false
	}

	freq := e.FreqScale
	if freq <= 0 {
		freq = 1
	}
	e.meter += (dt*(0.04+0.005*float64(g.Level)) + float64(e.pieces)*0.03) * freq
	e.pieces = 0
	if e.meter < 1 {
		return None, false
	}
	e.meter = 0

	k := e.pick()
	e.lastFired = k
	e.announce = 1.5

	switch k {
	case GarbageSurge:
		rows := 1 + e.rng.Intn(2)
		gap := e.rng.Intn(game.Width)
		g.AddGarbage(rows, gap)
		e.cooldown = 5
	case ColorScramble:
		e.buildRemap()
		e.active = k
		e.remaining = 7
	case GravitySpike:
		e.active = k
		e.remaining = 6
	case LightsDim:
		e.active = k
		e.remaining = 5
	case BonusFrenzy:
		e.active = k
		e.remaining = 8
	}
	return k, true
}

func (e *Engine) buildRemap() {
	list := []game.PieceType{game.I, game.O, game.T, game.S, game.Z, game.J, game.L}
	perm := make([]game.PieceType, len(list))
	copy(perm, list)
	e.rng.Shuffle(len(perm), func(i, j int) {
		perm[i], perm[j] = perm[j], perm[i]
	})
	e.remap = make(map[game.PieceType]game.PieceType, len(list))
	for i, t := range list {
		e.remap[t] = perm[i]
	}
}

type weighted struct {
	kind   Kind
	weight int
}

var catalog = []weighted{
	{GarbageSurge, 3},
	{GravitySpike, 3},
	{ColorScramble, 2},
	{LightsDim, 2},
	{BonusFrenzy, 3},
}

func (e *Engine) pick() Kind {
	total := 0
	for _, c := range catalog {
		total += c.weight
	}
	r := e.rng.Intn(total)
	for _, c := range catalog {
		if r < c.weight {
			return c.kind
		}
		r -= c.weight
	}
	return GravitySpike
}

func Name(k Kind) string {
	switch k {
	case GarbageSurge:
		return "GARBAGE SURGE"
	case GravitySpike:
		return "GRAVITY SPIKE"
	case ColorScramble:
		return "COLOR SCRAMBLE"
	case LightsDim:
		return "LIGHTS DIM"
	case BonusFrenzy:
		return "BONUS FRENZY"
	}
	return ""
}

func ShortName(k Kind) string {
	switch k {
	case GarbageSurge:
		return "GARBAGE"
	case GravitySpike:
		return "GRAVITY"
	case ColorScramble:
		return "SCRAMBLE"
	case LightsDim:
		return "DIM"
	case BonusFrenzy:
		return "FRENZY"
	}
	return ""
}
