# tmux-todo

Context-aware tmux todo manager with:

- Full popup manager (`tui`)
- Compact peek popup (`peek`)
- Focus alert popup for high-priority context tasks (`peek-high`)
- Global + context-scoped tasks
- Nested tasks
- Priority + tags
- AI-friendly CLI with JSON output

## Install

### Prerequisites

- tmux with popup support (`display-popup`, tmux 3.2+ recommended)
- Go 1.22+

Important:

- TPM installs/sources tmux plugin scripts.
- TPM does **not** build or install the `tmux-todo` Go binary.
- You must install the binary separately (choose one method below).

### Build binary

```bash
cd /path/to/tmux-todo
go build -o ~/.local/bin/tmux-todo ./cmd/tmux-todo
```

Make sure `~/.local/bin` is on PATH for tmux sessions, or set `@tmux-todo-bin` explicitly in `.tmux.conf`.

### Install binary with `go install` (recommended for users)

```bash
go install github.com/0x-JP/tmux-todo/cmd/tmux-todo@latest
```

If your `GOBIN`/`GOPATH/bin` is not on PATH in tmux, set `@tmux-todo-bin` to the absolute binary path.

### Install from release binary (recommended for team rollout)

When release assets are available, download the binary and place it on PATH (for example `~/.local/bin/tmux-todo`).

### TPM install

Add to `~/.tmux.conf`:

```tmux
set -g @plugin '0x-JP/tmux-todo'
```

For local development with TPM:

```bash
mkdir -p ~/.tmux/plugins
ln -sfn /path/to/tmux-todo ~/.tmux/plugins/tmux-todo
```

Then in `~/.tmux.conf`:

```tmux
set -g @plugin 'tmux-todo'
```

Keep TPM loader at bottom:

```tmux
run '~/.tmux/plugins/tpm/tpm'
```

Reload tmux:

```bash
tmux source-file ~/.tmux.conf
```

## Minimal `.tmux.conf` Setup

```tmux
# binary location
set -g @tmux-todo-bin "$HOME/.local/bin/tmux-todo"

# optional popup sizing
set -g @tmux-todo-popup-width "80%"
set -g @tmux-todo-popup-height "80%"
set -g @tmux-todo-peek-width "34%"
set -g @tmux-todo-peek-height "22%"
set -g @tmux-todo-quick-width "46%"
set -g @tmux-todo-quick-height "18%"
set -g @tmux-todo-focus-width "24%"
set -g @tmux-todo-focus-height "14%"

# focus alert behavior
set -g @tmux-todo-focus-alert "on"
set -g @tmux-todo-focus-on-context-switch "on"
set -g @tmux-todo-focus-include-global "off"
set -g @tmux-todo-focus-cooldown-sec "0"
set -g @tmux-todo-focus-duration-ms "2000"

# optional manual keybindings
bind-key T run-shell "~/.tmux/plugins/tmux-todo/scripts/popup.sh full '#{pane_current_path}'"
bind-key t run-shell "~/.tmux/plugins/tmux-todo/scripts/popup.sh peek '#{pane_current_path}'"
bind-key C-t run-shell "~/.tmux/plugins/tmux-todo/scripts/popup.sh quick '#{pane_current_path}'"
```

## Plugin Options

- `@tmux-todo-bin` default: `tmux-todo`
- `@tmux-todo-popup-width` default: `80%`
- `@tmux-todo-popup-height` default: `80%`
- `@tmux-todo-peek-width` default: `34%`
- `@tmux-todo-peek-height` default: `22%`
- `@tmux-todo-quick-width` default: `46%`
- `@tmux-todo-quick-height` default: `18%`
- `@tmux-todo-focus-width` default: `32%`
- `@tmux-todo-focus-height` default: `14%`
- `@tmux-todo-strikethrough` default: `on`
- `@tmux-todo-focus-alert` default: `off`
- `@tmux-todo-focus-on-context-switch` default: `on`
- `@tmux-todo-focus-include-global` default: `off`
- `@tmux-todo-focus-cooldown-sec` default: `0`
- `@tmux-todo-focus-duration-ms` default: `2000`
- `@tmux-todo-alert-duration-ms` default: `5000`
- `@tmux-todo-bind-full` default: empty
- `@tmux-todo-bind-peek` default: empty
- `@tmux-todo-bind-quick` default: `C-t`

## TUI Keys

- `?` help overlay
- `tab` cycle scope (context/global/all-contexts)
- `/` open filter input (`p:high tag:blocked`)
- `a` quick add (text-only)
- `A` guided add
- `[` previous context
- `]` next context
- `c` add child task
- `e` edit selected task
- `g` task tag picker for selected task (inline)
- `G` global tag manager (full-page)
- `1/2/3` set selected task priority (`1=high`, `2=med`, `3=low`)
- `!` clear selected task priority
- `b/r` toggle `blocked/review` on selected task
- `space` toggle done
- `d` delete selected task
- `j/k` or arrows move cursor
- `q` quit

## Quick Add Popup (`Ctrl-t`)

Use a tmux binding to open a small quick-add input popup:

```tmux
set -g @tmux-todo-bind-quick "C-t"
```

`@tmux-todo-bind-quick` is bound as a no-prefix key (`bind-key -n`), so pressing `Ctrl-t` directly opens quick add.

Input grammar:

- `task 1` -> add to current context
- `global | task 1` -> add to Global context
- `task 1 | p=1` -> add high-priority in current context
- `task 1 | p=high` -> same as above
- `global | task 1 | p=2` -> add Global medium-priority task
- `task 1 | t=blocked,review` -> add tags inline

