package ui

import (
	"testing"

	"github.com/jp/tmux-todo/internal/store"
)

func TestFlattenTodosScopeAndDepth(t *testing.T) {
	todos := []store.Todo{
		{ID: "a", Text: "parent"},
		{ID: "b", Text: "child", ParentID: "a"},
	}
	out := flattenTodos(todos, false, store.ScopeContext, "ctx1")
	if len(out) != 2 {
		t.Fatalf("len=%d", len(out))
	}
	if out[0].Depth != 0 || out[1].Depth != 1 {
		t.Fatalf("depths=%d,%d", out[0].Depth, out[1].Depth)
	}
	if out[1].Scope != store.ScopeContext || out[1].CtxKey != "ctx1" {
		t.Fatalf("unexpected scope info: %#v", out[1])
	}
}

func TestRenderMeta(t *testing.T) {
	got := renderMeta(store.Todo{
		Priority: store.PriorityMed,
		Tags:     []string{"blocked"},
	})
	if got == "" {
		t.Fatal("expected metadata rendering")
	}
}

func TestToggleTag(t *testing.T) {
	got := toggleTag([]string{"review"}, "blocked")
	if !hasTag(got, "blocked") || !hasTag(got, "review") {
		t.Fatalf("unexpected tags: %v", got)
	}
	got = toggleTag(got, "blocked")
	if hasTag(got, "blocked") {
		t.Fatalf("expected blocked removed: %v", got)
	}
}
