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
