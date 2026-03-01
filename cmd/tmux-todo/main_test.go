package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jp/tmux-todo/internal/config"
	"github.com/jp/tmux-todo/internal/gitctx"
	"github.com/jp/tmux-todo/internal/store"
)

func TestCLIJSONFlows(t *testing.T) {
	dir := t.TempDir()
	st, err := store.New(filepath.Join(dir, "todos.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.New(filepath.Join(dir, "config.json"), store.DefaultTags)
	if err != nil {
		t.Fatal(err)
	}
	ctx := gitctx.Context{RepoRoot: "/repo", WorktreeRoot: "/repo/wt", Branch: "feat"}

	out, err := captureStdout(func() error {
		return runAdd(st, ctx, []string{"--text", "json task", "--priority", "high", "--json"})
	})
	if err != nil {
		t.Fatal(err)
	}
	var add map[string]any
	if err := json.Unmarshal([]byte(out), &add); err != nil {
		t.Fatal(err)
	}
	if add["action"] != "add" {
		t.Fatalf("unexpected add action: %#v", add["action"])
	}
	todoMap := add["todo"].(map[string]any)
	id := todoMap["id"].(string)

	out, err = captureStdout(func() error {
		return runGet(st, ctx, []string{"--id", id, "--json"})
	})
	if err != nil {
		t.Fatal(err)
	}
	var get map[string]any
	if err := json.Unmarshal([]byte(out), &get); err != nil {
		t.Fatal(err)
	}
	if get["action"] != "get" {
		t.Fatalf("unexpected get action: %#v", get["action"])
	}

	out, err = captureStdout(func() error {
		return runList(st, ctx, []string{"--scope", "context", "--json"})
	})
	if err != nil {
		t.Fatal(err)
	}
	var list map[string]any
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		t.Fatal(err)
	}
	scopes, ok := list["scopes"].([]any)
	if !ok || len(scopes) == 0 {
		t.Fatalf("missing scopes in list output: %#v", list)
	}

	out, err = captureStdout(func() error {
		return runTags(st, cfg, []string{"add", "--tag", "whatever", "--json"})
	})
	if err != nil {
		t.Fatal(err)
	}
	var tagsAdd map[string]any
	if err := json.Unmarshal([]byte(out), &tagsAdd); err != nil {
		t.Fatal(err)
	}
	if tagsAdd["action"] != "tags.add" {
		t.Fatalf("unexpected tags.add action: %#v", tagsAdd["action"])
	}

	out, err = captureStdout(func() error {
		return runHasHigh(st, ctx, []string{"--json"})
	})
	if err != nil {
		t.Fatal(err)
	}
	var hh map[string]any
	if err := json.Unmarshal([]byte(out), &hh); err != nil {
		t.Fatal(err)
	}
	if _, ok := hh["has_high"]; !ok {
		t.Fatalf("missing has_high in output: %#v", hh)
	}
}

func captureStdout(fn func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	var b bytes.Buffer
	_, _ = io.Copy(&b, r)
	_ = r.Close()
	return b.String(), runErr
}
