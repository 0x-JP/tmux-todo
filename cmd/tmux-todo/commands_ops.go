package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jp/tmux-todo/internal/config"
	"github.com/jp/tmux-todo/internal/gitctx"
	"github.com/jp/tmux-todo/internal/store"
)

func runSummary(st *store.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("summary", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var contextKey string
	var windowHours int
	var jsonOut bool
	fs.StringVar(&contextKey, "context-key", "", "override context key")
	fs.IntVar(&windowHours, "window-hours", 24, "recent done window in hours")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if contextKey == "" {
		contextKey = ctx.Key()
	}
	if windowHours < 1 {
		windowHours = 1
	}
	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(windowHours) * time.Hour)
	d := st.Snapshot()

	count := func(list []store.Todo) map[string]int {
		open := 0
		high := 0
		recentDone := 0
		for _, t := range list {
			if !t.Done {
				open++
				if t.Priority == store.PriorityHigh {
					high++
				}
				continue
			}
			if t.CompletedAt != nil && t.CompletedAt.After(cutoff) {
				recentDone++
			}
		}
		return map[string]int{
			"open":        open,
			"high_open":   high,
			"recent_done": recentDone,
		}
	}

	ctxCounts := count(d.Contexts[contextKey])
	globalCounts := count(d.Global)
	combined := map[string]int{
		"open":        ctxCounts["open"] + globalCounts["open"],
		"high_open":   ctxCounts["high_open"] + globalCounts["high_open"],
		"recent_done": ctxCounts["recent_done"] + globalCounts["recent_done"],
	}
	payload := map[string]any{
		"ok":           true,
		"action":       "summary",
		"context_key":  contextKey,
		"window_hours": windowHours,
		"context":      ctxCounts,
		"global":       globalCounts,
		"combined":     combined,
	}
	if jsonOut {
		return printJSON(payload)
	}
	fmt.Printf("Summary (%dh)\n", windowHours)
	fmt.Printf("Context: open=%d high=%d recent_done=%d\n", ctxCounts["open"], ctxCounts["high_open"], ctxCounts["recent_done"])
	fmt.Printf("Global : open=%d high=%d recent_done=%d\n", globalCounts["open"], globalCounts["high_open"], globalCounts["recent_done"])
	fmt.Printf("Total  : open=%d high=%d recent_done=%d\n", combined["open"], combined["high_open"], combined["recent_done"])
	return nil
}

func runDoctor(st *store.Store, cfg *config.Store, ctx gitctx.Context, dataPath, cfgPath string, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var jsonOut bool
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_ = st.Snapshot()
	_ = cfg.Tags()

	warnings := []string{}
	checkWritable := func(path string) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			warnings = append(warnings, fmt.Sprintf("cannot create directory %s: %v", dir, err))
			return
		}
		f, err := os.CreateTemp(dir, ".tmux-todo-writecheck-*")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("directory not writable %s: %v", dir, err))
			return
		}
		_ = f.Close()
		_ = os.Remove(f.Name())
	}
	checkReadable := func(path string) {
		if _, err := os.Stat(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			warnings = append(warnings, fmt.Sprintf("path not accessible %s: %v", path, err))
		}
	}
	checkReadable(dataPath)
	checkReadable(cfgPath)
	checkWritable(dataPath)
	checkWritable(cfgPath)

	exe, err := os.Executable()
	if err != nil {
		warnings = append(warnings, "failed to resolve executable path: "+err.Error())
	}
	payload := map[string]any{
		"ok":          len(warnings) == 0,
		"action":      "doctor",
		"data_path":   dataPath,
		"config_path": cfgPath,
		"context_key": ctx.Key(),
		"is_git":      ctx.IsGit(),
		"warnings":    warnings,
		"executable":  exe,
	}
	if jsonOut {
		return printJSON(payload)
	}
	fmt.Println("tmux-todo doctor")
	fmt.Println("data:", dataPath)
	fmt.Println("config:", cfgPath)
	fmt.Println("context:", ctx.Key())
	fmt.Println("executable:", exe)
	if len(warnings) == 0 {
		fmt.Println("status: OK")
		return nil
	}
	fmt.Println("status: WARN")
	for _, w := range warnings {
		fmt.Println("- " + w)
	}
	return nil
}

func runExport(st *store.Store, cfg *config.Store, ctx gitctx.Context, args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var outPath string
	var jsonOut bool
	fs.StringVar(&outPath, "out", "", "output path (default stdout)")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	payload := map[string]any{
		"version":     1,
		"exported_at": time.Now().UTC(),
		"context_key": ctx.Key(),
		"data":        st.Snapshot(),
		"config": map[string]any{
			"tags": cfg.Tags(),
			"ui":   cfg.UI(),
		},
	}
	if outPath == "" {
		return printJSON(payload)
	}
	b, err := jsonMarshalIndent(payload)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{"ok": true, "action": "export", "out": outPath})
	}
	fmt.Println("exported", outPath)
	return nil
}

func runClearAll(st *store.Store, args []string) error {
	fs := flag.NewFlagSet("clear-all", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var yes bool
	var jsonOut bool
	fs.BoolVar(&yes, "yes", false, "confirm destructive clear")
	fs.BoolVar(&jsonOut, "json", false, "print machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !yes {
		return errors.New("refusing to clear tasks without --yes")
	}
	if err := st.Reset(); err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{
			"ok":     true,
			"action": "clear-all",
		})
	}
	fmt.Println("cleared all tasks")
	return nil
}
