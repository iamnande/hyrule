# conventions

the rules for writing code in this repo, why included - read before changing
anything. [architecture.md](architecture.md) has the big picture this all
sits inside; this doc is where topic depth actually lives.

## packages

- `internal/lib` - shared, no domain knowledge
- `internal/svc/<name>` - one service's own domain/api/data
- `cmd/<name>` - that service's entrypoint

see [architecture.md#service-topology](architecture.md#service-topology) for
the full shape. new service = new `internal/svc/<name>` + `cmd/<name>`,
following an existing one exactly; `internal/lib` shouldn't need to change.
`make new-service` scaffolds the wiring (not the domain logic) for a new
service - see [0006-service-scaffold](decisions/0006-service-scaffold.md).

**a domain package defines its own narrow interfaces for what it depends
on** - never imports the concrete types of the services/repositories it
calls. that's the whole trick that makes a hand-written fake enough for a
unit test (see testing, below): there's no concrete dependency to mock,
only an interface to satisfy.

## fx

- **config loading happens at the composition root** (`cmd/<name>/main.go`),
  not inside a service's own `Module`. a service's `Module` is logic and
  wiring only - keeping config loading separate is what lets tests supply
  fixed values directly instead of overriding real ones.
- **no `fx.Decorate`.** it only exists to override a value something
  upstream already provides - reaching for it is a sign the value is being
  provided somewhere it shouldn't be, not a reason to intercept it. fix the
  source instead.
- **validate config where it's loaded** (the `Load*()` function), returning
  a plain error - fx already surfaces provider errors natively, that's real
  infrastructure, use it. no custom `encoding.TextUnmarshaler` or other
  stdlib-interface hooks bolted on just to get validation to fire.
- **health API and router are free for every service** via
  `runtime.NewModule` - a service's own `Module` only supplies what's
  actually service-specific (`health.Probes`, its own `group:"apis"`
  handlers). `runtime.HTTPModule` (the `NewHealthAPI`/`NewRouter`
  providers) is split out from `NewModule` so tests can pull in a working
  `http.Handler` without also getting `fx.Invoke(rest.NewServer)`, which
  would bind a real port.

## types

**enum-shaped values get their own named string type** (`Environment`,
`Region`, ...), used end to end in struct fields - never `string` with a
cast at the boundary. Go marshals a named string type to JSON exactly like
a plain string, so there's no cost to keeping the type all the way through.
casts are only acceptable where the consumer hard-requires `string`
(`slog.String`, a `map[string]string`) - never because a struct field was
typed lazily.

## API design

**APIs are pure i/o.** a handler decodes a request, calls the domain,
encodes a response. there must be no business logic in the API layer - if
you're writing an `if` statement in a handler that isn't about http (status
codes, decoding errors), it's in the wrong layer.

**`rest.APIHandler` is `func(chi.Router)`, not a self-contained handler
Mounted at a path.** every service's paths are flat (see versioning,
below), so health and every service API all register routes on one shared
router. chi only allows one `Mount()` per pattern - two independent
sub-routers both trying to claim `"/"` panics. registering directly on the
shared router instead of Mounting a sub-router sidesteps that entirely.

**OAS-first.** the OpenAPI spec is the contract, not a byproduct of it.
types and validators are generated from it (`oapi-codegen`); handlers
implement the generated interface. the spec is written and reviewed before
the handler exists.

**the spec lives at `api/<service>/openapi.yaml`** - a top-level `api/`
tree, not tucked inside `internal/svc/<service>/api/`. it's the contract a
client reads, not an implementation detail of the Go package that happens
to generate code from it; `internal/svc/<service>/api/` stays Go-only (the
codegen config, generated types, and handlers).

