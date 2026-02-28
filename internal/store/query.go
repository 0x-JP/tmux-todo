package store

import (
	"sort"
	"strings"
	"time"
)

type Filter struct {
	ShowDone bool
	Overdue  bool
	Priority Priority
	Tag      string
	Sort     string
	Now      time.Time
}

func FilterAndSort(todos []Todo, f Filter) []Todo {
	out := make([]Todo, 0, len(todos))
	tag := strings.ToLower(strings.TrimSpace(f.Tag))
	for _, t := range todos {
		if !f.ShowDone && t.Done {
			continue
		}
		if f.Overdue && !IsOverdue(t, f.now()) {
			continue
		}
		if f.Priority != "" && t.Priority != f.Priority {
			continue
		}
		if tag != "" {
			ok := false
			for _, tg := range t.Tags {
				if strings.EqualFold(tg, tag) {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		out = append(out, t)
	}
	sortBy := strings.ToLower(strings.TrimSpace(f.Sort))
	if sortBy == "" {
		sortBy = "priority_due_created"
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch sortBy {
		case "created":
			return a.CreatedAt.Before(b.CreatedAt)
		case "due":
			return dueLess(a.DueAt, b.DueAt)
		default:
			ra := PriorityRank(a.Priority)
			rb := PriorityRank(b.Priority)
			if ra != rb {
				return ra > rb
			}
			if !sameDue(a.DueAt, b.DueAt) {
				return dueLess(a.DueAt, b.DueAt)
			}
			return a.CreatedAt.Before(b.CreatedAt)
		}
	})
	return out
}

func HasOpenHighPriority(d Data, contextKey string) bool {
	check := func(list []Todo) bool {
		for _, t := range list {
			if !t.Done && t.Priority == PriorityHigh {
				return true
			}
		}
		return false
	}
	if check(d.Contexts[contextKey]) {
		return true
	}
	return check(d.Global)
}

func HasOpenHighPriorityInContext(d Data, contextKey string) bool {
	for _, t := range d.Contexts[contextKey] {
		if !t.Done && t.Priority == PriorityHigh {
			return true
		}
	}
	return false
}

func (f Filter) now() time.Time {
	if f.Now.IsZero() {
		return time.Now().UTC()
	}
	return f.Now.UTC()
}

func dueLess(a, b *time.Time) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return a.Before(*b)
}

func sameDue(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
