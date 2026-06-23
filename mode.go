package main

type modeKind int

const (
	modeMarathon modeKind = iota
	modeSprint
	modeChaos
	modeClassic
)

type mode struct {
	kind         modeKind
	name         string
	slug         string
	tagline      string
	chaosEnabled bool
	chaosFreq    float64
	gravityScale float64
	sprintLines  int
}

var modes = []mode{
	{modeMarathon, "Marathon", "marathon", "Endless climb through the levels", false, 1, 1, 0},
	{modeSprint, "Sprint", "sprint", "Clear 40 lines, race the clock", false, 1, 1, 40},
	{modeChaos, "Chaos", "chaos", "Signature mode, events in overdrive", true, 1.7, 1, 0},
	{modeClassic, "Classic", "classic", "Chaos off, gentler gravity", false, 1, 0.7, 0},
}

func (m mode) timed() bool {
	return m.sprintLines > 0
}

func modeByKind(k modeKind) mode {
	for _, m := range modes {
		if m.kind == k {
			return m
		}
	}
	return modes[0]
}
