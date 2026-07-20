# Documentation index

“Read when” lookup for every doc under `docs/`.

| | |
| --- | --- |
| **Applies to** | Finding the right doc when you know your topic |
| **Audience** | Contributors and agents after [guide.md](./guide.md) |
| **Prerequisite** | [guide.md](./guide.md) for learning paths; [agent-map.md](./agent-map.md) for code paths |

## In this article

- [Overview](#overview)
- [Navigation](#navigation)
- [Reference and overview](#reference-and-overview)
- [Implementation and deep dive](#implementation-and-deep-dive)
- [See also](#see-also)

## Overview

Use [guide.md](./guide.md) to pick a learning path by goal. Use the tables below when you already know what you need.

> **Tip** — Agents: code locations live in [agent-map.md](./agent-map.md).

## Navigation

| Doc | Read when |
| --- | --- |
| [guide.md](./guide.md) | You are new or need to pick a learning path by goal |
| [agent-map.md](./agent-map.md) | You need a repository path for a subsystem you are editing |

## Reference and overview

| Doc | Read when |
| --- | --- |
| **[execute-and-verify.md](./execute-and-verify.md)** | **You create tasks or write done criteria (checklist items).** |
| [architecture.md](./architecture.md) | You need how `taskapi`, the store, the agent worker, and SSE fit together |
| [data-model.md](./data-model.md) | You touch tasks, projects, cycles/phases, dependencies, gates, or checklists |
| [api.md](./api.md) | You need the REST + SSE endpoint surface (handler code is authoritative for status codes) |
| [configuration.md](./configuration.md) | You change env vars, app settings, agent worker configuration, or schema migrate behavior (see **Schema migrations** in that doc) |
| [naming.md](./naming.md) | You need product/operator identifier conventions (Hamix, `HAMIX_*`, metrics) |
| [domain/worktrees-and-branches.md](./domain/worktrees-and-branches.md) | Git worktrees, branches, task binding, `/repo/*`, and worker checkout |

## Implementation and deep dive

| Doc | Read when |
| --- | --- |
| [web.md](./web.md) | You work on the `web/` SPA |
| [domain/](./domain/) | You need why a subsystem behaves as it does — index: [domain/README.md](./domain/README.md) |
| [omitted-features.md](./omitted-features.md) | A feature exists in code but is hidden for launch |
| [adr/](./adr/) | You need the historical reason behind a design decision — index: [adr/README.md](./adr/README.md) |

## See also

- [guide.md](./guide.md) — documentation layers and learning paths
- [agent-map.md](./agent-map.md) — subsystem code paths
- [CONTRIBUTING.md](../CONTRIBUTING.md) — setup and PR checklist
