#!/usr/bin/env bash

set -euo pipefail

cwd="${1:-$PWD}"
enabled="$(tmux show-option -gqv @tmux-todo-focus-alert)"
case "${enabled,,}" in
  on|true|1|yes) ;;
  *) exit 0 ;;
esac

bin="$(tmux show-option -gqv @tmux-todo-bin)"
if [ -z "$bin" ]; then
  bin="tmux-todo"
fi

context_switch_only="$(tmux show-option -gqv @tmux-todo-focus-on-context-switch)"
include_global="$(tmux show-option -gqv @tmux-todo-focus-include-global)"
cooldown_sec="$(tmux show-option -gqv @tmux-todo-focus-cooldown-sec)"
if [ -z "$cooldown_sec" ]; then
  cooldown_sec="0"
fi

key="$("$bin" --cwd "$cwd" context-key 2>/dev/null || echo "$cwd")"
safe_key="$(printf "%s" "$key" | tr '/|=:' '_' | tr -cd '[:alnum:]_.-')"
if [ -z "$safe_key" ]; then
  safe_key="default"
fi

if [ "${context_switch_only,,}" = "on" ] || [ "${context_switch_only,,}" = "true" ] || [ "${context_switch_only,,}" = "1" ] || [ "${context_switch_only,,}" = "yes" ]; then
  state_file="/tmp/tmux-todo-last-context"
  previous=""
  if [ -f "$state_file" ]; then
    previous="$(cat "$state_file" 2>/dev/null || true)"
  fi
  printf "%s" "$key" > "$state_file"
  if [ "$previous" = "$key" ]; then
    exit 0
  fi
fi

if [ "${include_global,,}" = "on" ] || [ "${include_global,,}" = "true" ] || [ "${include_global,,}" = "1" ] || [ "${include_global,,}" = "yes" ]; then
  if ! "$bin" --cwd "$cwd" has-high >/dev/null 2>&1; then
    exit 0
  fi
else
  if ! "$bin" --cwd "$cwd" has-high --context-only >/dev/null 2>&1; then
    exit 0
  fi
fi

stamp="/tmp/tmux-todo-focus-${safe_key}.stamp"
now="$(date +%s)"
last="0"
if [ -f "$stamp" ]; then
  last="$(cat "$stamp" 2>/dev/null || echo 0)"
fi

if [ $((now - last)) -lt "$cooldown_sec" ]; then
  exit 0
fi

printf "%s" "$now" > "$stamp"
"$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/popup.sh" peek-alert "$cwd"
