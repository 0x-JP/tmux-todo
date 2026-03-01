package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jp/tmux-todo/internal/gitctx"
	"github.com/jp/tmux-todo/internal/store"
)

func runAdd(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var text, scopeStr, contextKey, parentID, priorityRaw, tagsCSV string
	var jsonOut bool
	var tags flagStrings
	fs.StringVar(&text, "text", "", "todo text")
	fs.StringVar(&scopeStr, "scope", defaultScopeForContext(ctx), "target scope: context|global")
	fs.StringVar(&contextKey, "context-key", "", "override context key (scope=context)")
	fs.StringVar(&parentID, "parent", "", "parent todo ID for nested todo")
	fs.StringVar(&priorityRaw, "priority", "", "priority: low|med|high")
	fs.StringVar(&tagsCSV, "tags", "", "comma-separated tags")
	fs.Var(&tags, "tag", "single tag; can repeat")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if text == "" && len(fs.Args()) > 0 {
		text = strings.Join(fs.Args(), " ")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("missing todo text (use --text or positional text)")
	}

	scope, err := parseScope(scopeStr)
	if err != nil {
		return err
	}
	if scope == store.ScopeContext && contextKey == "" {
		contextKey = ctx.Key()
	}
	pr, err := store.NormalizePriority(priorityRaw)
	if err != nil {
		return err
	}
	allTags := append(tags, tagsCSV)
	t, err := st.AddWithParams(scope, contextKey, store.AddParams{
		Text:     text,
		ParentID: parentID,
		Priority: pr,
		Tags:     allTags,
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":      true,
			"action":  "add",
			"scope":   scopeDisplay(scope),
			"context": contextKey,
			"todo":    toJSONTodo(t, scope, contextKey),
		})
	}
	fmt.Printf("added %s [%s]\n", t.ID, scopeDisplay(scope))
	return nil
}

func runDone(st *store.Store, ctx gitctx.Context, args []string, done bool) error {
	fs := flag.NewFlagSet("done", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var id, scopeStr, contextKey string
	var jsonOut bool
	fs.StringVar(&id, "id", "", "todo ID")
	fs.StringVar(&scopeStr, "scope", defaultScopeForContext(ctx), "todo scope: context|global")
	fs.StringVar(&contextKey, "context-key", "", "override context key (scope=context)")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if id == "" {
		return errors.New("missing --id")
	}
	scope, err := parseScope(scopeStr)
	if err != nil {
		return err
	}
	if scope == store.ScopeContext && contextKey == "" {
		contextKey = ctx.Key()
	}
	if err := st.SetDone(scope, contextKey, id, done); err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":      true,
			"action":  map[bool]string{true: "done", false: "undone"}[done],
			"id":      id,
			"scope":   scopeDisplay(scope),
			"context": contextKey,
		})
	}
	if done {
		fmt.Printf("done %s [%s]\n", id, scopeDisplay(scope))
	} else {
		fmt.Printf("undone %s [%s]\n", id, scopeDisplay(scope))
	}
	return nil
}

func runEdit(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var id, scopeStr, contextKey, text, priorityRaw, tagsCSV string
	var clearPriority, clearTags bool
	var jsonOut bool
	var tags flagStrings
	fs.StringVar(&id, "id", "", "todo ID")
	fs.StringVar(&scopeStr, "scope", defaultScopeForContext(ctx), "todo scope: context|global")
	fs.StringVar(&contextKey, "context-key", "", "override context key (scope=context)")
	fs.StringVar(&text, "text", "", "new text")
	fs.StringVar(&priorityRaw, "priority", "", "new priority low|med|high")
	fs.BoolVar(&clearPriority, "clear-priority", false, "clear priority")
	fs.StringVar(&tagsCSV, "tags", "", "replace tags with CSV")
	fs.Var(&tags, "tag", "replace tags with repeated tags")
	fs.BoolVar(&clearTags, "clear-tags", false, "clear all tags")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if id == "" {
		return errors.New("missing --id")
	}
	scope, err := parseScope(scopeStr)
	if err != nil {
		return err
	}
	if scope == store.ScopeContext && contextKey == "" {
		contextKey = ctx.Key()
	}
	up := store.UpdateParams{}
	if strings.TrimSpace(text) != "" {
		v := strings.TrimSpace(text)
		up.Text = &v
	}
	if clearPriority {
		empty := store.Priority("")
		up.Priority = &empty
	} else if strings.TrimSpace(priorityRaw) != "" {
		p, err := store.NormalizePriority(priorityRaw)
		if err != nil {
			return err
		}
		up.Priority = &p
	}
	if clearTags {
		empty := []string{}
		up.Tags = &empty
	} else if tagsCSV != "" || len(tags) > 0 {
		allTags := append([]string(nil), tags...)
		if tagsCSV != "" {
			allTags = append(allTags, tagsCSV)
		}
		up.Tags = &allTags
	}
	t, err := st.Update(scope, contextKey, id, up)
	if err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":      true,
			"action":  "edit",
			"scope":   scopeDisplay(scope),
			"context": contextKey,
			"todo":    toJSONTodo(t, scope, contextKey),
		})
	}
	fmt.Printf("edited %s [%s]\n", id, scopeDisplay(scope))
	return nil
}

