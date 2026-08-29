# roadmap

the current prioritized backlog, grouped by concern. this is a living
document - update it as work lands or priorities shift, don't let it go
stale the way the rest of docs/ almost did before this was written.
[docs/decisions](decisions) records choices already made; this records
what's not decided or not built yet.

## next steps

- close the CI gap on the k3s/Tilt/Helm path - it has zero coverage today,
  everything else on this list builds on top of that path
- give `iam-jwks` real behavior, and get a real answer (not a guess) on
  whether a non-primary deployment needs poll, push, or eventing to stay
  current
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

### 2. iam-jwks: real domain + the sync-mechanism build-out

the scaffold exists, nothing behind it does. this is also where the
platform/eventing pattern actually gets built, not just discussed.

1. **spike**: the problem space is "how does a non-primary deployment stay
   current on key rotation" - poll on an interval, push from the
   authoritative writer, or eventing. pick the shape and write down why,
   plus the read/write role-separation question from the jwks
   conversation (a `hyrule_app_ro` role, strict `SELECT`-only). output is
   a decision doc, not code.
2. domain + `KeyStore` interface + postgres-backed repository + migration
   for the keys table, following `internal/svc/pings` as the reference.
3. the real API: a path in `api/iam-jwks/openapi.yaml`, oapi-codegen
   wiring, handlers, integration tests.

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
- error registry: config-driven generator + a docs page per code -
  currently hand-written on purpose, see
  [conventions.md#api-design](conventions.md#api-design)

### platform

- `iam-jwks` real domain + `KeyStore` interface (initiative 2)
- the sync-mechanism spike (initiative 2, ticket 1) picks the actual
  shape - poll, push, or eventing - and writes down why
- Kafka/Flink/normalized eventing with versioned schemas is the direction
  for cross-service data movement generally, based on how this has been
  built before - `iam-jwks` is where that pattern gets proven out first,
  scoped to the one real problem in front of it rather than built as
  platform infrastructure up front

### product

homelab's actual day-to-day product surface isn't decided, and shouldn't
be guessed at in this doc - what homelab should actually do for you is
the one category here that needs your input before it has real items.

### homelab

- survey homelab's current real state (initiative 3, ticket 1)
- first real deploy of `pings` (initiative 3)
- document the hyrule -> homelab handoff once it's happened for real
