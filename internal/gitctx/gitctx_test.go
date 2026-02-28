package gitctx

import (
	"errors"
	"testing"
)

type fakeRunner struct {
	out map[string]string
	err map[string]error
}

func (f fakeRunner) Run(_ string, args ...string) (string, error) {
	key := ""
	for i, a := range args {
		if i > 0 {
			key += " "
		}
		key += a
	}
	if e, ok := f.err[key]; ok {
		return "", e
	}
	return f.out[key], nil
}

func TestDetectWorktreeContext(t *testing.T) {
	r := fakeRunner{
		out: map[string]string{
			"rev-parse --is-inside-work-tree":                   "true",
			"rev-parse --show-toplevel":                         "/ws/repo-feature",
			"rev-parse --path-format=absolute --git-common-dir": "/src/repo/.git",
			"symbolic-ref --short HEAD":                         "feature/abc",
		},
		err: map[string]error{},
	}
	ctx, err := Detect("/ws/repo-feature", r)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.RepoRoot != "/src/repo" {
		t.Fatalf("repo root = %q, want /src/repo", ctx.RepoRoot)
	}
	if ctx.WorktreeRoot != "/ws/repo-feature" {
		t.Fatalf("worktree root = %q", ctx.WorktreeRoot)
	}
	if ctx.Branch != "feature/abc" {
		t.Fatalf("branch = %q", ctx.Branch)
	}
}

func TestDetectDetachedHead(t *testing.T) {
	r := fakeRunner{
		out: map[string]string{
			"rev-parse --is-inside-work-tree":                   "true",
			"rev-parse --show-toplevel":                         "/repo",
			"rev-parse --path-format=absolute --git-common-dir": "/repo/.git",
			"rev-parse --short HEAD":                            "a1b2c3d",
		},
		err: map[string]error{
			"symbolic-ref --short HEAD": errors.New("detached"),
		},
	}
	ctx, err := Detect("/repo", r)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Branch != "detached:a1b2c3d" {
		t.Fatalf("branch = %q", ctx.Branch)
	}
}

func TestDetectNotGit(t *testing.T) {
	r := fakeRunner{
		out: map[string]string{},
		err: map[string]error{
			"rev-parse --is-inside-work-tree": errors.New("no"),
		},
	}
	_, err := Detect("/tmp", r)
	if !errors.Is(err, ErrNotGitRepo) {
		t.Fatalf("err = %v, want ErrNotGitRepo", err)
	}
}
