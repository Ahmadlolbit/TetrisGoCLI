package main

import (
	"fmt"

	"awesomeProject/internal/store"
)

const maxScoresPerMode = 8

type scoreEntry struct {
	score int
	level int
	lines int
	combo int
	time  float64
}

type scoreboard struct {
	tables map[modeKind][]scoreEntry
}

func newScoreboard() *scoreboard {
	return &scoreboard{tables: make(map[modeKind][]scoreEntry)}
}

func (s *scoreboard) entries(k modeKind) []scoreEntry {
	return s.tables[k]
}

func (s *scoreboard) record(m mode, e scoreEntry) int {
	list := s.tables[m.kind]
	pos := len(list)
	for i, x := range list {
		if e.beats(x, m.timed()) {
			pos = i
			break
		}
	}
	list = append(list, scoreEntry{})
	copy(list[pos+1:], list[pos:])
	list[pos] = e
	if len(list) > maxScoresPerMode {
		list = list[:maxScoresPerMode]
	}
	s.tables[m.kind] = list
	if pos >= len(list) {
		return -1
	}
	return pos
}

func (s *scoreboard) export() map[string][]store.Entry {
	out := make(map[string][]store.Entry)
	for _, m := range modes {
		list := s.tables[m.kind]
		if len(list) == 0 {
			continue
		}
		es := make([]store.Entry, len(list))
		for i, e := range list {
			es[i] = store.Entry{Score: e.score, Level: e.level, Lines: e.lines, Combo: e.combo, Time: e.time}
		}
		out[m.slug] = es
	}
	return out
}

func (s *scoreboard) load(scores map[string][]store.Entry) {
	for _, m := range modes {
		for _, e := range scores[m.slug] {
			s.record(m, scoreEntry{score: e.Score, level: e.Level, lines: e.Lines, combo: e.Combo, time: e.Time})
		}
	}
}

func (e scoreEntry) beats(o scoreEntry, byTime bool) bool {
	if byTime {
		return e.time < o.time
	}
	return e.score > o.score
}

func comboChain(c int) int {
	if c > 1 {
		return c - 1
	}
	return 0
}

func formatTime(t float64) string {
	cs := int(t * 100)
	if cs < 0 {
		cs = 0
	}
	return fmt.Sprintf("%d:%02d.%02d", cs/6000, (cs/100)%60, cs%100)
}
