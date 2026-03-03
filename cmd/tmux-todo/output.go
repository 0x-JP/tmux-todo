package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/0x-JP/tmux-todo/internal/store"
)

func printList(title string, todos []store.Todo, filter store.Filter) {
	fmt.Println(title)
	filtered := store.FilterAndSort(todos, filter)
	entries := flattenForCLI(filtered, filter.ShowDone)
	if len(entries) == 0 {
		fmt.Println("  (empty)")
		fmt.Println()
		return
	}
	for _, e := range entries {
		mark := "- [ ]"
		if e.Todo.Done {
			mark = "- [x]"
		}
		meta := formatMeta(e.Todo)
		fmt.Printf("  %s%s %s (%s)%s\n", strings.Repeat("  ", e.Depth), mark, e.Todo.Text, e.Todo.ID, meta)
	}
	fmt.Println()
}

func formatMeta(t store.Todo) string {
	parts := []string{}
	if t.Priority != "" {
		parts = append(parts, "p="+string(t.Priority))
	}
	if len(t.Tags) > 0 {
		parts = append(parts, "tags="+strings.Join(t.Tags, ","))
	}
	if len(parts) == 0 {
		return ""
	}
	return " [" + strings.Join(parts, " ") + "]"
}

func toJSONTodo(t store.Todo, scope store.Scope, contextKey string) map[string]any {
	return map[string]any{
		"id":           t.ID,
		"text":         t.Text,
		"done":         t.Done,
		"parent_id":    t.ParentID,
		"priority":     string(t.Priority),
		"tags":         t.Tags,
		"scope":        scopeDisplay(scope),
		"context_key":  contextKey,
		"created_at":   t.CreatedAt,
		"completed_at": t.CompletedAt,
	}
}

func toJSONScope(scope, context string, list []store.Todo, showDone bool) map[string]any {
	entries := flattenForCLI(list, showDone)
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		item := toJSONTodo(e.Todo, map[string]store.Scope{"global": store.ScopeGlobal, "context": store.ScopeContext}[scope], context)
		item["depth"] = e.Depth
		out = append(out, item)
	}
	return map[string]any{
		"scope":       scope,
		"context_key": context,
		"todos":       out,
	}
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func jsonMarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func parseScope(v string) (store.Scope, error) {
	switch strings.ToLower(v) {
	case "context":
		return store.ScopeContext, nil
	case "general", "global":
		return store.ScopeGlobal, nil
	default:
		return "", fmt.Errorf("unknown scope %q (expected context|global)", v)
	}
}

func scopeDisplay(s store.Scope) string {
	if s == store.ScopeGlobal {
		return "global"
	}
	return "context"
}
