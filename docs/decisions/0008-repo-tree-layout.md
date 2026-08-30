# 0008: repo tree layout - go/, deploy/, local/

root had grown to ten tracked directories (`api`, `cmd`, `deploy`, `docs`,
`hack`, `internal`, `migrations`, `mk`, `stack`, `tests`) with no grouping
principle - Go source, real deploy artifacts, and local-only dev tooling
all sat as flat siblings. this repo is meant to become a monolith
(architecture.md's whole premise), so the layout needed to say something
about what kind of thing each top-level directory is before more services
and a second protocol (gRPC, alongside HTTP) make it worse.

## go/

`cmd`, `internal`, and `tests` moved under `go/` - `go/cmd/<name>`,
`go/internal/lib`, `go/internal/svc/<name>`, `go/tests`. `go.mod` stays at
repo root: a single Go module doesn't need `go.mod` next to the source it
owns, and keeping it at root means `make`, CI, and IDEs all keep working
without a `-C go` dance. `api/` and `migrations/` stay at root too - they're
contracts and schema, not Go concerns, even though Go code is generated
from both.

`internal/` stays nested (`go/internal/...`), not flattened to `go/svc`
/`go/lib`. the compiler enforcement is identical either way today - every
Go source file lives under `go/`, so `go/internal` and a hypothetical
`internal` at repo root would block exactly the same importers (none exist
outside this module). the boundary only pays off later, if this repo ever
gets a second, independently-published Go module that shouldn't reach into
`lib`/`svc` internals - nothing on the roadmap needs that yet, but the cost
of keeping the segment is one path component, so there's no reason to trade
away the option for zero present gain.

## deploy/ vs local/

`deploy/` now means only real deployable artifacts - `deploy/helm/*` and
`deploy/values/<service>/`, the same shape whether the target is the local
cluster or a real homelab one later.

everything that exists purely to support a laptop's dev loop moved into a
new `local/`:

- `local/compose.yml`, `local/postgres/init/*.sql` - was `stack/`
- `local/cluster/postgres.yaml`, `local/cluster/migrate-job.yaml` - was
  `deploy/local/*.yaml`
- `local/new-service.sh` - was `hack/new-service.sh`

these three didn't obviously belong together by function (one's a compose
stack, one's k8s dev scaffolding, one's a code generator), but they share
the property that actually matters for where they live: nothing in any of
them ships anywhere. `deploy/local/` already used "local" for exactly this
distinction; this just makes it a first-class top-level concept instead of
a subdirectory of the thing it's meant to be distinguished from.

## stack/compose.yml removed, not moved

compose (`hyrule-database` + `hyrule-mail-server`) predated the Rancher
Desktop/k3s pivot and had been a deliberately deferred duplicate
environment since (see 0004-local-cluster). moving to `local/` was going to
just relocate that duplication, not resolve it - so it's deleted instead:

- **local dev** already gets postgres from Tilt's `local_database()`
  (`local/cluster/postgres.yaml`, port-forwarded to `localhost:5432`) -
  compose's copy was redundant once that existed.
- **CI** (`integration`/`smoke` jobs) turned out to only ever need a
  reachable postgres, not a deployed stack: `test-integration` builds an
  in-process `fxtest` app and talks to the domain's `http.Handler`
  directly (`go/tests/utils/app.go`), and `test-smoke` builds and runs the
  real binary as a plain host process, curling `localhost:8000`. neither
  needs Tilt, Helm, or a cluster - they need `HYRULE_DATABASE_*` env vars
  to resolve. `make db-up`/`db-down` (`mk/database.mk`) replace
  `stack-up`/`stack-down`: a bare `docker run` against
  `postgres:17.2-alpine`, mounting `local/postgres/init/` the same way
  compose did, keeping the same "wait for the ready log line twice" logic
  CI already depended on (see conventions.md#ci).
- **mailhog** had no consumer anywhere in the codebase - dropped with no
  replacement, same as any other speculative dependency built ahead of a
  real need.

this is independent of roadmap.md's initiative 1 (CI coverage for the
k3s/Tilt/Helm path) - that initiative is about verifying the *deploy*
itself (the chart, the Tiltfile, a rendered manifest), which nothing in CI
does today and this change doesn't add. removing compose only closes the
gap between "CI's test jobs" and "what those jobs actually need," which
turned out to be much smaller than the compose file implied.
