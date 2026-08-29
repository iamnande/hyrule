# roadmap

the current prioritized backlog, grouped by concern. this is a living
document - update it as work lands or priorities shift, don't let it go
stale the way the rest of docs/ almost did before this was written.
[docs/decisions](decisions) records choices already made; this records
what's not decided or not built yet.

## next steps

- close the CI gap on the k3s/Tilt/Helm path - it has zero coverage today,
  everything else on this list builds on top of that path
- give `iam-jwks` a real API - domain/`KeyStore`/postgres repository are
  done, next is the actual JWKS endpoint (initiative 2, ticket 2)
- prove the hyrule -> homelab handoff for real: deploy `pings` to an
  actual homelab cluster, not just the local dev loop
- once CI covers the k3s path, retire `stack/compose.yml` - it's been a
  deliberately deferred duplicate environment since the Rancher Desktop
  pivot
- decide what homelab's first real product service actually is - this
  needs your input, not a guess from this doc
- close the `make bootstrap` gap for `golangci-lint`/`golang-migrate`/`sqlc`
- prove RLS end to end on the first entity that actually has an owner
  (blocked on an identity service existing)

## next 3 initiatives

### 1. CI coverage for the k3s/Tilt/Helm path

nothing about `deploy/helm`, the Tiltfile, or a chart change is verified
by CI today - the four existing jobs (`lint`/`unit`/`integration`/`smoke`)
only exercise `stack/compose.yml`.

1. **spike**: what should "CI-covered" mean here - a real `tilt ci` run
   against a `kind` cluster inside the Actions runner (real Docker there,
   none of the podman/Rancher Desktop friction from local dev), or is
   `helm lint` + `helm template` + rendered-output assertions enough
   without standing up a cluster at all? prototype both, see what actually
   catches a real regression and what it costs in CI time.
2. wire the winning approach into `.github/workflows/ci.yml` as its own
   independent job, same pattern as the existing four.
3. add the `app` chart's missing `helm-unittest` suites as part of the
   same job - closes the gap already on record in
   [0005-helm-chart-split](decisions/0005-helm-chart-split.md#deliberately-deferred).

### 2. iam-jwks: real domain

the spike is done - [0007](decisions/0007-iam-jwks-key-distribution.md):
the source of truth is a 1Password vault, region-local propagation is
the `onepassword-operator`'s job (off the shelf, restarts on rotation),
and `iam-jwks` itself just reads a local mount and caches in memory. no
poll/push/eventing code belongs in this service.

1. ~~domain + `KeyStore` interface + a postgres-backed implementation for
   local dev (`hyrule_app_ro`, strict `SELECT`-only) + migration for the
   keys table~~ - done.
2. the real API: a path in `api/iam-jwks/openapi.yaml`, oapi-codegen
   wiring, handlers, integration tests.
3. the secret-file-backed `KeyStore` implementation, wired in
   `cmd/iam-jwks/main.go` behind whatever config selects it - the
   1Password Connect server + operator themselves are separate
   infrastructure work, tracked under infrastructure below, sequenced
   behind there being a second real region to test against.

### 3. first real homelab deploy of pings

proves the whole pattern - chart, values, the hyrule -> homelab handoff -
actually holds outside the local dev loop. least concretely scoped of the
three on purpose: there's no visibility into homelab's current state from
inside this repo.

1. **spike**: survey homelab's actual current state - is there a real k3s
   cluster running anywhere yet, what would pulling pings' image into it
   require (registry, DNS, ingress), what's missing.
2. whatever the spike surfaces as the blocking gap - most likely getting
   pings' image published somewhere homelab's cluster can reach, and a
   values file for that real environment.
3. deploy, verify for real, and document the hyrule -> homelab handoff as
   it actually happened - probably a new decision doc once it's real
   instead of theoretical.

## by category

### devx

- provision `golangci-lint`/`golang-migrate`/`sqlc` via `mise`, closing
  the `make bootstrap` gap on record in
  [conventions.md#known-gaps](conventions.md#known-gaps)
- lint [docs/style.md](style.md)'s consumer-defined-interfaces rule -
  nothing stops a producer-side interface from creeping in today
- `app` chart's `values.schema.json` + generated `README.md`
  (readme-generator-for-helm), `values.yaml` is already annotated for it
- `helm-unittest` suites for the `app` chart (also initiative 1, ticket 3)
- `make new-service` non-interactive mode (flags instead of prompts), for
  scripting/future CI use

### infrastructure

- CI coverage for the k3s/Tilt/Helm path (initiative 1)
- retire `stack/compose.yml` once CI covers the k3s path
- `hyrule_app_ro` postgres role - strict `SELECT`-only, for consumers that
  should never be able to write regardless of physical topology
- 1Password Connect server + `onepassword-operator` in-cluster, feeding
  `deploy/helm/platform` via `additionalK8sObjects` - see
  [0007](decisions/0007-iam-jwks-key-distribution.md). sequenced behind
  there being a second real region to test replication against, not
  built ahead of that
- error registry: config-driven generator + a docs page per code -
  currently hand-written on purpose, see
  [conventions.md#api-design](conventions.md#api-design)

### platform

- `iam-jwks` real domain + `KeyStore` interface (initiative 2) -
  [0007](decisions/0007-iam-jwks-key-distribution.md) found this isn't
  the service that proves out eventing after all: 1Password's own
  replication plus the `onepassword-operator`'s restart-on-change already
  solves key distribution, no stream needed
- Kafka/Flink/normalized eventing with versioned schemas is still the
  direction for cross-service data movement generally, based on how this
  has been built before - it needs a real service with an actual
  streaming/fan-out problem to get proven out on, not a payload this
  small; nothing currently in flight is that service yet

### product

homelab's actual day-to-day product surface isn't decided, and shouldn't
be guessed at in this doc - what homelab should actually do for you is
the one category here that needs your input before it has real items.

### homelab

- survey homelab's current real state (initiative 3, ticket 1)
- first real deploy of `pings` (initiative 3)
- document the hyrule -> homelab handoff once it's happened for real
