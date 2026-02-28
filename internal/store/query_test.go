package store

import (
	"testing"
	"time"
)

func TestFilterAndSort(t *testing.T) {
	now := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC)
	early := now.Add(2 * time.Hour)
	late := now.Add(6 * time.Hour)
	todos := []Todo{
		{ID: "1", Text: "low", Priority: PriorityLow, CreatedAt: now.Add(-3 * time.Hour), DueAt: &late},
		{ID: "2", Text: "high", Priority: PriorityHigh, CreatedAt: now.Add(-2 * time.Hour), DueAt: &early, Tags: []string{"blocked"}},
		{ID: "3", Text: "done", Priority: PriorityHigh, Done: true, CreatedAt: now.Add(-1 * time.Hour)},
	}
	out := FilterAndSort(todos, Filter{Now: now})
	if len(out) != 2 {
		t.Fatalf("len=%d want 2", len(out))
	}
	if out[0].ID != "2" {
		t.Fatalf("expected high-priority first, got %s", out[0].ID)
	}

	tagged := FilterAndSort(todos, Filter{Now: now, ShowDone: true, Tag: "blocked"})
	if len(tagged) != 1 || tagged[0].ID != "2" {
		t.Fatalf("unexpected tag filter result: %#v", tagged)
	}
}

func TestHasOpenHighPriority(t *testing.T) {
	d := Data{
		Global: []Todo{{Priority: PriorityLow}},
		Contexts: map[string][]Todo{
			"k": {
				{Priority: PriorityHigh, Done: false},
			},
		},
	}
	if !HasOpenHighPriority(d, "k") {
		t.Fatal("expected high priority to be detected")
	}
}
