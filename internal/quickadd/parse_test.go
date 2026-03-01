package quickadd

import (
	"strings"
	"testing"

	"github.com/jp/tmux-todo/internal/store"
)

func TestParse(t *testing.T) {
	ctx := "repo=/r|wt=/r/w|br=main"
	tests := []struct {
		name    string
		input   string
		want    Spec
		wantErr bool
	}{
		{
			name:  "plain context",
			input: "task 1",
			want: Spec{
				Scope:      store.ScopeContext,
				ContextKey: ctx,
				Text:       "task 1",
			},
		},
		{
			name:  "global prefix",
			input: "global | task 1",
			want: Spec{
				Scope:      store.ScopeGlobal,
				ContextKey: "",
				Text:       "task 1",
			},
		},
		{
			name:  "global with numeric priority",
			input: "global | task 1 | p=1",
			want: Spec{
				Scope:      store.ScopeGlobal,
				ContextKey: "",
				Text:       "task 1",
				Priority:   store.PriorityHigh,
			},
		},
		{
			name:  "context with text priority",
			input: "task 1 | p=high",
			want: Spec{
				Scope:      store.ScopeContext,
				ContextKey: ctx,
				Text:       "task 1",
				Priority:   store.PriorityHigh,
			},
		},
		{
			name:  "med and context prefix",
			input: "context | task 2 | p=2",
			want: Spec{
				Scope:      store.ScopeContext,
				ContextKey: ctx,
				Text:       "task 2",
				Priority:   store.PriorityMed,
			},
		},
		{
			name:  "tags",
			input: "task 3 | t=blocked,review",
			want: Spec{
				Scope:      store.ScopeContext,
				ContextKey: ctx,
				Text:       "task 3",
				Tags:       []string{"blocked", "review"},
			},
		},
		{
			name:  "tags alias and priority",
			input: "global | task 4 | tag=blocked | p=1",
			want: Spec{
				Scope:      store.ScopeGlobal,
				ContextKey: "",
				Text:       "task 4",
				Priority:   store.PriorityHigh,
				Tags:       []string{"blocked"},
			},
		},
		{name: "missing text", input: "global | p=1", wantErr: true},
		{name: "bad option", input: "task 1 | x=1", wantErr: true},
		{name: "bad priority", input: "task 1 | p=9", wantErr: true},
		{name: "extra segment", input: "task 1 | note", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input, ctx)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Scope != tc.want.Scope ||
				got.ContextKey != tc.want.ContextKey ||
				got.Text != tc.want.Text ||
				got.Priority != tc.want.Priority ||
				strings.Join(got.Tags, ",") != strings.Join(tc.want.Tags, ",") {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
