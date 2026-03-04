package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigTagLifecycle(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cfg.json")
	s, err := New(p, []string{"review", "blocked"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddTag("Whatever"); err != nil {
		t.Fatal(err)
	}
	if err := s.RemoveTag("whatever"); err != nil {
		t.Fatal(err)
	}
	tags := s.Tags()
	for _, tg := range tags {
		if tg == "whatever" {
			t.Fatal("tag should have been removed")
		}
	}
}

func TestConfigUIStateLifecycle(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cfg.json")
	s, err := New(p, []string{"review", "blocked"})
	if err != nil {
		t.Fatal(err)
	}
	ui := UIState{MainMode: "all"}
	ui.Selected.Scope = "context"
	ui.Selected.ContextKey = "repo=/r|wt=/w|br=main"
	ui.Selected.ID = "abc123"
	if err := s.SaveUI(ui); err != nil {
		t.Fatal(err)
	}
	got := s.UI()
	if got.MainMode != "all" || got.Selected.Scope != "context" || got.Selected.ID != "abc123" {
		t.Fatalf("unexpected ui state: %+v", got)
	}
}

func TestKeybindingsNilWhenAbsent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cfg.json")
	s, err := New(p, []string{"review"})
	if err != nil {
		t.Fatal(err)
	}
	if kb := s.Keybindings(); kb != nil {
		t.Fatalf("expected nil keybindings, got %v", kb)
	}
}

func TestKeybindingsRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cfg.json")

	// Write a config with keybindings manually.
	initial := `{
		"version": 2,
		"tags": ["review"],
		"keybindings": {
			"quit": ["q", "ctrl+c"],
			"delete": ["x"]
		}
	}`
	if err := os.WriteFile(p, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := New(p, []string{"review"})
	if err != nil {
		t.Fatal(err)
	}
	kb := s.Keybindings()
	if kb == nil {
		t.Fatal("expected keybindings to be non-nil")
	}
	if len(kb["quit"]) != 2 || kb["quit"][0] != "q" || kb["quit"][1] != "ctrl+c" {
		t.Fatalf("unexpected quit keybindings: %v", kb["quit"])
	}
	if len(kb["delete"]) != 1 || kb["delete"][0] != "x" {
		t.Fatalf("unexpected delete keybindings: %v", kb["delete"])
	}
}
