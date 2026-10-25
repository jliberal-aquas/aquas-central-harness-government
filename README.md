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

## Initiatives

Each initiative opens its own map in the wiki (in Spanish):

- [E000 · Governance maintenance](docs/E000/README.md): guards, compiler and test quality of this repo.
- [E001 · Work governance](docs/E001/README.md): the fixed path of every request, backlog, wiki, agent memory and workers.
- [E002 · Child projects](docs/E002/README.md): install, build, update and report governance in each Aquas product.

## Quick start

```bash
git config core.hooksPath .githooks
go run ./cmd/aquas iniciar
go build -o .tmp/aquas.exe ./cmd/aquas && ./.tmp/aquas.exe verificar
```

Requires Go 1.26. The CLI is being migrated to English command names.

## License

Not yet defined.
