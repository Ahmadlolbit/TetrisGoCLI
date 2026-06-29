package main

import "testing"

func TestSprintWinRecords(t *testing.T) {
	t.Setenv("CHAOSBLOCKS_CONFIG_DIR", t.TempDir())
	a := &app{board: newScoreboard(), state: scrPlaying}
	a.sess = newSession(modeByKind(modeSprint), 1, 1, true)
	a.sess.g.Lines = modeByKind(modeSprint).sprintLines
	a.checkEnd()
	if !a.sess.won {
		t.Fatal("reaching the line target should win the sprint")
	}
	if a.state != scrGameOver {
		t.Fatalf("state should be game over, got %d", a.state)
	}
	if got := a.board.entries(modeSprint); len(got) != 1 {
		t.Fatalf("a won sprint should record one entry, got %d", len(got))
	}
}

func TestSprintFailDoesNotRecord(t *testing.T) {
	t.Setenv("CHAOSBLOCKS_CONFIG_DIR", t.TempDir())
	a := &app{board: newScoreboard(), state: scrPlaying}
	a.sess = newSession(modeByKind(modeSprint), 1, 1, true)
	a.sess.g.Lines = 20
	a.sess.g.Over = true
	a.checkEnd()
	if a.sess.won {
		t.Fatal("a sprint that did not reach the target is not a win")
	}
	if got := a.board.entries(modeSprint); len(got) != 0 {
		t.Fatalf("a failed sprint must not record, got %d", len(got))
	}
}
