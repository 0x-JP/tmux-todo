package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/0x-JP/tmux-todo/internal/config"
	"github.com/0x-JP/tmux-todo/internal/gitctx"
	"github.com/0x-JP/tmux-todo/internal/store"
)

func runHasHigh(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("has-high", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var contextKey string
	var contextOnly bool
	var jsonOut bool
	fs.StringVar(&contextKey, "context-key", "", "override context key")
	fs.BoolVar(&contextOnly, "context-only", false, "check only current context (exclude global)")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if contextKey == "" {
		contextKey = ctx.Key()
	}
	d := st.Snapshot()
	ok := false
	if contextOnly {
		ok = store.HasOpenHighPriorityInContext(d, contextKey)
	} else {
		ok = store.HasOpenHighPriority(d, contextKey)
	}
	if ok {
		if jsonOut {
			return printJSON(map[string]any{
				"ok":           true,
				"has_high":     true,
				"context_only": contextOnly,
				"context":      contextKey,
			})
		}
		return nil
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":           true,
			"has_high":     false,
			"context_only": contextOnly,
			"context":      contextKey,
		})
	}
	return errors.New("no open high-priority todos")
}

func runList(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var scopeStr, contextKey, priorityRaw, tag, sortBy string
	var showDone bool
	var jsonOut bool
	fs.StringVar(&scopeStr, "scope", defaultScopeForContext(ctx), "scope: context|global|all")
	fs.StringVar(&contextKey, "context-key", "", "override context key when scope=context")
	fs.BoolVar(&showDone, "all", false, "include done todos")
	fs.StringVar(&priorityRaw, "priority", "", "filter priority")
	fs.StringVar(&tag, "tag", "", "filter tag")
	fs.StringVar(&sortBy, "sort", "priority_due_created", "sort: priority_due_created|created")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := store.NormalizePriority(priorityRaw)
	if err != nil {
		return err
	}
	filter := store.Filter{
		ShowDone: showDone,
		Priority: p,
		Tag:      tag,
		Sort:     sortBy,
	}

	d := st.Snapshot()
	jsonScopes := []map[string]any{}
	switch strings.ToLower(scopeStr) {
	case "context":
		if contextKey == "" {
			contextKey = ctx.Key()
		}
		if jsonOut {
			list := store.FilterAndSort(d.Contexts[contextKey], filter)
			jsonScopes = append(jsonScopes, toJSONScope("context", contextKey, list, filter.ShowDone))
			break
		}
		printList("Context: "+ctx.Label(), d.Contexts[contextKey], filter)
	case "general", "global":
		if jsonOut {
			list := store.FilterAndSort(d.Global, filter)
			jsonScopes = append(jsonScopes, toJSONScope("global", "", list, filter.ShowDone))
			break
		}
		printList("Global", d.Global, filter)
	case "all":
		if jsonOut {
			list := store.FilterAndSort(d.Global, filter)
			jsonScopes = append(jsonScopes, toJSONScope("global", "", list, filter.ShowDone))
		} else {
			printList("Global", d.Global, filter)
		}
		keys := make([]string, 0, len(d.Contexts))
		for k := range d.Contexts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			label := k
			if meta, ok := d.Meta[k]; ok {
				label = fmt.Sprintf("%s/%s [%s]",
					filepath.Base(meta.RepoRoot),
					filepath.Base(meta.WorktreeRoot),
					meta.Branch)
			}
			if jsonOut {
				list := store.FilterAndSort(d.Contexts[k], filter)
				jsonScopes = append(jsonScopes, toJSONScope("context", k, list, filter.ShowDone))
				continue
			}
			printList("Context: "+label, d.Contexts[k], filter)
		}
	default:
		return fmt.Errorf("unknown scope %q", scopeStr)
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":        true,
			"action":    "list",
			"scope":     scopeStr,
			"show_done": showDone,
			"priority":  string(p),
			"tag":       tag,
			"sort":      sortBy,
			"scopes":    jsonScopes,
		})
	}
	return nil
}

func runGet(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var id, scopeStr, contextKey string
	var jsonOut bool
	fs.StringVar(&id, "id", "", "todo ID")
	fs.StringVar(&scopeStr, "scope", defaultScopeForContext(ctx), "scope: context|global")
	fs.StringVar(&contextKey, "context-key", "", "context key for context scope")
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
	t, err := getTodo(st.Snapshot(), scope, contextKey, id)
	if err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":      true,
			"action":  "get",
			"scope":   scopeDisplay(scope),
			"context": contextKey,
			"todo":    toJSONTodo(t, scope, contextKey),
		})
	}
	fmt.Printf("%s %s %s\n", t.ID, scopeDisplay(scope), t.Text)
	return nil
}

func runTags(st *store.Store, cfg *config.Store, args []string) error {
	if len(args) == 0 {
		return errors.New("missing tags subcommand (list|add|remove)")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("tags list", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		var jsonOut bool
		fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		tags := cfg.Tags()
		if jsonOut {
			return printJSON(map[string]any{"ok": true, "action": "tags.list", "tags": tags})
		}
		for _, t := range tags {
			fmt.Println(t)
		}
		return nil
	case "add":
		fs := flag.NewFlagSet("tags add", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		var tag string
		var jsonOut bool
		fs.StringVar(&tag, "tag", "", "tag name")
		fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if tag == "" && len(fs.Args()) > 0 {
			tag = fs.Args()[0]
		}
		if strings.TrimSpace(tag) == "" {
			return errors.New("missing tag")
		}
		if err := cfg.AddTag(tag); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(map[string]any{"ok": true, "action": "tags.add", "tag": strings.ToLower(strings.TrimSpace(tag))})
		}
		fmt.Println("added tag", strings.ToLower(strings.TrimSpace(tag)))
		return nil
	case "remove":
		fs := flag.NewFlagSet("tags remove", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		var tag string
		var clean bool
		var jsonOut bool
		fs.StringVar(&tag, "tag", "", "tag name")
		fs.BoolVar(&clean, "global-clean", true, "remove this tag from all existing todos")
		fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if tag == "" && len(fs.Args()) > 0 {
			tag = fs.Args()[0]
		}
		if strings.TrimSpace(tag) == "" {
			return errors.New("missing tag")
		}
		if err := cfg.RemoveTag(tag); err != nil {
			return err
		}
		if clean {
			if err := st.RemoveTag(tag); err != nil {
				return err
			}
		}
		if jsonOut {
			return printJSON(map[string]any{"ok": true, "action": "tags.remove", "tag": strings.ToLower(strings.TrimSpace(tag)), "global_clean": clean})
		}
		fmt.Println("removed tag", strings.ToLower(strings.TrimSpace(tag)))
		return nil
	default:
		return fmt.Errorf("unknown tags subcommand %q (expected list|add|remove)", args[0])
	}
}
