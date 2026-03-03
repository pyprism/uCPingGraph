package utils

import "testing"

func TestRandomFloatRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := RandomFloat()
		if v < 2.0 || v > 20.0 {
			t.Fatalf("RandomFloat() = %f, want [2.0, 20.0]", v)
		}
	}
}

func TestRandomIntRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := RandomInt()
		if v < 0 || v >= 10 {
			t.Fatalf("RandomInt() = %d, want [0, 10)", v)
		}
	}
}
