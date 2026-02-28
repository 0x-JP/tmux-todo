#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

tmux set-option -gq @tmux-todo-bin "tmux-todo"
tmux set-option -gq @tmux-todo-popup-width "80%"
tmux set-option -gq @tmux-todo-popup-height "80%"
tmux set-option -gq @tmux-todo-peek-width "34%"
tmux set-option -gq @tmux-todo-peek-height "22%"
tmux set-option -gq @tmux-todo-focus-width "32%"
tmux set-option -gq @tmux-todo-focus-height "14%"
tmux set-option -gq @tmux-todo-strikethrough "on"
tmux set-option -gq @tmux-todo-focus-alert "off"
tmux set-option -gq @tmux-todo-focus-cooldown-sec "0"
tmux set-option -gq @tmux-todo-alert-duration-ms "5000"
tmux set-option -gq @tmux-todo-focus-duration-ms "2000"
tmux set-option -gq @tmux-todo-focus-on-context-switch "on"
tmux set-option -gq @tmux-todo-focus-include-global "off"
tmux set-option -gq @tmux-todo-bind-full ""
tmux set-option -gq @tmux-todo-bind-peek ""

bind_full="$(tmux show-option -gqv @tmux-todo-bind-full)"
bind_peek="$(tmux show-option -gqv @tmux-todo-bind-peek)"

if [ -n "$bind_full" ]; then
  tmux bind-key "$bind_full" run-shell "$CURRENT_DIR/scripts/popup.sh full '#{pane_current_path}'"
fi
if [ -n "$bind_peek" ]; then
  tmux bind-key "$bind_peek" run-shell "$CURRENT_DIR/scripts/popup.sh peek '#{pane_current_path}'"
fi

focus_alert="$(tmux show-option -gqv @tmux-todo-focus-alert)"
case "${focus_alert,,}" in
  on|true|1|yes)
    hook_cmd="run-shell '$CURRENT_DIR/scripts/focus-alert.sh \"#{pane_current_path}\"'"
    if ! tmux show-hooks -g pane-focus-in 2>/dev/null | grep -Fq "$CURRENT_DIR/scripts/focus-alert.sh"; then
      tmux set-hook -ag pane-focus-in "$hook_cmd"
    fi
    ;;
esac
