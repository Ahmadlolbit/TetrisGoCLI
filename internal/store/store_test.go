package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("CHAOSBLOCKS_CONFIG_DIR", t.TempDir())
	in := State{
		Settings: Settings{Theme: 2, StartLevel: 7},
		Scores: map[string][]Entry{
			"Marathon": {{Score: 1000, Level: 5, Lines: 42, Combo: 3}},
			"Sprint":   {{Score: 500, Level: 2, Lines: 40, Combo: 1, Time: 83.45}},
		},
	}
	if err := Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	out := Load()
	if out.Settings != in.Settings {
		t.Fatalf("settings = %+v, want %+v", out.Settings, in.Settings)
	}
	if len(out.Scores["Marathon"]) != 1 || out.Scores["Marathon"][0].Score != 1000 {
		t.Fatalf("marathon not round-tripped: %+v", out.Scores["Marathon"])
	}
	if out.Scores["Sprint"][0].Time != 83.45 {
		t.Fatalf("sprint time not round-tripped: %v", out.Scores["Sprint"][0].Time)
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	t.Setenv("CHAOSBLOCKS_CONFIG_DIR", t.TempDir())
	s := Load()
	if s.Scores == nil || len(s.Scores) != 0 {
		t.Fatalf("missing file should yield empty scores, got %+v", s.Scores)
	}
	if s.Settings.StartLevel != 0 {
		t.Fatalf("missing file should yield zero settings, got %+v", s.Settings)
	}
}

func TestLoadCorruptReturnsDefaults(t *testing.T) {
	t.Setenv("CHAOSBLOCKS_CONFIG_DIR", t.TempDir())
	if err := os.MkdirAll(filepath.Dir(Path()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(), []byte("{not valid json!!"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := Load()
	if s.Scores == nil || len(s.Scores) != 0 {
		t.Fatalf("corrupt file should yield empty defaults, got %+v", s.Scores)
	}
}
