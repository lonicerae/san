# Feature 7: Permission System

Controls whether tool calls are allowed, denied, or prompted to the user.
Same pipeline runs in the main loop and in subagents — see
[`docs/gen-permission.md`](../gen-permission.md) for the user-facing spec
and [`docs/permission.md`](../permission.md) for the implementation walk.

## Modes

Main-loop modes (selectable via `Shift+Tab`, `--permission-mode`, or
`defaultMode` in settings):

| Mode | Auto-approves | Use case |
|------|---------------|----------|
| `default` | Reads only | Sensitive work; review every mutation |
| `acceptEdits` | Reads + edit/write tools | Code iteration |
| `dontAsk` | Pre-allowed only; everything else silent Deny | Non-interactive runs (`gen -p ...`, scripts, CI) |
| `bypassPermissions` | Everything | Containers / CI; bypass-immune still enforced |
| `auto` *(TODO)* | Currently aliased to `acceptEdits` | Long-running unattended sessions; classifier pending |

Subagent-only mode:

| Mode | Behavior |
|------|----------|
| `explore` | Read-only investigation pass; mutations explicitly Deny so the agent does not waste turns trying them |

Subagents collapse `Ask → Deny` automatically, so a subagent in `default`
behaves like a read-only assistant.

## Rule Syntax

`Tool(pattern)` glob string. Used identically in:

- `permissions.allow` / `permissions.ask` / `permissions.deny` in `settings.json`
- `allow_tools` / `deny_tools` in agent frontmatter

```
Bash(git:*)              # any git subcommand
Bash(npm run *)          # space-style natural pattern
Read(./src/**)           # cwd-relative path
WebFetch(domain:github.com)
Skill(git:*)
```

Bash compound commands (`&&`, `||`, `;`, `|`, `&`) are split via shell AST.
**Allow rules require every subcommand to match**; deny / ask rules fire on
any subcommand match. Bypass-immune destructive checks (`rm -rf`, `git push
--force`, etc.) apply per subcommand and cannot be silenced.

## Priority

```
deny > ask > allow > mode default
```

Both `settings.{allow,deny,ask}` and `agent.{allow_tools,deny_tools}` feed
the same pipeline.

Sensitive paths and destructive commands remain bypass-immune even in
`bypassPermissions` mode.

## UI Interactions (main loop only)

- **Confirmation dialog**: shows tool name + input; `y` approve, `n` deny,
  `a` always allow.
- **Denied tool**: shows an inline error in the conversation.
- **Allow-list match**: tool runs silently without any dialog.

## Automated Tests

```bash
go test ./internal/setting/... -v -run TestPermission
go test ./internal/tool/perm/... -v
go test ./internal/subagent/... -v -run 'Mode|Permission|Allow|Deny'
go test ./tests/integration/permission/... -v
```

Key cases:

```
# Unified pipeline (setting)
TestMatchRule                                — rule pattern matching
TestBuildRule                                — rule string construction
TestCheckPermission                          — deny / allow / ask / mode interactions
TestBypassPermissionsMode                    — bypass + bypass-immune still enforced
TestDontAskMode                              — Ask coerced to Deny
TestDenyRulesPriorityOverSession             — deny absolute
TestSafeToolAllowlist                        — safe tools auto-allow
TestIsDestructiveCommand                     — destructive pattern catch
TestIsSensitivePath                          — sensitive path catch
TestCheckBashSecurity                        — injection / obfuscation
TestBashSecurityBypassImmune                 — security checks always fire

# Subagent gate (steps 1-4 of unified pipeline + Ask→Deny coercion)
TestExploreModeAllowsOnlyGitDiffBash         — allow_tools per-subcommand match
TestDefaultModeRestrictsConfiguredBash       — allow_tools whitelist
TestDenyToolRulesMatchPatterns               — deny_tools any-subcommand
TestExploreModeFiltersMutatingToolSchemas    — schema filter
TestAcceptEditsModeFiltersApprovalOnlyToolSchemas
TestBypassModeAllowsEverything
TestNormalizePermissionModeDefaultsEmpty

# Read-only / safe tool detection
TestIsReadOnlyToolMatchesConfig
TestIsSafeToolMatchesConfig
```

## Interactive Tests

`gen agent run --type <name> --prompt "..."` runs an AGENT.md fixture
through the full subagent pipeline. Use this to verify allow / deny / mode
behavior end-to-end without a TUI.

```bash
mkdir -p .gen/agents
cat > .gen/agents/test-perm.md <<'EOF'
---
name: test-perm
description: Permission gate fixture
mode: explore
allow_tools:
  - Read
  - Bash(git diff*)
  - Bash(git log*)
deny_tools:
  - Bash(git stash*)
---
You are a test fixture. Run exactly what the user asks.
After each call output: RESULT: <tool>(<args>) -> <ALLOWED|DENIED: reason>
EOF

# Allow path
./bin/gen agent run --type test-perm --prompt 'Run bash: git diff --stat'
#  → ALLOWED

# Per-subcommand allow rejects when one part doesn't match
./bin/gen agent run --type test-perm --prompt 'Run bash: git diff && git status'
#  → DENIED: tool Bash call is outside the allow_tools constraint

# Deny wins over allow
./bin/gen agent run --type test-perm --prompt 'Run bash: git stash list'
#  → DENIED: tool Bash is blocked by deny_tools

# Bypass-immune wins over everything below it
./bin/gen agent run --type test-perm --prompt 'Run bash: git diff && rm -rf /tmp/dummy'
#  → DENIED: destructive command
```

For TUI scenarios (approval dialog, working-directory enforcement, hook
integration), spawn a tmux session and drive `gen` interactively:

```bash
mkdir -p /tmp/perm_test/.gen
cat > /tmp/perm_test/.gen/settings.local.json <<'EOF'
{"permissions": {"allow": ["Bash(echo *)", "Bash(ls *)"]}}
EOF

tmux new-session -d -s t_perm -x 220 -y 60
tmux send-keys -t t_perm 'cd /tmp/perm_test && gen' Enter
sleep 2
tmux send-keys -t t_perm 'Run bash: echo hi && ls /tmp' Enter
sleep 5
tmux capture-pane -t t_perm -p
# Expected: runs without dialog (every subcommand matches an allow rule)

tmux send-keys -t t_perm 'Run bash: echo hi && cat /etc/hosts' Enter
sleep 5
tmux capture-pane -t t_perm -p
# Expected: approval dialog (cat /etc/hosts has no allow rule)

tmux kill-session -t t_perm
rm -rf /tmp/perm_test
```
