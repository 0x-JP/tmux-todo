#!/usr/bin/env bash

set -euo pipefail

mode="${1:-full}"
cwd="${2:-$PWD}"

bin="$(tmux show-option -gqv @tmux-todo-bin)"
if [ -z "$bin" ]; then
  bin="tmux-todo"
fi

popup_width="$(tmux show-option -gqv @tmux-todo-popup-width)"
popup_height="$(tmux show-option -gqv @tmux-todo-popup-height)"
peek_width="$(tmux show-option -gqv @tmux-todo-peek-width)"
peek_height="$(tmux show-option -gqv @tmux-todo-peek-height)"
focus_width="$(tmux show-option -gqv @tmux-todo-focus-width)"
focus_height="$(tmux show-option -gqv @tmux-todo-focus-height)"
strike_opt="$(tmux show-option -gqv @tmux-todo-strikethrough)"
alert_duration_ms="$(tmux show-option -gqv @tmux-todo-alert-duration-ms)"
focus_duration_ms="$(tmux show-option -gqv @tmux-todo-focus-duration-ms)"
strike_flag="true"
case "${strike_opt,,}" in
  off|false|0|no)
    strike_flag="false"
    ;;
esac

escape_sq() {
  printf "%s" "$1" | sed "s/'/'\\\\''/g"
}

cwd_esc="$(escape_sq "$cwd")"
bin_esc="$(escape_sq "$bin")"

if [ "$mode" = "peek" ]; then
  tmux display-popup \
    -E \
    -w "${peek_width:-34%}" \
    -h "${peek_height:-22%}" \
    -x R \
    -y 1 \
    "'$bin_esc' --cwd '$cwd_esc' --strikethrough=$strike_flag --peek-duration-ms=${alert_duration_ms:-5000} peek"
  exit 0
fi

if [ "$mode" = "peek-alert" ]; then
  tmux display-popup \
    -E \
    -w "${focus_width:-32%}" \
    -h "${focus_height:-14%}" \
    -x R \
    -y 1 \
    "'$bin_esc' --cwd '$cwd_esc' --peek-duration-ms=${focus_duration_ms:-2000} peek-high"
  exit 0
fi

tmux display-popup \
  -E \
  -w "${popup_width:-80%}" \
  -h "${popup_height:-80%}" \
  -x C \
  -y C \
  "'$bin_esc' --cwd '$cwd_esc' --strikethrough=$strike_flag tui"