func runDelete(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var id, scopeStr, contextKey string
	var jsonOut bool
	fs.StringVar(&id, "id", "", "todo ID")
	fs.StringVar(&scopeStr, "scope", defaultScopeForContext(ctx), "todo scope: context|global")
	fs.StringVar(&contextKey, "context-key", "", "override context key (scope=context)")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if id == "" {
		return errors.New("missing --id")
	}
	scope, err := parseScope(scopeStr)
	if err != nil {
		return err
	}
	if scope == store.ScopeContext && contextKey == "" {
		contextKey = ctx.Key()
	}
	if err := st.Delete(scope, contextKey, id); err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":      true,
			"action":  "delete",
			"id":      id,
			"scope":   scopeDisplay(scope),
			"context": contextKey,
		})
	}
	fmt.Printf("deleted %s [%s]\n", id, scopeDisplay(scope))
	return nil
}

func runMove(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("move", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var id, fromScope, fromCtx, toScope, toCtx, toParent string
	var jsonOut bool
	fs.StringVar(&id, "id", "", "todo ID")
	fs.StringVar(&fromScope, "from-scope", defaultScopeForContext(ctx), "source scope")
	fs.StringVar(&fromCtx, "from-context-key", "", "source context key")
	fs.StringVar(&toScope, "to-scope", "global", "target scope")
	fs.StringVar(&toCtx, "to-context-key", "", "target context key")
	fs.StringVar(&toParent, "to-parent", "", "target parent id")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if id == "" {
		return errors.New("missing --id")
	}
	src, err := parseScope(fromScope)
	if err != nil {
		return err
	}
	dst, err := parseScope(toScope)
	if err != nil {
		return err
	}
	if src == store.ScopeContext && fromCtx == "" {
		fromCtx = ctx.Key()
	}
	if dst == store.ScopeContext && toCtx == "" {
		toCtx = ctx.Key()
	}
	if err := st.Move(src, fromCtx, id, dst, toCtx, toParent); err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":           true,
			"action":       "move",
			"id":           id,
			"from_scope":   scopeDisplay(src),
			"from_context": fromCtx,
			"to_scope":     scopeDisplay(dst),
			"to_context":   toCtx,
			"to_parent_id": toParent,
		})
	}
	fmt.Printf("moved %s [%s -> %s]\n", id, scopeDisplay(src), scopeDisplay(dst))
	return nil
}

func runReparent(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("reparent", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var id, scopeStr, contextKey, parentID string
	var jsonOut bool
	fs.StringVar(&id, "id", "", "todo ID")
	fs.StringVar(&scopeStr, "scope", defaultScopeForContext(ctx), "scope: context|global")
	fs.StringVar(&contextKey, "context-key", "", "context key for context scope")
	fs.StringVar(&parentID, "parent", "", "new parent id; empty for root")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if id == "" {
		return errors.New("missing --id")
	}
	scope, err := parseScope(scopeStr)
	if err != nil {
		return err
	}
	if scope == store.ScopeContext && contextKey == "" {
		contextKey = ctx.Key()
	}
	if err := st.Reparent(scope, contextKey, id, parentID); err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":        true,
			"action":    "reparent",
			"id":        id,
			"scope":     scopeDisplay(scope),
			"context":   contextKey,
			"parent_id": parentID,
		})
	}
	fmt.Printf("reparented %s [%s]\n", id, scopeDisplay(scope))
	return nil
}
