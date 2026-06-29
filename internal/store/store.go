package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Entry struct {
	Score int     `json:"score"`
	Level int     `json:"level"`
	Lines int     `json:"lines"`
	Combo int     `json:"combo"`
	Time  float64 `json:"time"`
}

type Settings struct {
	Theme      int `json:"theme"`
	StartLevel int `json:"startLevel"`
	ColorMode  int `json:"colorMode"`
}

type State struct {
	Settings Settings           `json:"settings"`
	Scores   map[string][]Entry `json:"scores"`
}

func dir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = filepath.Join(os.TempDir(), "chaosblocks-config")
	}
	return filepath.Join(base, "chaosblocks")
}

func Path() string {
	return filepath.Join(dir(), "state.json")
}

func Load() State {
	def := State{Scores: map[string][]Entry{}}
	data, err := os.ReadFile(Path())
	if err != nil {
		return def
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return def
	}
	if s.Scores == nil {
		s.Scores = map[string][]Entry{}
	}
	return s
}

func Save(s State) error {
	if err := os.MkdirAll(dir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := Path() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, Path()); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
