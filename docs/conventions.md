# conventions

the rules for writing code in this repo, why included - read before changing
anything. [architecture.md](architecture.md) has the big picture this all
sits inside; this doc is where topic depth actually lives.

## packages

`internal/lib` (shared, no domain knowledge) vs `internal/svc/<name>` (one
service's own domain/api/data) vs `cmd/<name>` (that service's entrypoint) -
see [architecture.md#service-topology](architecture.md#service-topology) for
the full shape. new service = new `internal/svc/<name>` + `cmd/<name>`,
following an existing one exactly; `internal/lib` shouldn't need to change.

a domain package defines the narrow interfaces it needs from the
services/repositories it depends on, in its own file - never imports their
concrete types. that's the whole trick that makes a hand-written fake enough
for a unit test (see testing, below) - there's no concrete dependency to
mock, only an interface to satisfy.

## fx

a service's own `Module` is its logic and wiring only - config loading
happens at the composition root (`cmd/<name>/main.go`), not inside the
service module. keeping those separate is what lets tests supply fixed
values directly instead of overriding real ones.

no `fx.Decorate`. it only exists to override a value something upstream
already provides - if you're reaching for it, that's a sign the value is
being provided somewhere it shouldn't be, not a reason to intercept it.
fix the source instead.

validate config where it's loaded (the `Load*()` function), returning a
plain error - fx already surfaces provider errors natively, that's real
infrastructure, use it. no custom `encoding.TextUnmarshaler` or other
stdlib-interface hooks bolted on just to get validation to fire.

health API and router are free for every service via `runtime.NewModule` -
a service's own `Module` only supplies what's actually service-specific
(`health.Probes`, its own `group:"apis"` handlers). `runtime.HTTPModule`
(the `NewHealthAPI`/`NewRouter` providers) is split out from `NewModule`
so tests can pull in a working `http.Handler` without also getting
`fx.Invoke(rest.NewServer)`, which would bind a real port.

## types

enum-shaped values (`Environment`, `Region`, ...) get their own named string
type, used end to end in struct fields - never `string` with a cast at the
boundary. Go marshals a named string type to JSON exactly like a plain
string, so there's no cost to keeping the type all the way through. casts
are only acceptable where the consumer hard-requires `string` (`slog.String`,
a `map[string]string`) - never because a struct field was typed lazily.

## API design

**APIs are pure i/o.** a handler decodes a request, calls the domain, encodes
a response. there must be no business logic in the API layer - if you're
writing an `if` statement in a handler that isn't about http (status codes,
decoding errors), it's in the wrong layer.

**OAS-first.** the OpenAPI spec is the contract, not a byproduct of it.
types and validators are generated from it (`oapi-codegen`); handlers
implement the generated interface. the spec is written and reviewed before
the handler exists.

## URL structure & versioning

```
domain: api.morethq.com
path:   /<entity>
header: Hyrule-Version: 2026-08-22
```

the URL identifies the resource - it should describe *what*, not which
binary serves it today or which contract snapshot the client wants. version
lives in a header, not the path: Stripe (`Stripe-Version`) and ngrok
(`Ngrok-Version`) both do this, and a resource's address shouldn't change
just because its contract did.

**version is scoped per service, not per entity.** a service ships one
version for its entire surface, covering every entity it owns. an identity
service owning `account`/`workspace`/`user`/`team` bumps once, together - a
client pinned to a version gets one coherent view of the whole service, not
four independently-drifting numbers to reconcile. this only stays humane if
most changes are additive and need no bump at all; a genuine breaking change
earns a new version, and a transform/shim layer translates old pins to the
current internal shape, so a break in `user` doesn't actually break a client
still reading `team` on an older pin.

version format is dated (`2026-08-22`), not an incrementing integer - a date
doubles as an implicit changelog entry, a bare integer doesn't.

the tell to split into a *new service* isn't "it has two entities" - it's
when entities stop sharing a bounded context (data ownership, operational
lifecycle, team ownership) and are only colocated for convenience.

**relationships stay flat by default.** reference the related resource by a
field, filter with a query param (`/v1/pings?filter[id]=X`, matching
Stripe's `/v1/charges?filter[customer_id]=X`). nest a resource under its
parent only for true composition - the child has no identity or lifecycle
apart from the parent (e.g. `/v1/customers/{id}/sources`).

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

span and tag names: `<component>.<action>[.<qualifier>]`, dots only - e.g.
`dependencies.check.all`, `dependencies.check.database`. one separator, not
whatever felt right at the call site - colons snuck in early and it was a
mess.

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
  [docker/postgres/init](../docker/postgres/init) on first container start;
  migrations still run as the owner role (`POSTGRES_USER`, see
  `DATABASE_URL` in `mk/database.mk`).
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

migrations live in [migrations](../migrations) at the repo root, applied
with `make db-migrate-up` (`make db-migrate-create NAME=...` to scaffold
one) - see [mk/database.mk](../mk/database.mk). `sqlc` generation isn't
wired up yet; that lands with the first real table.

## testing

`test-unit` (`./internal/...`) vs `test-integration` (`./tests/...`, against
the real local stack). anything that can't be faked (RLS, above) is
integration-only, never mocked - it's enforced by the postgres engine
itself.

ginkgo/gomega everywhere, including domain/service unit tests - colocated as
`<file>_test.go` + one `suite_test.go` per package, not just for integration
specs. no mocking framework: a domain's own narrow interfaces (see packages,
above) make a hand-written fake enough.

## config

no `.env`, ever. real environment variables, sane `envDefault` tags cover
local dev. local-only values (compose stack credentials) get hardcoded
where they're used - they aren't secrets. real secrets get injected at
invocation time (e.g. `op run -- make run`), never touch disk.

## container image

`cmd/Dockerfile`'s final stage is pinned to `gcr.io/distroless/base-debian12`,
not `gcr.io/distroless/base` or `-static`. `base` (no suffix) was verified to
ship a busybox toolkit baked into its filesystem on the digest tested here -
the opposite of what "distroless" is for. `-static` has no CA certificates,
which breaks outbound HTTPS (Sentry). `-debian12` is the one that's actually
both minimal and functional.

`TARGETOS`/`TARGETARCH` (not `TARGET_OS`/`TARGET_ARCH`) because those are
buildkit/buildah-reserved arg names, auto-populated from `--platform` -
renaming them means losing that for a readability preference. their default
values match this repo's usual dev machine, not a claim that's the only
target.

the final-stage `COPY --from=build` destination (`/service`) is a fixed
path, not `${SERVICE_NAME}` - the exec-form `ENTRYPOINT` below it doesn't get
build-arg substitution, so a parameterized path there would silently break.

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

errors wrapped with `%w` and enough context to debug without a stack trace;
structured logging via `slog`; sentry spans around domain-level operations.
none of this is enforced by tooling yet, so it's easy to drift from without
noticing - this is the reminder.

## known gaps

- no committed `.golangci.yml` - `test-lint` runs default rules by accident,
  not a chosen ruleset.
- no CI - `test-unit`/`test-integration`/`test-lint` all run locally, on
  faith.
- `golangci-lint` and `golang-migrate` aren't provisioned by `make
  bootstrap` - installed locally on faith (e.g. `brew install golangci-lint
  golang-migrate`), same as the two gaps above.
- [docs/style.md](style.md)'s consumer-defined-interfaces rule isn't
  linted - nothing stops a producer-side interface from creeping in.
