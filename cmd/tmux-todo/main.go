package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jp/tmux-todo/internal/config"
	"github.com/jp/tmux-todo/internal/store"
	"github.com/jp/tmux-todo/internal/ui"
)

func main() {
	root := flag.NewFlagSet("tmux-todo", flag.ExitOnError)
	var cwd string
	var dataPath string
	var strikethrough bool
	var peekDurationMS int
	root.StringVar(&cwd, "cwd", "", "working directory used for git context detection")
	root.StringVar(&dataPath, "data", "", "path to todo data file")
	root.BoolVar(&strikethrough, "strikethrough", true, "render completed todo text with strikethrough")
	root.IntVar(&peekDurationMS, "peek-duration-ms", 5000, "auto-close duration for peek mode")
	_ = root.Parse(os.Args[1:])

	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			fatal(err)
		}
	}
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		fatal(err)
	}
	if dataPath == "" {
		dataPath, err = defaultDataPath()
		if err != nil {
			fatal(err)
		}
	}
	cfgPath, err := defaultConfigPath()
	if err != nil {
		fatal(err)
	}
	st, err := store.New(dataPath)
	if err != nil {
		fatal(err)
	}
	cfg, err := config.New(cfgPath, store.DefaultTags)
	if err != nil {
		fatal(err)
	}
	ctx, err := detectContext(absCWD)
	if err != nil {
		fatal(err)
	}
	if ctx.IsGit() {
		_ = st.SetContextMeta(ctx.Key(), store.MetaInfo{
			RepoRoot:     ctx.RepoRoot,
			WorktreeRoot: ctx.WorktreeRoot,
			Branch:       ctx.Branch,
		})
	}

	args := root.Args()
	cmd := "tui"
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "help":
		printHelp()
	case "tui":
		m := ui.NewMainModel(st, cfg, ctx, strikethrough)
		if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
			fatal(err)
		}
	case "peek":
		m := ui.NewPeekModel(st, ctx, strikethrough, time.Duration(peekDurationMS)*time.Millisecond)
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fatal(err)
		}
	case "quick":
		m := ui.NewQuickAddModel(st, ctx)
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fatal(err)
		}
	case "peek-high":
		m := ui.NewHighPeekModel(st, ctx, time.Duration(peekDurationMS)*time.Millisecond)
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fatal(err)
		}
	case "add":
		if err := runAdd(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "list":
		if err := runList(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "done":
		if err := runDone(st, ctx, args[1:], true); err != nil {
			fatal(err)
		}
	case "undone":
		if err := runDone(st, ctx, args[1:], false); err != nil {
			fatal(err)
		}
	case "edit":
		if err := runEdit(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "delete":
		if err := runDelete(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "move":
		if err := runMove(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "reparent":
		if err := runReparent(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "get":
		if err := runGet(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "has-high":
		if err := runHasHigh(st, ctx, args[1:]); err != nil {
			fatal(err)
		}
	case "tags":
		if err := runTags(st, cfg, args[1:]); err != nil {
			fatal(err)
		}
	case "context-key":
		fmt.Println(ctx.Key())
	default:
		fatal(fmt.Errorf("unknown command %q (run `tmux-todo help`)", cmd))
	}
}