The full TUI now restores your last main view and selected task on reopen.

## CLI Reference

Tip: run `tmux-todo help` for built-in command help.

### Add

```bash
tmux-todo add --text "Investigate flaky test"
tmux-todo add --scope global --text "Book travel"
tmux-todo add --text "Child task" --parent <PARENT_ID>
tmux-todo add --text "Fix CI" --priority high --tag blocked --tag review
```

### List

```bash
tmux-todo list
tmux-todo list --scope global
tmux-todo list --scope all --all
tmux-todo list --priority high
tmux-todo list --tag blocked --sort created
```

### State Changes

```bash
tmux-todo done --id <TODO_ID>
tmux-todo undone --id <TODO_ID>
tmux-todo edit --id <TODO_ID> --text "Updated" --priority med --tags blocked,review
tmux-todo delete --id <TODO_ID>
tmux-todo move --id <TODO_ID> --from-scope context --to-scope global
tmux-todo reparent --id <TODO_ID> --parent <PARENT_ID>
```

### Read Helpers

```bash
tmux-todo get --id <TODO_ID>
tmux-todo get --scope global --id <TODO_ID>
tmux-todo context-key
tmux-todo has-high
tmux-todo has-high --context-only
tmux-todo summary --json
tmux-todo doctor --json
tmux-todo export --out ~/tmux-todo-export.json --json
tmux-todo clear-all --yes --json
```

`clear-all` is destructive and requires `--yes`.

### Tag Registry (Per User)

```bash
tmux-todo tags list
tmux-todo tags add review
tmux-todo tags remove whatever --global-clean
```

`tags remove --global-clean` removes the tag from:

- your local tag registry
- all existing todos in local state

### JSON Output (for Agents)

Use `--json` on:

- `add`, `list`, `get`, `done`, `undone`, `edit`, `delete`, `move`, `reparent`, `has-high`
- `tags list`, `tags add`, `tags remove`

Examples:

```bash
tmux-todo add --text "fix flaky test" --priority high --json
tmux-todo list --scope all --all --json
tmux-todo get --id <TODO_ID> --json
tmux-todo done --id <TODO_ID> --json
tmux-todo tags list --json
```

## Claude Code Global Setup (Recommended)

For teams using Claude Code, keep this user-scoped file in `~/.claude/` so behavior is global across repos/sessions.

Create:

```bash
mkdir -p ~/.claude
cat > ~/.claude/CLAUDE.md <<'EOF'
When managing my tasks, use tmux-todo CLI (never edit todo files directly).

Workflow:
1. Detect current context:
   - tmux-todo context-key
2. Add tasks as work progresses:
   - tmux-todo add --text "<task>" [--priority high|med|low] [--tag <tag>] --json
3. Update status:
   - tmux-todo done --id <id> --json
   - tmux-todo undone --id <id> --json
   - tmux-todo edit --id <id> ... --json
4. Read context tasks during planning:
   - tmux-todo list --scope context --json
5. Prefer canonical tags:
   - blocked, review
   - tmux-todo tags list --json

Rules:
- Do not modify repo-local CLAUDE.md for task logging.
- Keep task updates in current context unless explicitly told otherwise.
- Prefer --json for machine-readable operations.
EOF
```

Why this works with parallel sessions: each Claude Code session has its own cwd/worktree, and tmux-todo scopes tasks by that context automatically.

## Optional: Codex User-Scoped Setup

If you also use Codex globally, keep a parallel file and instruct Codex to follow it.

```bash
mkdir -p ~/.config/tmux-todo
cat > ~/.config/tmux-todo/AGENT_TASKS.md <<'EOF'
Use tmux-todo CLI for task management, scoped to current context.
Prefer --json for all reads/writes.
EOF
```

Practical agent examples:

```bash
# create task in current context
tmux-todo add --text "Refactor parser edge-case handling" --priority high --tag review --json

# list active context tasks
tmux-todo list --scope context --json

# finish a task
tmux-todo done --id <TODO_ID> --json
```

## Data and Config Paths

Todo data:

- `$XDG_STATE_HOME/tmux-todo/todos.json`
- fallback: `~/.local/state/tmux-todo/todos.json`

User config (tag registry):

- `$XDG_CONFIG_HOME/tmux-todo/config.json`
- fallback: `~/.config/tmux-todo/config.json`

Important:

- `config.json` does **not** control where `todos.json` is stored.
- `config.json` stores user settings (tags + UI state), not todo data path.

You can override todo data path:

```bash
tmux-todo --data /path/to/todos.json tui
```

Path resolution priority for todo data:

1. `--data /custom/path.json`
2. `$XDG_STATE_HOME/tmux-todo/todos.json`
3. `~/.local/state/tmux-todo/todos.json`

tmux popup modes (`T`, `t`, `Ctrl-t`) use the environment of your tmux server.
If you want a custom default location without passing `--data`, set `XDG_STATE_HOME`
before starting tmux (or restart tmux server after changing it).

## Troubleshooting

### Focus alert does not appear

1. Confirm option:
```bash
tmux show-option -gqv @tmux-todo-focus-alert
```
2. Confirm hook:
```bash
tmux show-hooks -g pane-focus-in
```
3. Confirm high-priority task exists:
```bash
tmux-todo has-high --context-only
```

### Popup too big/small

Adjust:

```tmux
set -g @tmux-todo-focus-width "24%"
set -g @tmux-todo-focus-height "14%"
set -g @tmux-todo-peek-width "34%"
set -g @tmux-todo-peek-height "22%"
```

## Development

```bash
go test ./...
go build ./cmd/tmux-todo
```
