package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyMapCompleteness(t *testing.T) {
	km := DefaultKeyMap()

	// Verify all bindings have at least one key assigned.
	bindings := map[string]key.Binding{
		"Global.Quit":         km.Global.Quit,
		"Global.Help":         km.Global.Help,
		"Global.Filter":       km.Global.Filter,
		"Main.CycleScope":     km.Main.CycleScope,
		"Main.PrevContext":    km.Main.PrevContext,
		"Main.NextContext":    km.Main.NextContext,
		"Main.MoveUp":         km.Main.MoveUp,
		"Main.MoveDown":       km.Main.MoveDown,
		"Main.QuickAdd":       km.Main.QuickAdd,
		"Main.GuidedAdd":      km.Main.GuidedAdd,
		"Main.AddChild":       km.Main.AddChild,
		"Main.Edit":           km.Main.Edit,
		"Main.PriorityHigh":   km.Main.PriorityHigh,
		"Main.PriorityMed":    km.Main.PriorityMed,
		"Main.PriorityLow":    km.Main.PriorityLow,
		"Main.PriorityClear":  km.Main.PriorityClear,
		"Main.ToggleBlocked":  km.Main.ToggleBlocked,
		"Main.ToggleReview":   km.Main.ToggleReview,
		"Main.TagPicker":      km.Main.TagPicker,
		"Main.TagManager":     km.Main.TagManager,
		"Main.ToggleDone":     km.Main.ToggleDone,
		"Main.Delete":         km.Main.Delete,
		"TagPicker.Close":     km.TagPicker.Close,
		"TagPicker.MoveUp":    km.TagPicker.MoveUp,
		"TagPicker.MoveDown":  km.TagPicker.MoveDown,
		"TagPicker.Toggle":    km.TagPicker.Toggle,
		"TagPicker.DeleteTag": km.TagPicker.DeleteTag,
		"TagPicker.NewTag":    km.TagPicker.NewTag,
		"GuidedAdd.Cancel":    km.GuidedAdd.Cancel,
		"GuidedAdd.Confirm":   km.GuidedAdd.Confirm,
		"Filter.Cancel":       km.Filter.Cancel,
		"Filter.Apply":        km.Filter.Apply,
		"QuickAdd.Cancel":     km.QuickAdd.Cancel,
		"QuickAdd.Save":       km.QuickAdd.Save,
		"Help.Close":          km.Help.Close,
	}
	for name, b := range bindings {
		if len(b.Keys()) == 0 {
			t.Errorf("binding %s has no keys", name)
		}
		h := b.Help()
		if h.Key == "" || h.Desc == "" {
			t.Errorf("binding %s missing help text: key=%q desc=%q", name, h.Key, h.Desc)
		}
	}
}

func TestApplyOverrides(t *testing.T) {
	km := DefaultKeyMap()
	overrides := map[string][]string{
		"delete":  {"x"},
		"move_up": {"up", "w"},
		"quit":    {"q", "ctrl+c", "ctrl+q"},
	}
	km = ApplyOverrides(km, overrides)

	if keys := km.Main.Delete.Keys(); len(keys) != 1 || keys[0] != "x" {
		t.Fatalf("delete keys = %v, want [x]", keys)
	}
	if keys := km.Main.MoveUp.Keys(); len(keys) != 2 || keys[0] != "up" || keys[1] != "w" {
		t.Fatalf("move_up keys = %v, want [up w]", keys)
	}
	if keys := km.Global.Quit.Keys(); len(keys) != 3 {
		t.Fatalf("quit keys = %v, want 3 keys", keys)
	}
}

func TestApplyOverridesSyncsTagPickerNav(t *testing.T) {
	km := DefaultKeyMap()
	overrides := map[string][]string{
		"move_up":   {"up", "w"},
		"move_down": {"down", "s"},
	}
	km = ApplyOverrides(km, overrides)

	// Tag picker nav should mirror main nav after overrides.
	if keys := km.TagPicker.MoveUp.Keys(); len(keys) != 2 || keys[0] != "up" || keys[1] != "w" {
		t.Fatalf("tag picker move_up = %v, want [up w]", keys)
	}
	if keys := km.TagPicker.MoveDown.Keys(); len(keys) != 2 || keys[0] != "down" || keys[1] != "s" {
		t.Fatalf("tag picker move_down = %v, want [down s]", keys)
	}
}

func TestApplyOverridesIgnoresUnknownKeys(t *testing.T) {
	km := DefaultKeyMap()
	original := km.Main.Delete.Keys()
	overrides := map[string][]string{
		"nonexistent_binding": {"x", "y"},
	}
	km = ApplyOverrides(km, overrides)

	// Delete should be unchanged.
	if keys := km.Main.Delete.Keys(); len(keys) != len(original) || keys[0] != original[0] {
		t.Fatalf("delete keys changed unexpectedly: %v", keys)
	}
}

func TestMainHelpKeyMapInterfaces(t *testing.T) {
	km := DefaultKeyMap()
	hkm := mainHelpKeyMap{km: km}

	short := hkm.ShortHelp()
	if len(short) == 0 {
		t.Fatal("ShortHelp returned empty slice")
	}

	full := hkm.FullHelp()
	if len(full) == 0 {
		t.Fatal("FullHelp returned empty slice")
	}
	for i, group := range full {
		if len(group) == 0 {
			t.Errorf("FullHelp group %d is empty", i)
		}
	}
}
