package effects

import (
	"testing"

	"awesomeProject/internal/render"
)

func TestShakeDecaysToZero(t *testing.T) {
	e := New(1)
	e.Shake(2.0, 0.3)
	if !e.Active() {
		t.Fatal("a shake should make the engine active")
	}
	e.Update(0.5)
	if x, y := e.ShakeOffset(); x != 0 || y != 0 {
		t.Fatalf("shake offset should be zero after expiry, got %d,%d", x, y)
	}
	if e.Active() {
		t.Fatal("engine should be inactive once the shake expires")
	}
}

func TestShakeKeepsStrongerLonger(t *testing.T) {
	e := New(1)
	e.Shake(2.0, 0.5)
	e.Shake(0.1, 0.1)
	if e.shakeMag != 2.0 {
		t.Fatalf("a weaker shake should not reduce magnitude, got %v", e.shakeMag)
	}
	if e.shakeDur != 0.5 {
		t.Fatalf("a shorter shake should not reduce duration, got %v", e.shakeDur)
	}
}

func TestParticlesExpire(t *testing.T) {
	e := New(1)
	e.Burst(5, 5, render.RGB(255, 0, 0), 10, 5)
	if len(e.particles) != 10 {
		t.Fatalf("burst should add 10 particles, got %d", len(e.particles))
	}
	for i := 0; i < 100; i++ {
		e.Update(0.1)
	}
	if len(e.particles) != 0 {
		t.Fatalf("all particles should expire, %d remain", len(e.particles))
	}
}

func TestClearResets(t *testing.T) {
	e := New(1)
	e.Burst(1, 1, render.RGB(1, 2, 3), 5, 2)
	e.Flash(0, 0, 4, 2, render.RGB(9, 9, 9), 1)
	e.Shake(1, 1)
	e.Clear()
	if e.Active() {
		t.Fatal("Clear should leave the engine inactive")
	}
}
