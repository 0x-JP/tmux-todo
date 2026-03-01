package gitctx

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNotGitRepo = errors.New("not in git repository")

type Context struct {
	RepoRoot     string
	WorktreeRoot string
	Branch       string
}

func (c Context) Key() string {
	if !c.IsGit() {
		return "global"
	}
	return "repo=" + c.RepoRoot + "|wt=" + c.WorktreeRoot + "|br=" + c.Branch
}

func (c Context) IsGit() bool {
	return c.RepoRoot != "" && c.WorktreeRoot != "" && c.Branch != "" && c.Branch != "no-git"
}

func (c Context) Label() string {
	if !c.IsGit() {
		return "Global"
	}
	if c.RepoRoot == "" {
		return c.WorktreeRoot + " [" + c.Branch + "]"
	}
	repo := filepath.Base(c.RepoRoot)
	wt := filepath.Base(c.WorktreeRoot)
	return repo + "/" + wt + " [" + c.Branch + "]"
}

type Runner interface {
	Run(dir string, args ...string) (string, error)
}

type RealRunner struct{}

func (RealRunner) Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func Detect(cwd string, runner Runner) (Context, error) {
	if _, err := runner.Run(cwd, "rev-parse", "--is-inside-work-tree"); err != nil {
		return Context{}, ErrNotGitRepo
	}
	worktreeRoot, err := runner.Run(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return Context{}, err
	}
	commonDir, err := runner.Run(cwd, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return Context{}, err
	}
	branch, err := runner.Run(cwd, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		branch, err = runner.Run(cwd, "rev-parse", "--short", "HEAD")
		if err != nil {
			return Context{}, err
		}
		branch = "detached:" + branch
	}
	repoRoot := filepath.Dir(commonDir)
	return Context{
		RepoRoot:     filepath.Clean(repoRoot),
		WorktreeRoot: filepath.Clean(worktreeRoot),
		Branch:       branch,
	}, nil
}
