package store

import (
	"path/filepath"
	"testing"
)

func TestStoreAddToggleDeleteByScope(t *testing.T) {
	dir := t.TempDir()
	st, err := New(filepath.Join(dir, "todos.json"))
	if err != nil {
		t.Fatal(err)
	}
	ctxKey := "repo=/r|wt=/r/wt|br=feat"

	globalTodo, err := st.Add(ScopeGlobal, "", "general task", "")
	if err != nil {
		t.Fatal(err)
	}
	ctxTodo, err := st.Add(ScopeContext, ctxKey, "ctx task", "")
	if err != nil {
		t.Fatal(err)
	}

	snap := st.Snapshot()
	if got := len(snap.Global); got != 1 {
		t.Fatalf("global count = %d, want 1", got)
	}
	if got := len(snap.Contexts[ctxKey]); got != 1 {
		t.Fatalf("ctx count = %d, want 1", got)
	}

	if err := st.Toggle(ScopeContext, ctxKey, ctxTodo.ID); err != nil {
		t.Fatal(err)
	}
	snap = st.Snapshot()
	if !snap.Contexts[ctxKey][0].Done {
		t.Fatal("expected context todo to be done")
	}
	if snap.Global[0].Done {
		t.Fatal("global todo should be unaffected")
	}

	if err := st.Delete(ScopeGlobal, "", globalTodo.ID); err != nil {
		t.Fatal(err)
	}
	snap = st.Snapshot()
	if got := len(snap.Global); got != 0 {
		t.Fatalf("global count = %d, want 0", got)
	}
}

func TestContextKeyRequired(t *testing.T) {
	dir := t.TempDir()
	st, err := New(filepath.Join(dir, "todos.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.Add(ScopeContext, "", "x", ""); err == nil {
		t.Fatal("expected error for missing context key")
	}
}

func TestDeleteParentCascadesChildren(t *testing.T) {
	dir := t.TempDir()
	st, err := New(filepath.Join(dir, "todos.json"))
	if err != nil {
		t.Fatal(err)
	}
	parent, err := st.Add(ScopeGlobal, "", "parent", "")
	if err != nil {
		t.Fatal(err)
	}
	child, err := st.Add(ScopeGlobal, "", "child", parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.Add(ScopeGlobal, "", "grandchild", child.ID); err != nil {
		t.Fatal(err)
	}

	if err := st.Delete(ScopeGlobal, "", parent.ID); err != nil {
		t.Fatal(err)
	}
	snap := st.Snapshot()
	if len(snap.Global) != 0 {
		t.Fatalf("expected empty list after cascading delete, got %d", len(snap.Global))
	}
}

func TestMoveSubtreeAcrossScopes(t *testing.T) {
	dir := t.TempDir()
	st, err := New(filepath.Join(dir, "todos.json"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := "repo=/r|wt=/r/w|br=b"
	parent, err := st.Add(ScopeContext, ctx, "parent", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.Add(ScopeContext, ctx, "child", parent.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.Move(ScopeContext, ctx, parent.ID, ScopeGlobal, "", ""); err != nil {
		t.Fatal(err)
	}
	snap := st.Snapshot()
	if len(snap.Contexts[ctx]) != 0 {
		t.Fatalf("context should be empty after move, got %d", len(snap.Contexts[ctx]))
	}
	if len(snap.Global) != 2 {
		t.Fatalf("global should contain moved subtree, got %d", len(snap.Global))
	}
}

func TestReparentCycleBlocked(t *testing.T) {
	dir := t.TempDir()
	st, err := New(filepath.Join(dir, "todos.json"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := st.Add(ScopeGlobal, "", "a", "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := st.Add(ScopeGlobal, "", "b", a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Reparent(ScopeGlobal, "", a.ID, b.ID); err == nil {
		t.Fatal("expected cycle detection error")
	}
}
