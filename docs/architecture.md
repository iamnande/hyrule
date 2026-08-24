# architecture

`hyrule` is, in effect, a reference implementation home - a playbook for
solving real problems, built as working code instead of a wiki page. patterns
here come from scenarios out of real-world engineering situations at Nike, Kong,
and ngrok. it's also the seed for [homelab](https://github.com/iamnande/homelab)
services - patterns proven here are what `homelab/services/<name>` ends up
running.

this doc owns the big picture - what stays true regardless of which entities
or services exist yet. topic depth (API design, probes, the data layer, ...)
lives in [conventions.md](conventions.md), why and rule together. when a
topic outgrows its section there, it graduates into its own doc.

---

## service topology

`internal/lib` is shared across every service - runtime, rest, config,
tracing, logging, version. it carries no domain knowledge. `internal/svc/<name>`
is one service's own domain, api, and (eventually) data access - nothing else
reaches into it. `cmd/<name>` is that service's entrypoint. a new service
repeats this shape; `internal/lib` shouldn't need to change for it to exist.

for current service names, endpoints, and how to run things, see
[README.md](../README.md) - this doc is about the shape, not the roster.

---

## layers

```
┌─────────────────────────┐
│           api           │  pure i/o - decode, call, encode
├─────────────────────────┤
│          domain         │  business logic, orchestration
├─────────────────────────┤
│          service        │  reusable capabilities (cross-domain behavior)
├─────────────────────────┤
│        repository       │  persistence
└─────────────────────────┘
```

---

## runtime

every service shares one process-lifecycle recipe via
[`runtime.NewModule`](../internal/lib/runtime/module.go) - timeouts, logging,
tracing, startup/shutdown, serving HTTP. values live in
[`runtime/timeouts.go`](../internal/lib/runtime/timeouts.go); the one
non-obvious constraint: `StopTimeout` must clear `DrainTimeout` with real
margin. `DrainTimeout` is a sleep on SIGTERM before the rest of shutdown
runs, giving a load balancer time to stop routing traffic here first - if
`StopTimeout` doesn't exceed it, fx kills the shutdown hook mid-sleep instead
of ever waiting for it. this is enforced, not just documented - an `init()`
in `timeouts.go` panics with a clear message if the invariant is ever broken,
so it fails loudly at process startup rather than silently at 2am during a
deploy.
