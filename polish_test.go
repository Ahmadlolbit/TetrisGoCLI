package main

import "testing"

func TestChaosMasterSwitch(t *testing.T) {
	on := newSession(modeByKind(modeChaos), 1, 1, true)
	if !on.ch.Enabled {
		t.Fatal("chaos mode with the master switch on should enable chaos")
	}
	off := newSession(modeByKind(modeChaos), 1, 1, false)
	if off.ch.Enabled {
		t.Fatal("the master switch off should disable chaos even in chaos mode")
	}
	if !off.chaosToggled {
		t.Fatal("chaos mode with chaos disabled should be tainted (kept off the leaderboard)")
	}
	if on.chaosToggled {
		t.Fatal("a normal chaos run should not be tainted")
	}
	cl := newSession(modeByKind(modeClassic), 1, 1, true)
	if cl.ch.Enabled {
		t.Fatal("classic mode should never enable chaos")
	}
}

func TestChaosSettingMapping(t *testing.T) {
	if chaosSetting(true) != 0 {
		t.Fatal("chaos on should persist as 0 (the zero-value default)")
	}
	if chaosSetting(false) != 1 {
		t.Fatal("chaos off should persist as 1")
	}
}

func TestClearWord(t *testing.T) {
	cases := map[int]string{1: "SINGLE", 2: "DOUBLE", 3: "TRIPLE", 4: ""}
	for n, want := range cases {
		if got := clearWord(n); got != want {
			t.Errorf("clearWord(%d) = %q, want %q", n, got, want)
		}
	}
}
