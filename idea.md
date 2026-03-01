# Notion Agent Sync Plan (tmux-todo)

## Goal
Add an optional integration that syncs `tmux-todo` tasks to a Notion task system through `notion-agent-sdk-go`, while keeping local task management fast and reliable.

## Product Requirements
- Keep `tmux-todo` local-first. No network dependency for core add/list/done workflows.
- Support explicit sync operations (`create`, `update`, `close`, `batch`).
- Capture enough context (git branch/worktree/PR summary/task metadata) for intelligent routing on the Notion side.
- Allow user-level defaults (project, prompt template) with per-command overrides.
- Never auto-close local tasks unless remote sync succeeds.

## Non-Goals (for first version)
- Full bi-directional sync from Notion back into local tasks.
- Background daemon/always-on sync.
- Tight coupling between TUI and network calls.

## High-Level Architecture
1. `tmux-todo` core remains unchanged for local state.
2. Add `sync` command group in CLI.
3. `sync` command builds a structured payload:
- task data
- context data
- optional git/PR data
- user prompt instructions
4. Payload is sent to a Notion routing agent via `notion-agent-sdk-go`.
5. Response is validated and applied:
- persist remote identifiers and sync metadata
- optionally close/mark local tasks only after successful remote write
6. Write an audit entry for each sync run.

## Data Model Additions
Add optional sync fields to each local task:
- `notion_page_id` (string)
- `sync_state` (`never`, `ok`, `error`)
- `last_synced_at` (timestamp)
- `last_sync_error` (string)
- `remote_project` (string)
- `pr_url` (string)

Add top-level sync audit log file (JSONL) in state dir:
- default: `$XDG_STATE_HOME/tmux-todo/sync.log` (or `~/.local/state/tmux-todo/sync.log`)

## Configuration
Extend user config (`~/.config/tmux-todo/config.json`):

```json
{
  "sync": {
    "enabled": true,
    "provider": "notion-agent",
    "default_project": "Core Engineering",
    "default_mode": "suggest",
    "default_prompt_template": "default",
    "auto_close_local_on_remote_close": false,
    "require_confirmation_for_close": true,
    "include_git_diff": "summary",
    "include_pr_context": true,
    "max_batch_close": 20
  },
  "notion_agent": {
    "endpoint": "http://localhost:7777/agent",
    "api_key_env": "NOTION_AGENT_API_KEY",
    "timeout_ms": 20000
  }
}
```

Notes:
- Config is user-scoped, not repo-scoped.
- CLI flags always override config for one-off runs.

## Prompt and Instruction Strategy
Use **template + structured payload** instead of free-form prompting by default.

### Default strategy
- Fixed template selected by operation:
  - `sync_create`
  - `sync_update`
  - `sync_close`
  - `sync_batch`
- Structured payload includes task/context/git/PR fields.

### Override strategy
- `--instruction "..."` appends user instruction to selected template.
- `--template <name>` picks alternate template.
- `--raw-prompt-file <path>` bypasses templates (advanced/debug only).

This gives predictable behavior for team usage while still supporting operator intent.

## Context Capture Strategy
Collect context in layers:

1. Local task context (always)
- task id, text, status, priority, tags, parent/child, timestamps
- scope (`context` or `global`)

2. Git context (if available)
- repo root, worktree root, branch
- HEAD SHA
- recent commits summary (`N=5`, configurable)
- changed files summary (`name-status`, configurable)

3. PR context (if available and enabled)
- detect via `gh pr view --json ...` for current branch
- fallback: no PR data if unavailable

4. User intent
- operation type (`create`, `update`, `close`, `batch`)
- project override
- optional instruction override

## Project Routing
Project selection precedence:
1. `--project` flag
2. task-level `remote_project` if already set
3. `sync.default_project` from config
4. agent-side routing default

Recommended behavior:
- First sync writes chosen project back to local task metadata.
- Subsequent updates use stored project unless explicitly overridden.

## CLI Design (Proposed)
```bash
tmux-todo sync create --id <task-id> [--project X] [--instruction "..."] [--dry-run]
tmux-todo sync update --id <task-id> [--project X] [--instruction "..."] [--dry-run]
tmux-todo sync close  --id <task-id> [--project X] [--instruction "..."] [--dry-run]
tmux-todo sync batch  --scope context|global|all [--done-since 24h] [--project X] [--dry-run]
tmux-todo sync status --id <task-id>
tmux-todo sync doctor
```

