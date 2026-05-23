<div align="center">
  <h1>&lt; GEN ✦ /&gt;</h1>
  <p><strong>Open-source AI coding assistant for the terminal</strong></p>
  <p>
    <a href="https://github.com/genai-io/gen-code/releases"><img src="https://img.shields.io/github/v/release/genai-io/gen-code?style=flat-square" alt="Release"></a>
    <a href="https://goreportcard.com/report/github.com/genai-io/gen-code"><img src="https://goreportcard.com/badge/github.com/genai-io/gen-code?style=flat-square" alt="Go Report Card"></a>
    <a href="https://pkg.go.dev/github.com/genai-io/gen-code"><img src="https://pkg.go.dev/badge/github.com/genai-io/gen-code.svg" alt="Go Reference"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue?style=flat-square" alt="License"></a>
  </p>
  <p>
    <strong>English</strong> · <a href="README.zh.md">简体中文</a>
  </p>
</div>

Gen Code is a terminal coding assistant with pluggable LLM providers, search engines, personas, and skill/extension surfaces — fully compatible with [Claude Code](https://claude.ai/code) skills, plugins, and MCP servers. Implemented in Go as a single binary with concurrent multi-agent orchestration.

## Features

### Open architecture

<p align="center">
  <img src="docs/diagrams/open-architecture.svg" alt="Gen Code open architecture — four dimensions: model providers, search engines, role switching, and skills/extensions" width="100%">
</p>

- **LLM providers** — Anthropic, OpenAI, Google, Moonshot, Alibaba, MiniMax, Z.ai (GLM); swap via `/model`.
- **Search backends** — Exa, Tavily, Brave, Serper; swap via `/search`.
- **Personas** — Markdown identities scoped to user or project; swap via `/identity` ([details](docs/concepts/harness-channels.md#identity-custom-personas)).
- **Skills & extensions** — Claude Code skills, plugins, and MCP servers run unmodified; sandboxed subagents; lifecycle hooks (shell, LLM, agent, HTTP); auto-loaded project memory.

### Engineering

- **Native performance** — Single Go binary; see [benchmark](#benchmark-gencode-vs-claude-code) for measured numbers.
- **Event-driven coordination** — Parallel subagent execution via a pub/sub hub ([architecture](docs/packages/subagent.md)).
- **Session persistence** — Auto-save, resume, fork, and automatic context compaction.
- **Prompt prediction** — Speculative completion of likely next prompts to reduce latency.


## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/genai-io/gen-code/main/install.sh | bash
```

Re-run to upgrade. To uninstall:

```bash
curl -fsSL https://raw.githubusercontent.com/genai-io/gen-code/main/install.sh | bash -s uninstall
```

<details>
<summary><b>Other methods</b></summary>

**Go Install**

```bash
go install github.com/genai-io/gen-code/cmd/gen@latest
```

**Build from Source**

```bash
git clone https://github.com/genai-io/gen-code.git
cd gen-code
go build -o gen ./cmd/gen
mkdir -p ~/.local/bin && mv gen ~/.local/bin/
```

</details>

## Usage

### Start

```bash
gen                            # interactive
gen "explain this function"    # one-shot
cat main.go | gen "review"     # piped input
gen --continue                 # resume latest session
gen --resume                   # pick a past session
```

### Configure a model

On first launch, run `/model` to pick a provider and model. Gen Code will
prompt for the API key — or you can set the matching env var (see the
**Credentials** table below) and Gen Code will pick it up automatically.
Switch models any time with `/model`; the choice is saved to
`~/.gen/providers.json`.

### Adjust thinking

For models that support extended reasoning (Claude, GPT-5/o-series,
Gemini, etc.), press `Ctrl+T` — or run `/think` — to cycle the thinking
budget. The available levels are provider-specific (Claude:
`off / normal / high / ultra`; OpenAI reasoning models:
`none / low / medium / high / xhigh`). The current level is shown in
the status bar; higher levels trade latency for deeper reasoning.

### Other essentials

- `/help` lists every slash command (`/identity`, `/search`, `/skills`, `/agents`, `/mcp`, `/compact`, `/resume`, …).
- `Shift+Tab` cycles permission modes (ask · auto-accept · plan).
- `Ctrl+O` expands tool details · `Ctrl+C` cancels the current turn · `Ctrl+D` exits.

See [`docs/guides/getting-started.md`](docs/guides/getting-started.md)
for a longer walkthrough and
[`docs/reference/slash-commands.md`](docs/reference/slash-commands.md)
for the full command list.

### Configuration

Config lives in `~/.gen/` (user) and `<project>/.gen/` (project, overrides user). A `GEN.md` or `CLAUDE.md` at the project root is auto-loaded into the system prompt.

<details>
<summary><b>Credentials</b></summary>

| Service | Variable |
|:--------|:---------|
| **Anthropic** (Claude) | `ANTHROPIC_API_KEY` or [Vertex AI](https://cloud.google.com/vertex-ai/generative-ai/docs/partner-models/claude) |
| **OpenAI** (GPT, o-series, Codex) | `OPENAI_API_KEY` |
| **Google** (Gemini) | `GOOGLE_API_KEY` |
| **Moonshot** (Kimi) | `MOONSHOT_API_KEY` |
| **Alibaba** (Qwen, DeepSeek) | `DASHSCOPE_API_KEY` |
| **MiniMax** | `MINIMAX_API_KEY` |
| **Z.ai** (GLM) | `BIGMODEL_API_KEY` |
| **Exa** search | _none_ (default) |
| **Tavily** search | `TAVILY_API_KEY` |
| **Brave** search | `BRAVE_API_KEY` |
| **Serper** search | `SERPER_API_KEY` |

</details>

<details>
<summary><b>Directory layout</b></summary>

User-level (`~/.gen/`):

```
providers.json    # Provider connections and current model
settings.json     # Permissions, hooks, env, identity
skills.json       # Skill states
identities/       # Custom personas (see /identity)
skills/           # Custom skill definitions
agents/           # Custom agent definitions
commands/         # Custom slash commands
plugins/          # Installed plugins
projects/         # Session transcripts + indexes
```

Project-level (`.gen/`):

```
settings.json      # Permissions, hooks, disabled tools
mcp.json           # MCP server definitions
identities/*.md    # Project-scoped personas (override user-level)
agents/*.md        # Subagent definitions
skills/*/SKILL.md  # Skills
commands/*.md      # Slash commands
```

</details>

## Benchmark: Gen Code vs Claude Code

Compared with [Claude Code](https://claude.ai/code) v2.1.112 on Apple Silicon, same model (`claude-sonnet-4-6`):

| Metric | Gen Code | Claude Code | Advantage |
|--------|---------|-------------|-----------|
| Download size | 12 MB | 63 MB (+ Node.js 112 MB) | **5x smaller** |
| Disk footprint | 38 MB | 175 MB | **4.6x smaller** |
| Startup time | ~0.01s | ~0.20s | **20x faster** |
| Startup memory | ~32 MB | ~189 MB | **5.8x less** |
| Simple task | ~2.4s / 39 MB | ~10.4s / 286 MB | **4.3x faster, 7.3x less memory** |
| Tool-use task | ~3.3s / 39 MB | ~26.0s / 285 MB | **7.9x faster, 7.2x less memory** |

Both tools have comparable features (hooks, skills, plugins, session, MCP, etc.). The performance gap comes from Go's native compilation, minimal architecture design, and lean prompt engineering — vs Node.js V8/JIT/GC runtime overhead.

See full details: [docs/operations/benchmark.md](docs/operations/benchmark.md)

## Documentation

- [Documentation Index](docs/index.md) — map of architecture, features, operations, and references
- [Architecture](docs/architecture.md) — architecture entrypoint and reading order
- [Package Map](docs/reference/package-map.md) — package ownership and dependency boundaries
- [System Prompt](docs/concepts/harness-channels.md) — Slot model, identity, skill/agent injection
- [Subagents](docs/packages/subagent.md) · [Skills](docs/packages/skill.md) · [Plugins](docs/packages/plugin.md) · [MCP](docs/packages/mcp.md)
- [Hooks](docs/packages/hook.md) · [Permissions](docs/concepts/permission-model.md) · [Tasks](docs/packages/task.md)
- Per-package design under [`docs/packages/`](docs/packages/) — start at [Package Index](docs/packages/index.md)

## Related Projects

- [Claude Code](https://claude.ai/code) — Anthropic's AI coding assistant
- [Aider](https://github.com/paul-gauthier/aider) — AI pair programming in terminal
- [Continue](https://github.com/continuedev/continue) — Open-source AI code assistant

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
