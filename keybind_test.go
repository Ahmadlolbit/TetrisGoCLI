package main

import (
	"testing"

	"awesomeProject/internal/input"
)

func TestRebindSwap(t *testing.T) {
	a := &app{keymap: input.DefaultKeymap()}
	a.rebind(input.MoveLeft, "x")
	if a.keymap[input.MoveLeft] != "x" {
		t.Errorf("MoveLeft = %q, want x", a.keymap[input.MoveLeft])
	}
	if a.keymap[input.RotateCW] != "h" {
		t.Errorf("RotateCW = %q, want h (swapped from MoveLeft)", a.keymap[input.RotateCW])
	}
}

func TestRebindFreshKey(t *testing.T) {
	a := &app{keymap: input.DefaultKeymap()}
	a.rebind(input.MoveLeft, "a")
	if a.keymap[input.MoveLeft] != "a" {
		t.Errorf("MoveLeft = %q, want a", a.keymap[input.MoveLeft])
	}
	for _, e := range input.Bindable {
		if e == input.MoveLeft {
			continue
		}
		if a.keymap[e] != input.DefaultKeymap()[e] {
			t.Errorf("action %v changed unexpectedly to %q", e, a.keymap[e])
		}
	}
}

func TestRebindUniqueInvariant(t *testing.T) {
	a := &app{keymap: input.DefaultKeymap()}
	a.rebind(input.SoftDrop, "l")
	a.rebind(input.HardDrop, "j")
	seen := map[input.Key]input.Event{}
	for _, e := range input.Bindable {
		k := a.keymap[e]
		if prev, dup := seen[k]; dup {
			t.Errorf("key %q bound to both %v and %v", k, prev, e)
		}
		seen[k] = e
	}
}
