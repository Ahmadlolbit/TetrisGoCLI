package main

import "testing"

func TestScoreboardExportLoadRoundTrip(t *testing.T) {
	s := newScoreboard()
	s.record(modeByKind(modeMarathon), scoreEntry{score: 300, level: 4, lines: 30, combo: 2})
	s.record(modeByKind(modeMarathon), scoreEntry{score: 100, level: 1, lines: 5})
	s.record(modeByKind(modeSprint), scoreEntry{time: 5, level: 2, lines: 40})

	exp := s.export()
	if _, ok := exp["marathon"]; !ok {
		t.Fatalf("export should key by stable slug, got keys %v", exp)
	}
	if _, ok := exp["Marathon"]; ok {
		t.Fatal("export must not key by display name")
	}

	loaded := newScoreboard()
	loaded.load(exp)

	got := scoresOf(loaded, modeMarathon)
	if len(got) != 2 || got[0] != 300 || got[1] != 100 {
		t.Fatalf("marathon order after round-trip = %v, want [300 100]", got)
	}
	sprint := loaded.entries(modeSprint)
	if len(sprint) != 1 || sprint[0].time != 5 {
		t.Fatalf("sprint not round-tripped: %+v", sprint)
	}
	if marathon := loaded.entries(modeMarathon); marathon[0].combo != 2 {
		t.Fatalf("combo not preserved: %+v", marathon[0])
	}
}
