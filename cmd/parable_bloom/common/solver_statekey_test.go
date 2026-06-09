//go:build ignore
// +build ignore

package common

import "testing"

func TestParableBloomTestsRemoved(t *testing.T) {
	t.Skip("Parable Bloom eng CLI tests removed; use tools/level-builder tests instead.")
}

func TestStateKeyDeterministic(t *testing.T) {
	a := make(map[string]bool)
	a["b"] = true
	a["a"] = true

	key := stateKey(a)
	if key != "a,b," {
		t.Fatalf("expected ordered key 'a,b,', got %q", key)
	}

	b := make(map[string]bool)
	b["a"] = true
	b["b"] = true
	key2 := stateKey(b)
	if key2 != key {
		t.Fatalf("stateKey not deterministic: %q vs %q", key, key2)
	}
}