**errors are catalog entries, not inline literals.** every distinct error
is a `Definition` (code, name, message, status) in one registry
(`internal/lib/rest/transport/errors/registry.go`), looked up by code -
never a `BaseError{...}` struct literal hand-typed at each call site.
shaped after [ngrok's error reference](https://ngrok.com/docs/errors/reference):
stable enough to depend on, one code per class of failure. hand-written
today - a generator reading a config file, and a docs page per code, are
deliberately deferred. this is the record-keeping start, not the full
system.

codes are `ERR_HYRULE_NNN`, numbered by range per domain:

| range   | domain                                    |
|---------|--------------------------------------------|
| 000-099 | internal/system - not the caller's fault  |
| 100-199 | request validation - the caller's fault   |

## URL structure & versioning

```
domain: api.morethq.com
path:   /<entity>
header: Hyrule-Version: 2026-08-22
```

**the URL identifies the resource** - it should describe *what*, not which
binary serves it today or which contract snapshot the client wants. version
lives in a header, not the path: ngrok (`Ngrok-Version`) does this, and a
resource's address shouldn't change just because its contract did.

**version is scoped per service, not per entity.** a service ships one
version for its entire surface, covering every entity it owns. an identity
service owning `account`/`workspace`/`user`/`team` bumps once, together - a
client pinned to a version gets one coherent view of the whole service, not
four independently-drifting numbers to reconcile. this only stays humane if
most changes are additive and need no bump at all; a genuine breaking change
earns a new version, and a transform/shim layer translates old pins to the
current internal shape, so a break in `user` doesn't actually break a client
still reading `team` on an older pin.

**version format is dated** (`2026-08-22`), not an incrementing integer - a
date doubles as an implicit changelog entry, a bare integer doesn't.

**the tell to split into a new service isn't "it has two entities"** - it's
when entities stop sharing a bounded context (data ownership, operational
lifecycle, team ownership) and are only colocated for convenience.

**relationships stay flat by default.** reference the related resource by a
field, filter with a query param (`/pings?filter[id]=X`). nest a resource
under its parent only for true composition - the child has no identity or
lifecycle apart from the parent (e.g. `/customers/{id}/sources`).

## probes

```
/startupz  startup probe
/livez     liveness probe
/readyz    readiness probe
/healthz   dependency diagnostics - not a probe target itself
```

startup, readiness, and liveness are three different questions, not one
endpoint with three names - credit to Sam Rose's
[Kubernetes Probes](https://ngrok.com/blog/probes) for the framing below.
paths match kubelet's own convention (`kube-apiserver` exposes
`/livez`/`/readyz` the same way) over a bespoke scheme; `/healthz` predates
the livez/readyz split and is deprecated upstream, which frees it up here
for the aggregate dependency view instead.

- **startup** answers "is one-time initialization done" - config loaded,
  database reachable, caches warmed. it's fine for this to be the heaviest
  and slowest of the three: kubelet polls it with a generous failure
  threshold before anything else runs, and once it succeeds it's never
  polled again. this is where the real dependency check belongs.
- **readiness** answers "should traffic route to *this* replica, right now" -
  local, fast, no network calls to other services. it must not fail on
  shared-dependency issues (database, external apis) - taking every replica
  out of rotation at once for a shared blip is the exact cascading failure
  probes exist to prevent. a false failure here is cheap (traffic just
  avoids this replica until it recovers), so it's fine to be a little
  trigger-happy.
- **liveness** answers "is this process stuck in a way only a restart
  fixes" - a deadlock, a stalled event loop, a watchdog that stopped
  advancing. the bar here is higher than readiness, because the consequence
  is a restart, not just reduced traffic. a static 200 (no real check) is a
  legitimate default until a specific stuck-process failure mode has
  actually been observed - don't invent one preemptively.
- hard/soft dependency checks (`/healthz`) feed operator diagnostics, never
  the probes themselves.

see [internal/lib/rest/capabilities/health/doc.go](../internal/lib/rest/capabilities/health/doc.go)
for the implementation.

## tracing

**span and tag names: `<component>.<action>[.<qualifier>]`, dots only** -
e.g. `dependencies.check.all`, `dependencies.check.database`. one
separator, not whatever felt right at the call site - colons snuck in early
and it was a mess.

[internal/lib/tracing](../internal/lib/tracing) gives you three ways to
start a span, cheapest first:

- `tracing.Start(ctx)` - the default. names itself after the calling
  function, so you never type a name at all.
- `tracing.StartNamedf(ctx, format, args...)` - one span per item in a loop
  (a dependency check, a resource) where the name needs a data-dependent
  segment.
- `tracing.StartNamed(ctx, name)` - static name that isn't the caller's.
  rare enough that if you're reaching for it, `Start` or `StartNamedf` is
  probably what you actually want.

## data layer

**postgres.** RLS is a hard requirement eventually (see below), which rules
out sqlite outright - it has no RLS concept.

**no ORM.** past experience (Kong) with a heavyweight ORM was fine until a
query got complex, then the abstraction fought back. instead:

- **`sqlc`** generates typed Go from real SQL - table DDL and query files,
  never a query-builder DSL. this is the data-layer mirror of OAS-first: the
  SQL is the source of truth, generated code is the boring, disposable part.
  there's no abstraction ceiling to hit because there's no abstraction over
  the query language itself.
- **`pgx`/`pgxpool`** as the driver and pool.
- **`golang-migrate`** for migrations - see
  [docs/decisions/0001-migration-tool.md](decisions/0001-migration-tool.md)
  for why over `atlas`.

**RLS-readiness, built now even though no policy exists yet:**

- every repository call runs inside an explicit transaction, even a single
  `SELECT` - RLS policies key off session GUCs (`SET LOCAL app.account_id =
  ...`), which only live for the current transaction. `database.WithTx`
  (`internal/lib/database/tx.go`) is that helper, with the GUC step
  (`setGUCs`) a no-op today. building this now means adding a real policy
  later is additive; building it later means reworking every repository
  call site.
- the app connects as a dedicated, non-owner postgres role (`hyrule_app`)
  from the start - RLS is silently bypassed for the table owner and
  superusers unless `FORCE ROW LEVEL SECURITY` is set, and it's an easy trap
  to inherit from whatever role created the tables in local dev. locally
  that role is bootstrapped by
  [stack/postgres/init](../stack/postgres/init) on first container start;
  migrations still run as the owner role (`POSTGRES_USER`, see
  `DATABASE_URL` in `mk/database.mk`).
- **`hyrule_app_ro` (strict `SELECT`, no write grants) is for a service
  whose read path should never be able to write, at the database level,
  not just in application code** - `iam-jwks` is the first to use it (see
  [0007](decisions/0007-iam-jwks-key-distribution.md)). bootstrapped
  alongside `hyrule_app` in
  [stack/postgres/init](../stack/postgres/init).
- entities that have an *owner* (account/workspace/user/team, when they
  exist) get their scoping column in their first migration, populated from
  day one, even before any policy references it. `pings` does not get one -
  it's self-reported homelab telemetry with no tenant, and a scoping column
  with nothing to scope by is the over-application of this pattern, not the
  point of it.
- the health package's dependency check is a direct `pgx` ping
  (`internal/lib/database/health.go`), replacing the old `dynamodb.Scan`
  call. not generalized behind a cross-backend interface - there's exactly
  one backend to support, and building that abstraction before a second
  backend exists is premature.

**migrations live in [migrations](../migrations)** at the repo root,
applied with `make db-migrate-up` (`make db-migrate-create NAME=...` to
scaffold one) - see [mk/database.mk](../mk/database.mk). one shared
schema, every service's tables in it - so a table is named
`<service>_<entity>` (`iam_jwks_keys`), not just `<entity>`, once there's
more than one service. `pings` predates this and stays unprefixed rather
than a needless rename.

**`sqlc` generates into `repository/<entity>`, not `repository/generated`.**
config at [sqlc.yaml](../sqlc.yaml); each service gets its own `sql:`
entry, reading only that service's own migration file(s) - not the whole
`migrations/` directory - and its own `repository/queries/*.sql`.
pointing `schema` at the whole directory generates a Go struct for
*every* table in the shared schema into *every* service's package,
including tables it has nothing to do with; a service's generated
package should only ever contain its own models. the package is named
after the entity (e.g. `repository/ping`), so call sites read as
`ping.Upsert(...)`, not a stutter. run `sqlc generate` after any
migration or query change.

**every query gets a supporting index, in the same migration.** if a query
filters, joins, or orders on a column and that column isn't already covered
by an existing index (a primary key covers itself - `pings.name` needs
nothing extra for `WHERE name = $1` or `ORDER BY name`), the migration adding
the query adds the index too. retrofitting an index after a table has real
rows means an `ALTER` that locks or a `CREATE INDEX CONCURRENTLY` dance -
cheap to avoid by never letting a query ship unindexed in the first place.

## testing

**`test-unit` (`./internal/...`) vs `test-integration`** (`./tests/...`,
against the real local stack). anything that can't be faked (RLS, above) is
integration-only, never mocked - it's enforced by the postgres engine
itself.

**ginkgo/gomega everywhere**, including domain/service unit tests -
colocated as `<file>_test.go` + one `suite_test.go` per package, not just
for integration specs. no mocking framework: a domain's own narrow
interfaces (see packages, above) make a hand-written fake enough.

**`test-integration` runs via the `ginkgo` CLI** (a `go.mod` tool, `go tool
ginkgo`), not plain `go test`:

- `-p`-style parallelism across suites, and across a suite's own top-level
  `Describe` blocks, only happens through Ginkgo's own multi-process
  runner - not `go test`'s package-level concurrency.
- within one `Describe`, specs stay `Ordered` (required for
  `BeforeAll`/`AfterAll`) - they share one running server/DB connection and
  are inherently sequential (a `GET` asserting on what a prior `POST`
  created), not a parallelism opportunity.
- the parallelism is *across* `Describe` blocks and suites, not within one.

## config

**no `.env`, ever.** real environment variables, sane `envDefault` tags
cover local dev. local-only values (compose stack credentials) get
hardcoded where they're used - they aren't secrets. real secrets get
injected at invocation time (e.g. `op run -- make run`), never touch disk.

## container image

- **final stage pinned to `gcr.io/distroless/base-debian12`**, not
  `gcr.io/distroless/base` or `-static`. `base` (no suffix) was verified to
  ship a busybox toolkit baked into its filesystem on the digest tested
  here - the opposite of what "distroless" is for. `-static` has no CA
  certificates, which breaks outbound HTTPS (Sentry). `-debian12` is the
  one that's actually both minimal and functional.
- **`TARGETOS`/`TARGETARCH`, not `TARGET_OS`/`TARGET_ARCH`** - those are
  buildkit/buildah-reserved arg names, auto-populated from `--platform`.
  renaming them means losing that for a readability preference; their
  default values match this repo's usual dev machine, not a claim that's
  the only target.
- **the final-stage `COPY --from=build` destination (`/service`) is a
  fixed path**, not `${SERVICE_NAME}` - the exec-form `ENTRYPOINT` below it
  doesn't get build-arg substitution, so a parameterized path there would
  silently break.

## commits

[Conventional Commits](https://www.conventionalcommits.org/), `type(scope):
description`. scope is usually a package/service name; `*` for something
repo-wide. `!` after the scope for a breaking change.

```
feat(pings): add ping ingest endpoint
fix(database): pool leak on failed health check
chore(*): bump go to 1.27
feat(*)!: new hyrule, who dis
```

## already in practice, worth continuing deliberately

- errors wrapped with `%w` and enough context to debug without a stack
  trace (`errorlint` catches the worst drift here - see `.golangci.yml`)
- structured logging via `slog`
- sentry spans around domain-level operations

the latter two aren't enforced by tooling yet, so it's easy to drift from
without noticing - this is the reminder.

## CI

[.github/workflows/ci.yml](../.github/workflows/ci.yml): four independent
jobs (`lint`, `unit`, `integration`, `smoke`).

- **no `needs:` chaining between jobs** - they don't share expensive setup,
  so failing fast on one doesn't save meaningful time, and running them in
  parallel gets feedback faster.
- **`integration`/`smoke` run the exact same `make stack-up` /
  `make db-migrate-up` / `make test-*`** a human runs locally, not a
  CI-specific reimplementation - "works in CI" and "works locally" stay the
  same claim.
- **toolchain comes from `mise.toml` via `jdx/mise-action`**, same as
  `make bootstrap`.
- **`make stack-up` waits for postgres by grepping `docker logs`** for
  "database system is ready to accept connections" *twice*, not
  `pg_isready`. the official postgres image starts twice on first boot -
  once to run `docker-entrypoint-initdb.d`, then it stops and restarts for
  real - and `pg_isready` can pass in the gap between those two starts,
  right before the reset. waiting for the ready line twice (once per
  start) doesn't have that race; this is what caught CI's `integration`
  job cold the first time this workflow ran.

## local orchestration

[0003-runtime](decisions/0003-runtime.md) (k3s) +
[0004-local-cluster](decisions/0004-local-cluster.md) (Rancher Desktop) +
[0005-helm-chart-split](decisions/0005-helm-chart-split.md) (app +
platform + wrapper). `make cluster-up` / `cluster-down` / `cluster-status`
/ `dev` (runs `tilt up`) / `helm-vendor` - see [Tiltfile](../Tiltfile) and
[deploy/](../deploy).

- **one generic chart (`deploy/helm/app`) serves every service** - no
  per-service chart. a new service adds a `deploy/values/<service>/
  values.yaml`, not a new chart directory. `deploy/helm/platform` holds
  homelab integration plumbing (near-empty today), `deploy/helm/app-platform`
  wraps both - see 0005 for the full shape and why.
- **`deploy/helm/app-platform/charts/*.tgz` and `Chart.lock` are
  committed, vendored artifacts** - run `make helm-vendor` after editing
  `app` or `platform` and commit the result, same as any other
  generated-and-committed output in this repo.
- **`deploy/local/postgres.yaml` is dev-only** - ephemeral storage, no
  chart, reuses `stack/postgres/init/01-app-role.sql` verbatim (the
  Tiltfile reads that file directly into a ConfigMap rather than
  duplicating its contents). not a template for how a real homelab
  postgres should run. named `hyrule-database`, not after any one
  service - every database-backed service shares it.
- **the Tiltfile discovers services from `deploy/values/*`** rather than
  listing them - see
  [0006-service-scaffold](decisions/0006-service-scaffold.md).
- **every service takes config through env vars, full stop** - `app` has
  no mechanism for mounting a config file into a pod. `env` map keys are
  the literal env var names (no prefix magic; pings' own values file
  writes `HYRULE_DATABASE_HOST` etc. in full - the chart has no opinion
  on any service's naming convention). see
  [0005-helm-chart-split](decisions/0005-helm-chart-split.md#config-env-vars-only-no-config-files).
- **Rancher Desktop runs in containerd mode, not dockerd/moby** - matches
  k3s's own embedded runtime. the [Tiltfile](../Tiltfile) uses the
  `ext://nerdctl` extension's `nerdctl_build` (Tilt's own documented path
  for this mode) in place of `docker_build` - same shape, different
  backend, no separate image-load step since builds land directly in the
  containerd store k3s already reads from.
- **kubeconfig context is `rancher-desktop`** - created by the app itself,
  nothing to check in.

## known gaps

- `golangci-lint`, `golang-migrate`, and `sqlc` aren't provisioned by `make
  bootstrap` - installed locally on faith (e.g. `brew install golangci-lint
  golang-migrate sqlc`). CI installs `golangci-lint` and `golang-migrate`
  itself (pinned versions, see the workflow), but `make bootstrap` still
  doesn't. `oapi-codegen` doesn't have this problem - it's a `go.mod` tool
  dependency, invoked as `go tool oapi-codegen`.
- [docs/style.md](style.md)'s consumer-defined-interfaces rule isn't
  linted - nothing stops a producer-side interface from creeping in.
- the `app` Helm chart has no `values.schema.json`/generated `README.md`
  and no `helm-unittest` suites yet - see
  [0005-helm-chart-split](decisions/0005-helm-chart-split.md#deliberately-deferred).
- CI still runs against `stack/compose.yml`, not the k3s/Tilt loop - the
  `deploy/helm`/Tiltfile path has no CI coverage at all yet.
