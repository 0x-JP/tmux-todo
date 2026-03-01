package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jp/tmux-todo/internal/gitctx"
)

func detectContext(cwd string) (gitctx.Context, error) {
	ctx, err := gitctx.Detect(cwd, gitctx.RealRunner{})
	if err != nil {
		if !errors.Is(err, gitctx.ErrNotGitRepo) {
			return gitctx.Context{}, err
		}
		return gitctx.Context{
			RepoRoot:     "",
			WorktreeRoot: "",
			Branch:       "global",
		}, nil
	}
	return ctx, nil
}

func defaultScopeForContext(ctx gitctx.Context) string {
	if ctx.IsGit() {
		return "context"
	}
	return "global"
}

func defaultDataPath() (string, error) {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "tmux-todo", "todos.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", "tmux-todo", "todos.json"), nil
}

func defaultConfigPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tmux-todo", "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tmux-todo", "config.json"), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "tmux-todo:", err)
	os.Exit(1)
}