Common flags:
- `--project`
- `--template`
- `--instruction`
- `--include-pr/--no-include-pr`
- `--include-diff summary|none`
- `--dry-run`
- `--json`

## Payload Contract (Proposed)
```json
{
  "operation": "update",
  "project": "Core Engineering",
  "task": {
    "id": "abc123",
    "text": "Refactor sync layer",
    "done": false,
    "priority": 1,
    "tags": ["blocked", "review"],
    "scope": "context",
    "context_key": "repo=...|wt=...|br=..."
  },
  "context": {
    "label": "tmux-todo/tmux-todo [main]",
    "repo_root": "/repo",
    "worktree_root": "/repo",
    "branch": "main",
    "head_sha": "abcd1234"
  },
  "git": {
    "recent_commits": ["..."],
    "changed_files": ["M internal/ui/model.go"]
  },
  "pr": {
    "url": "https://github.com/org/repo/pull/123",
    "title": "...",
    "state": "OPEN"
  },
  "instruction_override": "Tie this to release milestone"
}
```

## Response Contract (Proposed)
```json
{
  "ok": true,
  "action": "updated",
  "notion_page_id": "xxxxxxxx",
  "project": "Core Engineering",
  "status": "In Progress",
  "close_local": false,
  "notes": "Linked to PR #123"
}
```

Validation rules:
- `ok` must be true to mutate local sync fields.
- Local `done=true` should only be set when:
  - operation is `close`
  - response returns `close_local=true`
  - and close confirmation rules are satisfied.

## Implementation Phases

### Phase 1: Foundation
- Add sync metadata fields in store model.
- Add config structures and loading defaults.
- Add sync audit logger.
- Add `sync doctor` command.

### Phase 2: Single Task Sync
- Implement `sync create|update|close --id`.
- Build payload with task + context + optional git data.
- Wire provider client abstraction and Notion Agent implementation.
- Support `--dry-run` and `--json`.

### Phase 3: Batch Sync
- Implement `sync batch`.
- Add safeguards (`max_batch_close`, confirmation gates).
- Add summary output and per-item status.

### Phase 4: PR-Aware Context
- Add GitHub PR capture (best effort).
- Include PR fields in payload.
- Improve error reporting for missing `gh`/auth.

### Phase 5: UX and Automation
- Add optional TUI actions (`sync current task`, `sync done tasks`).
- Add optional post-sync local close behavior.
- Add docs/examples for Claude/automation workflows.

## Testing Plan

### Unit tests
- Config parsing/default precedence.
- Payload builder with/without git context.
- Project precedence resolution.
- Response validator and local mutation logic.
- Sync metadata persistence in store.

### Integration tests (mock provider)
- `sync create/update/close` success path.
- Provider failure path (no local close on failure).
- `--dry-run` no mutation behavior.
- Batch processing limits and confirmation behavior.

### End-to-end smoke tests
- Local repo with branch + mock PR data.
- Run sync commands and inspect JSON outputs + store file.
- Verify audit log entries and error reporting.

## Failure Handling
- If provider call fails:
  - do not alter task completion state
  - set `sync_state=error` and `last_sync_error`
  - append audit log with failure details
- If partial batch fails:
  - continue with next task unless `--fail-fast`
  - emit final summary counts (ok/failed/skipped)

## Security and Privacy
- Never store raw API keys in config; only env var references.
- Redact sensitive fields in logs (`api_key`, tokens, full diff if needed).
- Keep prompt payload size bounded and configurable.

## Open Decisions
1. Should PR context use only `gh` CLI initially, or include provider-specific APIs now?
2. Should `sync close` default to requiring an explicit `--confirm-close`?
3. Should templates live inside repo (`internal/sync/templates`) or user config directory?
4. What max payload size should be enforced for diff/commit context?

## Recommended First Slice
Ship Phase 1 + Phase 2 without PR context first:
- deterministic and testable
- immediate value for manual create/update/close sync
- minimal blast radius on existing TUI workflows
