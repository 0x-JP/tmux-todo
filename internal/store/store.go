package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/0x-JP/tmux-todo/internal/fileio"
)

type Todo struct {
	ID          string     `json:"id"`
	ParentID    string     `json:"parent_id,omitempty"`
	Text        string     `json:"text"`
	Priority    Priority   `json:"priority,omitempty"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Data struct {
	Version  int                 `json:"version"`
	Global   []Todo              `json:"global"`
	Contexts map[string][]Todo   `json:"contexts"`
	Meta     map[string]MetaInfo `json:"meta,omitempty"`
}

type MetaInfo struct {
	RepoRoot     string `json:"repo_root"`
	WorktreeRoot string `json:"worktree_root"`
	Branch       string `json:"branch"`
}

type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeContext Scope = "context"
)

type Store struct {
	mu   sync.Mutex
	path string
	data Data
}

type AddParams struct {
	Text     string
	ParentID string
	Priority Priority
	DueAt    *time.Time
	Tags     []string
}

type UpdateParams struct {
	Text     *string
	Priority *Priority
	DueAt    **time.Time
	Tags     *[]string
	ParentID *string
}

func New(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Snapshot() Data {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.data
	out.Global = append([]Todo(nil), s.data.Global...)
	out.Contexts = make(map[string][]Todo, len(s.data.Contexts))
	for k, v := range s.data.Contexts {
		out.Contexts[k] = append([]Todo(nil), v...)
	}
	out.Meta = make(map[string]MetaInfo, len(s.data.Meta))
	for k, v := range s.data.Meta {
		out.Meta[k] = v
	}
	return out
}

func (s *Store) SetContextMeta(key string, meta MetaInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Meta == nil {
		s.data.Meta = map[string]MetaInfo{}
	}
	s.data.Meta[key] = meta
	return s.saveLocked()
}

func (s *Store) KnownContextKeys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]string, 0, len(s.data.Contexts))
	for k := range s.data.Contexts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *Store) Add(scope Scope, contextKey, text, parentID string) (Todo, error) {
	return s.AddWithParams(scope, contextKey, AddParams{
		Text:     text,
		ParentID: parentID,
	})
}

func (s *Store) AddWithParams(scope Scope, contextKey string, params AddParams) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if scope == ScopeContext && contextKey == "" {
		return Todo{}, errors.New("context key required")
	}
	if strings.TrimSpace(params.Text) == "" {
		return Todo{}, errors.New("todo text cannot be empty")
	}
	if params.Priority != "" {
		norm, err := NormalizePriority(string(params.Priority))
		if err != nil {
			return Todo{}, err
		}
		params.Priority = norm
	}
	params.Tags = NormalizeTags(params.Tags)

	id, err := newID()
	if err != nil {
		return Todo{}, err
	}
	t := Todo{
		ID:        id,
		ParentID:  params.ParentID,
		Text:      strings.TrimSpace(params.Text),
		Priority:  params.Priority,
		DueAt:     params.DueAt,
		Tags:      params.Tags,
		CreatedAt: time.Now().UTC(),
	}
	switch scope {
	case ScopeGlobal:
		s.data.Global = append(s.data.Global, t)
	case ScopeContext:
		if s.data.Contexts == nil {
			s.data.Contexts = map[string][]Todo{}
		}
		s.data.Contexts[contextKey] = append(s.data.Contexts[contextKey], t)
	default:
		return Todo{}, fmt.Errorf("unknown scope %q", scope)
	}
	return t, s.saveLocked()
}

func (s *Store) Update(scope Scope, contextKey, id string, params UpdateParams) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.listForScope(scope, contextKey)
	if err != nil {
		return Todo{}, err
	}
	idx := -1
	for i := range list {
		if list[i].ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return Todo{}, fmt.Errorf("todo id %q not found", id)
	}

	t := list[idx]
	if params.Text != nil {
		v := strings.TrimSpace(*params.Text)
		if v == "" {
			return Todo{}, errors.New("todo text cannot be empty")
		}
		t.Text = v
	}
	if params.Priority != nil {
		if *params.Priority == "" {
			t.Priority = ""
		} else {
			norm, err := NormalizePriority(string(*params.Priority))
			if err != nil {
				return Todo{}, err
			}
			t.Priority = norm
		}
	}
	if params.DueAt != nil {
		t.DueAt = *params.DueAt
	}
	if params.Tags != nil {
		t.Tags = NormalizeTags(*params.Tags)
	}
	if params.ParentID != nil {
		if err := validateReparent(list, id, *params.ParentID); err != nil {
			return Todo{}, err
		}
		t.ParentID = *params.ParentID
	}
	list[idx] = t
	s.setListForScope(scope, contextKey, list)
	return t, s.saveLocked()
}

func (s *Store) Toggle(scope Scope, contextKey, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var list []Todo
	switch scope {
	case ScopeGlobal:
		list = s.data.Global
	case ScopeContext:
		if contextKey == "" {
			return errors.New("context key required")
		}
		list = s.data.Contexts[contextKey]
	default:
		return fmt.Errorf("unknown scope %q", scope)
	}

	found := false
	now := time.Now().UTC()
	for i := range list {
		if list[i].ID == id {
			found = true
			list[i].Done = !list[i].Done
			if list[i].Done {
				list[i].CompletedAt = &now
			} else {
				list[i].CompletedAt = nil
			}
			break
		}
	}
	if !found {
		return fmt.Errorf("todo id %q not found", id)
	}

	if scope == ScopeGlobal {
		s.data.Global = list
	} else {
		s.data.Contexts[contextKey] = list
	}
	return s.saveLocked()
}

func (s *Store) SetDone(scope Scope, contextKey, id string, done bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var list []Todo
	switch scope {
	case ScopeGlobal:
		list = s.data.Global
	case ScopeContext:
		if contextKey == "" {
			return errors.New("context key required")
		}
		list = s.data.Contexts[contextKey]
	default:
		return fmt.Errorf("unknown scope %q", scope)
	}

	found := false
	now := time.Now().UTC()
	for i := range list {
		if list[i].ID == id {
			found = true
			list[i].Done = done
			if done {
				list[i].CompletedAt = &now
			} else {
				list[i].CompletedAt = nil
			}
			break
		}
	}
	if !found {
		return fmt.Errorf("todo id %q not found", id)
	}
	if scope == ScopeGlobal {
		s.data.Global = list
	} else {
		s.data.Contexts[contextKey] = list
	}
	return s.saveLocked()
}

func (s *Store) Delete(scope Scope, contextKey, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.listForScope(scope, contextKey)
	if err != nil {
		return err
	}

	idx := -1
	for i := range list {
		if list[i].ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("todo id %q not found", id)
	}

	targets := map[string]struct{}{id: {}}
	for {
		before := len(targets)
		for _, t := range list {
			if t.ParentID == "" {
				continue
			}
			if _, ok := targets[t.ParentID]; ok {
				targets[t.ID] = struct{}{}
			}
		}
		if len(targets) == before {
			break
		}
	}
	filtered := make([]Todo, 0, len(list))
	for _, t := range list {
		if _, drop := targets[t.ID]; drop {
			continue
		}
		filtered = append(filtered, t)
	}
	list = filtered
	s.setListForScope(scope, contextKey, list)
	return s.saveLocked()
}

func (s *Store) Move(scope Scope, contextKey, id string, targetScope Scope, targetContextKey, targetParentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	src, err := s.listForScope(scope, contextKey)
	if err != nil {
		return err
	}
	dst, err := s.listForScope(targetScope, targetContextKey)
	if err != nil {
		return err
	}

	targets := map[string]struct{}{}
	for _, t := range src {
		if t.ID == id {
			targets[t.ID] = struct{}{}
			break
		}
	}
	if len(targets) == 0 {
		return fmt.Errorf("todo id %q not found", id)
	}
	for {
		before := len(targets)
		for _, t := range src {
			if _, ok := targets[t.ParentID]; ok {
				targets[t.ID] = struct{}{}
			}
		}
		if before == len(targets) {
			break
		}
	}

	moved := make([]Todo, 0, len(targets))
	kept := make([]Todo, 0, len(src))
	for _, t := range src {
		if _, ok := targets[t.ID]; ok {
			moved = append(moved, t)
			continue
		}
		kept = append(kept, t)
	}
	for i := range moved {
		if moved[i].ID == id {
			moved[i].ParentID = targetParentID
		}
	}
	dst = append(dst, moved...)
	s.setListForScope(scope, contextKey, kept)
	s.setListForScope(targetScope, targetContextKey, dst)
	return s.saveLocked()
}

func (s *Store) Reparent(scope Scope, contextKey, id, newParentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	list, err := s.listForScope(scope, contextKey)
	if err != nil {
		return err
	}
	found := false
	for _, t := range list {
		if t.ID == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("todo id %q not found", id)
	}
	if err := validateReparent(list, id, newParentID); err != nil {
		return err
	}
	for i := range list {
		if list[i].ID == id {
			list[i].ParentID = newParentID
			break
		}
	}
	s.setListForScope(scope, contextKey, list)
	return s.saveLocked()
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.data = Data{
				Version:  2,
				Global:   []Todo{},
				Contexts: map[string][]Todo{},
				Meta:     map[string]MetaInfo{},
			}
			return nil
		}
		return err
	}
	var d Data
	if err := json.Unmarshal(b, &d); err != nil {
		return err
	}
	if d.Contexts == nil {
		d.Contexts = map[string][]Todo{}
	}
	if d.Meta == nil {
		d.Meta = map[string]MetaInfo{}
	}
	if d.Version == 0 {
		d.Version = 1
	}
	if d.Version < 2 {
		d.Version = 2
	}
	s.data = d
	return nil
}

func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return fileio.WriteFileAtomic(s.path, b, 0o644)
}

func newID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *Store) RemoveTag(tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tag = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(tag, "@")))
	if tag == "" {
		return nil
	}
	clean := func(list []Todo) []Todo {
		out := make([]Todo, len(list))
		copy(out, list)
		for i := range out {
			filtered := make([]string, 0, len(out[i].Tags))
			for _, tg := range out[i].Tags {
				v := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(tg, "@")))
				if v == tag || v == "" {
					continue
				}
				filtered = append(filtered, v)
			}
			out[i].Tags = filtered
		}
		return out
	}
	s.data.Global = clean(s.data.Global)
	for k, v := range s.data.Contexts {
		s.data.Contexts[k] = clean(v)
	}
	return s.saveLocked()
}

func (s *Store) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Global = []Todo{}
	s.data.Contexts = map[string][]Todo{}
	s.data.Meta = map[string]MetaInfo{}
	if s.data.Version < 2 {
		s.data.Version = 2
	}
	return s.saveLocked()
}

func (s *Store) listForScope(scope Scope, contextKey string) ([]Todo, error) {
	switch scope {
	case ScopeGlobal:
		return append([]Todo(nil), s.data.Global...), nil
	case ScopeContext:
		if contextKey == "" {
			return nil, errors.New("context key required")
		}
		return append([]Todo(nil), s.data.Contexts[contextKey]...), nil
	default:
		return nil, fmt.Errorf("unknown scope %q", scope)
	}
}

func (s *Store) setListForScope(scope Scope, contextKey string, list []Todo) {
	switch scope {
	case ScopeGlobal:
		s.data.Global = list
	case ScopeContext:
		if s.data.Contexts == nil {
			s.data.Contexts = map[string][]Todo{}
		}
		s.data.Contexts[contextKey] = list
	}
}

func validateReparent(list []Todo, id, newParentID string) error {
	if newParentID == id {
		return errors.New("todo cannot be its own parent")
	}
	if newParentID == "" {
		return nil
	}
	parents := map[string]string{}
	exists := map[string]struct{}{}
	for _, t := range list {
		parents[t.ID] = t.ParentID
		exists[t.ID] = struct{}{}
	}
	if _, ok := exists[newParentID]; !ok {
		return fmt.Errorf("parent id %q not found", newParentID)
	}
	cur := newParentID
	for cur != "" {
		if cur == id {
			return errors.New("reparent would create a cycle")
		}
		cur = parents[cur]
	}
	return nil
}
