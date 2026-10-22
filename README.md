# Aquas Central Harness Government

The central governance harness for every Aquas product built with Claude Code.
It defines how work is done, enforces it with programs, and installs it into each project.

## Why

Each product (a landing page, a publishing house, a book, a FastAPI service) would otherwise carry its own hand-copied rules, drifting apart over time.
And an AI agent is fast but unreliable without controls: it forgets, drifts out of scope, invents, or claims "done" without proof.

## What it is

1. **Governance written once.** Constitution, methodology and agents live in sealed canonical files (`gobierno/canonico/`).
2. **Enforced by programs.** Anything checkable is checked by a Go gate or a Claude Code hook, not by the agent's goodwill.
3. **Ordered work.** Initiative → feature → slice; a failing test comes before the code.
4. **Distribution to child projects.** `install`, `update` and `report` commands (in progress, initiative E002).
5. **Explained for humans.** A wiki in Spanish under [`docs/`](docs/README.md).

## Quick start

```bash
git config core.hooksPath .githooks
go run ./cmd/aquas iniciar
go run ./cmd/aquas verificar
```

Requires Go 1.26. The CLI is being migrated to English command names.

## License

Not yet defined.
