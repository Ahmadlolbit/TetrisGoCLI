package effects

import (
	"math"
	"math/rand"

	"awesomeProject/internal/render"
)

type particle struct {
	x     float64
	y     float64
	vx    float64
	vy    float64
	life  float64
	max   float64
	color render.Color
	glyph rune
}

type flash struct {
	x     int
	y     int
	w     int
	h     int
	life  float64
	max   float64
	color render.Color
}

type Engine struct {
	particles []particle
	flashes   []flash
	shakeMag  float64
	shakeTime float64
	shakeDur  float64
	rng       *rand.Rand
}

func New(seed int64) *Engine {
	return &Engine{rng: rand.New(rand.NewSource(seed))}
}

func (e *Engine) Clear() {
	e.particles = e.particles[:0]
	e.flashes = e.flashes[:0]
	e.shakeMag = 0
	e.shakeTime = 0
}

func (e *Engine) Active() bool {
	return len(e.particles) > 0 || len(e.flashes) > 0 || e.shakeTime > 0
}

func (e *Engine) Update(dt float64) {
	live := e.particles[:0]
	for _, p := range e.particles {
		p.life -= dt
		if p.life <= 0 {
			continue
		}
		p.x += p.vx * dt
		p.y += p.vy * dt
		p.vy += 22 * dt
		live = append(live, p)
	}
	e.particles = live

	keep := e.flashes[:0]
	for _, f := range e.flashes {
		f.life -= dt
		if f.life <= 0 {
			continue
		}
		keep = append(keep, f)
	}
	e.flashes = keep

	if e.shakeTime > 0 {
		e.shakeTime -= dt
		if e.shakeTime <= 0 {
			e.shakeTime = 0
			e.shakeMag = 0
		}
	}
}

func (e *Engine) Shake(mag, dur float64) {
	if dur >= e.shakeTime {
		e.shakeTime = dur
		e.shakeDur = dur
	}
	if mag > e.shakeMag {
		e.shakeMag = mag
	}
}

func (e *Engine) ShakeOffset() (int, int) {
	if e.shakeTime <= 0 || e.shakeDur <= 0 {
		return 0, 0
	}
	m := e.shakeMag * (e.shakeTime / e.shakeDur)
	dx := (e.rng.Float64()*2 - 1) * m
	dy := (e.rng.Float64()*2 - 1) * m * 0.5
	return int(math.Round(dx)), int(math.Round(dy))
}

func (e *Engine) Flash(x, y, w, h int, col render.Color, dur float64) {
	e.flashes = append(e.flashes, flash{x: x, y: y, w: w, h: h, life: dur, max: dur, color: col})
}

func (e *Engine) Burst(x, y float64, col render.Color, n int, spread float64) {
	for i := 0; i < n; i++ {
		ang := e.rng.Float64() * 2 * math.Pi
		spd := (0.4 + e.rng.Float64()) * spread
		life := 0.3 + e.rng.Float64()*0.5
		e.particles = append(e.particles, particle{
			x:     x,
			y:     y,
			vx:    math.Cos(ang) * spd,
			vy:    math.Sin(ang)*spd - spread*0.4,
			life:  life,
			max:   life,
			color: col,
		})
	}
}

var ramp = []rune{'·', '∘', '•', '*', '▪', '▒'}

func (e *Engine) Apply(b *render.Buffer, offX, offY int) {
	for _, f := range e.flashes {
		t := (f.life / f.max) * 0.7
		for yy := f.y; yy < f.y+f.h; yy++ {
			for xx := f.x; xx < f.x+f.w; xx++ {
				px := xx + offX
				py := yy + offY
				if px < 0 || py < 0 || px >= b.W || py >= b.H {
					continue
				}
				c := b.Cells[py*b.W+px]
				c.BG = render.Lerp(c.BG, f.color, t)
				c.FG = render.Lerp(c.FG, f.color, t)
				b.Cells[py*b.W+px] = c
			}
		}
	}

	for _, p := range e.particles {
		px := int(math.Round(p.x)) + offX
		py := int(math.Round(p.y)) + offY
		if px < 0 || py < 0 || px >= b.W || py >= b.H {
			continue
		}
		t := p.life / p.max
		gi := int(t * float64(len(ramp)-1))
		if gi < 0 {
			gi = 0
		}
		if gi >= len(ramp) {
			gi = len(ramp) - 1
		}
		bg := b.Cells[py*b.W+px].BG
		b.Set(px, py, render.Cell{Ch: ramp[gi], FG: render.Lerp(bg, p.color, t), BG: bg})
	}
}
