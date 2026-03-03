package main

import (
	"fmt"
	"strings"

	"github.com/0x-JP/tmux-todo/internal/store"
)

type cliEntry struct {
	Todo  store.Todo
	Depth int
}

func flattenForCLI(todos []store.Todo, showDone bool) []cliEntry {
	byParent := make(map[string][]store.Todo, len(todos))
	known := make(map[string]struct{}, len(todos))
	for _, t := range todos {
		known[t.ID] = struct{}{}
	}
	for _, t := range todos {
		p := t.ParentID
		if p != "" {
			if _, ok := known[p]; !ok {
				p = ""
			}
		}
		byParent[p] = append(byParent[p], t)
	}
	out := make([]cliEntry, 0, len(todos))
	seen := map[string]struct{}{}
	var walk func(string, int)
	walk = func(parent string, depth int) {
		for _, t := range byParent[parent] {
			if _, ok := seen[t.ID]; ok {
				continue
			}
			seen[t.ID] = struct{}{}
			if showDone || !t.Done {
				out = append(out, cliEntry{Todo: t, Depth: depth})
			}
			walk(t.ID, depth+1)
		}
	}
	walk("", 0)
	for _, t := range todos {
		if _, ok := seen[t.ID]; ok {
			continue
		}
		if showDone || !t.Done {
			out = append(out, cliEntry{Todo: t, Depth: 0})
		}
	}
	return out
}

type flagStrings []string

func (f *flagStrings) String() string { return strings.Join(*f, ",") }
func (f *flagStrings) Set(v string) error {
	*f = append(*f, v)
	return nil
}

func getTodo(d store.Data, scope store.Scope, contextKey, id string) (store.Todo, error) {
	var list []store.Todo
	if scope == store.ScopeGlobal {
		list = d.Global
	} else {
		list = d.Contexts[contextKey]
	}
	for _, t := range list {
		if t.ID == id {
			return t, nil
		}
	}
	return store.Todo{}, fmt.Errorf("todo id %q not found", id)
}
