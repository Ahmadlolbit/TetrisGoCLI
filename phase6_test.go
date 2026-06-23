package main

import "testing"

func scoresOf(s *scoreboard, k modeKind) []int {
	var out []int
	for _, e := range s.entries(k) {
		out = append(out, e.score)
	}
	return out
}

func TestScoreboardRanksByScoreDescending(t *testing.T) {
	s := newScoreboard()
	m := modeByKind(modeMarathon)
	if r := s.record(m, scoreEntry{score: 100}); r != 0 {
		t.Fatalf("first insert rank = %d, want 0", r)
	}
	if r := s.record(m, scoreEntry{score: 300}); r != 0 {
		t.Fatalf("top insert rank = %d, want 0", r)
	}
	if r := s.record(m, scoreEntry{score: 200}); r != 1 {
		t.Fatalf("middle insert rank = %d, want 1", r)
	}
	got := scoresOf(s, modeMarathon)
	want := []int{300, 200, 100}
	if len(got) != len(want) {
		t.Fatalf("length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestScoreboardCapsAndRejectsWorst(t *testing.T) {
	s := newScoreboard()
	m := modeByKind(modeMarathon)
	for i := 1; i <= maxScoresPerMode; i++ {
		s.record(m, scoreEntry{score: i * 10})
	}
	if got := len(s.entries(modeMarathon)); got != maxScoresPerMode {
		t.Fatalf("filled length = %d, want %d", got, maxScoresPerMode)
	}
	if r := s.record(m, scoreEntry{score: 5}); r != -1 {
		t.Fatalf("worst-when-full rank = %d, want -1", r)
	}
	if got := len(s.entries(modeMarathon)); got != maxScoresPerMode {
		t.Fatalf("post-reject length = %d, want %d", got, maxScoresPerMode)
	}
}

func TestScoreboardRanksByTimeAscending(t *testing.T) {
	s := newScoreboard()
	m := modeByKind(modeSprint)
	if !m.timed() {
		t.Fatal("sprint mode should be timed")
	}
	s.record(m, scoreEntry{time: 10})
	if r := s.record(m, scoreEntry{time: 5}); r != 0 {
		t.Fatalf("faster time rank = %d, want 0", r)
	}
	if r := s.record(m, scoreEntry{time: 8}); r != 1 {
		t.Fatalf("middle time rank = %d, want 1", r)
	}
	got := s.entries(modeSprint)
	want := []float64{5, 8, 10}
	for i := range want {
		if got[i].time != want[i] {
			t.Fatalf("time order[%d] = %v, want %v", i, got[i].time, want[i])
		}
	}
}

func TestFormatTime(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0:00.00"},
		{83.45, "1:23.45"},
		{5.5, "0:05.50"},
	}
	for _, c := range cases {
		if got := formatTime(c.in); got != c.want {
			t.Fatalf("formatTime(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestModeConfig(t *testing.T) {
	if modeByKind(modeSprint).sprintLines != 40 {
		t.Fatal("sprint should target 40 lines")
	}
	if modeByKind(modeMarathon).timed() {
		t.Fatal("marathon should not be timed")
	}
	if !modeByKind(modeChaos).chaosEnabled {
		t.Fatal("chaos mode should enable chaos")
	}
	if modeByKind(modeClassic).gravityScale >= 1 {
		t.Fatal("classic should use gentler gravity")
	}
}
