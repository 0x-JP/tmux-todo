package ui

import (
	"testing"

	"github.com/0x-JP/tmux-todo/internal/gitctx"
	"github.com/0x-JP/tmux-todo/internal/quickadd"
	"github.com/0x-JP/tmux-todo/internal/store"
)

func TestNormalizeQuickSpecForContext(t *testing.T) {
	t.Run("non git defaults context scope to global", func(t *testing.T) {
		ctx := gitctx.Context{Branch: "global"}
		spec := quickadd.Spec{
			Scope:      store.ScopeContext,
			ContextKey: "global",
			Text:       "x",
		}
		got := normalizeQuickSpecForContext(ctx, spec)
		if got.Scope != store.ScopeGlobal || got.ContextKey != "" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("explicit global unchanged", func(t *testing.T) {
		ctx := gitctx.Context{Branch: "global"}
		spec := quickadd.Spec{
			Scope: store.ScopeGlobal,
			Text:  "x",
		}
		got := normalizeQuickSpecForContext(ctx, spec)
		if got.Scope != store.ScopeGlobal {
			t.Fatalf("got %+v", got)
		}
	})
}
