package quickadd

import (
	"fmt"
	"strings"

	"github.com/jp/tmux-todo/internal/store"
)

type Spec struct {
	Scope      store.Scope
	ContextKey string
	Text       string
	Priority   store.Priority
}

func Parse(input, defaultContextKey string) (Spec, error) {
	spec := Spec{
		Scope:      store.ScopeContext,
		ContextKey: defaultContextKey,
	}
	parts := strings.Split(input, "|")
	for i, raw := range parts {
		part := strings.TrimSpace(raw)
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		if i == 0 {
			switch lower {
			case "global", "general":
				spec.Scope = store.ScopeGlobal
				spec.ContextKey = ""
				continue
			case "context":
				spec.Scope = store.ScopeContext
				spec.ContextKey = defaultContextKey
				continue
			}
		}
		if strings.Contains(part, "=") {
			key, value, _ := strings.Cut(part, "=")
			key = strings.ToLower(strings.TrimSpace(key))
			value = strings.TrimSpace(value)
			switch key {
			case "p", "priority":
				p, err := parsePriority(value)
				if err != nil {
					return Spec{}, err
				}
				spec.Priority = p
			default:
				return Spec{}, fmt.Errorf("unknown option %q (supported: p=high|med|low or p=1|2|3)", key)
			}
			continue
		}
		if spec.Text == "" {
			spec.Text = part
			continue
		}
		return Spec{}, fmt.Errorf("unexpected segment %q", part)
	}
	spec.Text = strings.TrimSpace(spec.Text)
	if spec.Text == "" {
		return Spec{}, fmt.Errorf("missing task text")
	}
	return spec, nil
}

func parsePriority(v string) (store.Priority, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1":
		return store.PriorityHigh, nil
	case "2":
		return store.PriorityMed, nil
	case "3":
		return store.PriorityLow, nil
	default:
		return store.NormalizePriority(v)
	}
}
