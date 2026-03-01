package config

import (
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
