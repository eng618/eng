package immich

import (
	"testing"
)

func TestNewCommandStructure(t *testing.T) {
	root := NewCommand()
	if root.Use != "immich" {
		t.Errorf("expected Use immich, got %q", root.Use)
	}

	want := map[string]bool{
		"status": false, "backup": false, "restore": false,
		"start": false, "stop": false, "restart": false, "logs": false,
	}
	for _, sub := range root.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
		if sub.Short == "" {
			t.Errorf("subcommand %s should have a short description", sub.Name())
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected subcommand %q in tree", name)
		}
	}
}

func TestNewCommandIndependentFlagState(t *testing.T) {
	first := NewCommand()
	second := NewCommand()

	s1, _, err := first.Find([]string{"status"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s1.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}

	s2, _, err := second.Find([]string{"status"})
	if err != nil {
		t.Fatal(err)
	}
	if got := s2.Flags().Lookup("json").Value.String(); got != "false" {
		t.Errorf("expected independent flag state between built trees, got json=%s", got)
	}
}
