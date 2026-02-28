package store

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Priority string

const (
	PriorityLow  Priority = "low"
	PriorityMed  Priority = "med"
	PriorityHigh Priority = "high"
)

var DefaultTags = []string{"blocked", "review"}

func NormalizePriority(v string) (Priority, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return "", nil
	case "low", "l", "p3":
		return PriorityLow, nil
	case "med", "medium", "m", "p2":
		return PriorityMed, nil
	case "high", "h", "p1":
		return PriorityHigh, nil
	default:
		return "", fmt.Errorf("invalid priority %q (expected low|med|high)", v)
	}
}

func ParseDue(v string) (*time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	if d, err := time.Parse("2006-01-02", v); err == nil {
		u := d.UTC()
		return &u, nil
	}
	if d, err := time.Parse(time.RFC3339, v); err == nil {
		u := d.UTC()
		return &u, nil
	}
	return nil, fmt.Errorf("invalid due date %q (expected YYYY-MM-DD or RFC3339)", v)
}

func NormalizeTags(tags []string) []string {
	set := map[string]struct{}{}
	for _, t := range tags {
		for _, p := range strings.Split(t, ",") {
			v := strings.ToLower(strings.TrimSpace(p))
			if strings.HasPrefix(v, "@") {
				v = strings.TrimPrefix(v, "@")
			}
			if v == "" {
				continue
			}
			set[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func KnownTags(d Data) []string {
	set := map[string]struct{}{}
	for _, t := range DefaultTags {
		set[t] = struct{}{}
	}
	collect := func(list []Todo) {
		for _, td := range list {
			for _, tg := range td.Tags {
				v := strings.ToLower(strings.TrimSpace(tg))
				if v != "" {
					set[v] = struct{}{}
				}
			}
		}
	}
	collect(d.Global)
	for _, list := range d.Contexts {
		collect(list)
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func PriorityRank(p Priority) int {
	switch p {
	case PriorityHigh:
		return 3
	case PriorityMed:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

func IsOverdue(t Todo, now time.Time) bool {
	if t.Done || t.DueAt == nil {
		return false
	}
	return t.DueAt.Before(now.UTC())
}
