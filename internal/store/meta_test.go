package store

import (
	"testing"
	"time"
)

func TestNormalizePriority(t *testing.T) {
	p, err := NormalizePriority("HIGH")
	if err != nil {
		t.Fatal(err)
	}
	if p != PriorityHigh {
		t.Fatalf("got %q", p)
	}
	if _, err := NormalizePriority("urgent"); err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestParseDue(t *testing.T) {
	d, err := ParseDue("2026-03-01")
	if err != nil {
		t.Fatal(err)
	}
	if d == nil || d.Year() != 2026 || d.Month() != 3 || d.Day() != 1 {
		t.Fatalf("unexpected parsed due date: %#v", d)
	}
	if _, err := ParseDue("03/01/2026"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestNormalizeTags(t *testing.T) {
	got := NormalizeTags([]string{"@Blocked", "review", "blocked,needs-test", "  "})
	want := []string{"blocked", "needs-test", "review"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestIsOverdue(t *testing.T) {
	now := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC)
	due := now.Add(-time.Hour)
	if !IsOverdue(Todo{DueAt: &due}, now) {
		t.Fatal("expected overdue")
	}
	if IsOverdue(Todo{DueAt: &due, Done: true}, now) {
		t.Fatal("done todo should not be overdue")
	}
}
