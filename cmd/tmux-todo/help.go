package main

import "fmt"

func printHelp() {
	fmt.Println(`tmux-todo - context-aware todo manager for tmux

Usage:
  tmux-todo [global flags] <command> [command flags]

Global flags:
  --cwd <dir>               working directory used for git context detection
  --data <path>             todo data file path
  --strikethrough <bool>    render completed todo text with strikethrough
  --peek-duration-ms <ms>   auto-close duration for peek/peek-high

Commands:
  tui                       open full TUI popup
  quick                     open minimal quick-add input
  peek                      open compact peek popup
  peek-high                 open compact high-priority alert view
  add                       add a todo
  list                      list todos
  get                       get one todo by id
  done                      mark todo done
  undone                    mark todo not done
  edit                      edit todo fields
  delete                    delete todo (cascades children)
  move                      move todo subtree across scopes/contexts
  reparent                  change parent id
  has-high                  detect open high-priority tasks
  summary                   print context/global summary counters
  doctor                    validate paths and local setup
  export                    export data/config snapshot
  clear-all                 clear all tasks (requires --yes)
  context-key               print current context key
  tags list|add|remove      manage per-user tag registry
  help                      show this help

Use --json on CLI mutation/read commands for machine-readable output.`)
}
