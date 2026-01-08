package main

import (
	"os"
	"testing"
)

func TestWriteSweepToFile_Smoke(t *testing.T) {
	out := "test_sweep_out.json"
	defer os.Remove(out)
	if err := WriteSweepToFile("Seedling", 1, out); err != nil {
		t.Fatalf("WriteSweepToFile failed: %v", err)
	}
	fi, err := os.Stat(out)
	if err != nil {
		t.Fatalf("expected output file, got error: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("expected non-empty output file")
	}
}
