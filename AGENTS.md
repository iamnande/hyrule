# Project notes for agents

For how this repo is built and the rules for changing it, see
[README.md](README.md), [docs/architecture.md](docs/architecture.md),
[docs/conventions.md](docs/conventions.md), and [docs/style.md](docs/style.md)
(general Go style, as opposed to conventions.md's repo-specific rules) -
read `docs/conventions.md` before making changes.

Deliberate decisions in this repo - do NOT silently revert them: see
[docs/decisions](docs/decisions). current priorities and what's not built
yet: [docs/roadmap.md](docs/roadmap.md). PRDs and tech specs for planned
initiatives live in [docs/proposals](docs/proposals), tracked normally like
decisions - not a scratch/gitignored planning directory.

## Maintaining this file

Keep this file for knowledge useful to almost every future agent session in this project.
Do not repeat what the codebase already shows; point to the authoritative file or command instead.
Prefer rewriting or pruning existing entries over appending new ones.
When updating this file, preserve this bar for all agents and keep entries concise.
