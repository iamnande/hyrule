# roadmap

the current prioritized backlog, grouped by concern. this is a living
document - update it as work lands or priorities shift, don't let it go
stale the way the rest of docs/ almost did before this was written.
[docs/decisions](decisions) records choices already made; this records
what's not decided or not built yet.

## next steps

- close the CI gap on the k3s/Tilt/Helm path - it has zero coverage today,
  everything else on this list builds on top of that path
- prove the hyrule -> homelab handoff for real: deploy `pings` to an
  actual homelab cluster, not just the local dev loop
- decide what homelab's first real product service actually is - this
  needs your input, not a guess from this doc
- close the `make bootstrap` gap for `golangci-lint`/`golang-migrate`/`sqlc`
- prove RLS end to end on the first entity that actually has an owner
  (blocked on an identity service existing)
- spike proto-to-OAS contract strategy once a real gRPC service exists -
  conventions.md is OAS-first today (spec is the contract, Go types
  generated from it); folding gRPC in without deciding a source of truth
  risks two contracts drifting. no gRPC service exists yet to prove a
  toolchain (e.g. grpc-gateway + protoc-gen-openapiv2) against, so this
  stays a spike, not a decision, until one does

## decisions pending write-up

already decided (in conversation, not yet in `docs/`) - each just needs
the actual doc written, not more deciding. tracked separately from `next
steps` above since none of this is open backlog, it's doc debt on top of
settled direction.

- public/internal API visibility (new decision doc): only `api`, `cli`
  (via `api`), `bff`, and `mcp` get public ingress - every other service
  (`pings`, `iam-jwks`, whatever's next) is internal-only, cluster-local
  or ngrok-reachable by exception, decided per service as it comes up.
  each service gets its own `api/<service>/{public,internal}/` contracts
  (populated with only what that service actually needs - most will only
  ever have `internal/`), and isolated routers in the Go API layer
  (`InternalXyzAPI` vs `XyzAPI`, not one router serving both) so that
  authz/scopes/ABAC lands later without having to split an
  already-mixed router apart. `iam-jwks` is the first concrete case:
  `api` hits its JWKS endpoint behind the scenes, it gets no public
  endpoint of its own - possibly gRPC instead of HTTP once the dual
  listener below exists, still open
- standardized dual listener (new decision doc): every service gets
  HTTP `:8000` always, gRPC `:9000` opt-in - scoped to the
  transport-layer plumbing only (`runtime.NewModule`, the `app` chart,
  per-port probes). does not decide the proto-to-OAS contract-generation
  spike above; `iam-jwks` (internal-only, see above) is a plausible
  first gRPC-only candidate once this exists
- resource-sizing addendum: no bin-packing or affinity rules yet, same
  static-sizing reasoning [0018](decisions/0018-resource-sizing-autoscaling.md)
  already used for HPA/VPA - homelab-scale, not enough real multi-node
  topology or replica counts for either to matter. revisit with real
  evidence, not before
- fx constructor convention, into conventions.md: a Params struct
  (`fx.In`) over positional args, even at one or two params today -
  forward-looking, no downside at the current arg counts
- module.go convention, into conventions.md: a service's `module.go`
  holds only the `fx.Module(...)` declaration, nothing else - every
  other constructor (the domain wiring, the API handler wiring) moves to
  its own named file. already applied in code - pings' `newRegistry`
  moved to `registry.go`, `newAPIHandler`/`apiHandlerResult` to
  `handler.go`; iam-jwks' `newKeySet` to `keyset.go`, same `handler.go`
  split
- domain/repository naming convention, into conventions.md: name a
  domain's core type for what it does, not `Service`; name a repository
  for its backend once more than one is real or anticipated. already
  applied in code - `pings.domain.Service` -> `Registry`,
  `iam-jwks.domain.Service` -> `KeySet`, `pings.repository.Repository`
  -> `PostgresRepository` (matching `iam-jwks`'s existing
  `EnvRepository`)

## open decisions (pre-ArgoCD gap review)

stubs, not decisions yet - context and the open question only, each to be
worked and filled in in order. roughly sequenced: the deployment-shape
questions first (later ones depend on the answers), then the ones that
don't block a deploy but have sat undecided regardless.

1. ~~[0009: environments & promotion model](decisions/0009-environments-and-promotion.md)~~ - decided.
2. ~~[0014: namespacing convention](decisions/0014-namespacing.md)~~ - decided.
3. ~~[0010: image registry & publishing](decisions/0010-image-registry-and-publishing.md)~~ - decided.
4. ~~[0011: secrets, generalized](decisions/0011-secrets-generalized.md)~~ - decided.
5. ~~[0012: migrations under GitOps](decisions/0012-migrations-under-gitops.md)~~ - decided.
6. ~~[0023: postgres replication topology](decisions/0023-postgres-replication-topology.md)~~ -
   decided (emerged out of 0012's discussion, numbered after the fact).
7. ~~[0013: ingress, DNS, TLS](decisions/0013-ingress-dns-tls.md)~~ - decided.
7a. ~~[0024: cross-region private networking](decisions/0024-cross-region-networking.md)~~ -
    decided (ngrok, same mechanism as 0013 - not a separate VPN mesh).
8. ~~[0015: authN/authZ model](decisions/0015-authn-authz.md)~~ - decided.
9. ~~[0016: observability backend](decisions/0016-observability-backend.md)~~ - decided.
10. ~~[0017: recurring / scheduled workloads](decisions/0017-scheduled-workloads.md)~~ - decided.
11. ~~[0018: resource sizing & autoscaling policy](decisions/0018-resource-sizing-autoscaling.md)~~ - decided.
12. ~~[0019: postgres backup & DR](decisions/0019-postgres-backup-dr.md)~~ - decided.
13. ~~[0020: hyrule -> homelab versioning](decisions/0020-hyrule-homelab-versioning.md)~~ - decided.
14. ~~[0021: shell script style](decisions/0021-shell-script-style.md)~~ - decided.
15. ~~[0022: Helm / YAML style](decisions/0022-helm-yaml-style.md)~~ - decided.

## initiatives

### 1. CI coverage for the k3s/Tilt/Helm path

nothing about `deploy/helm`, the Tiltfile, or a chart change is verified
by CI today - `integration`/`smoke` only need a standalone postgres
(`make db-up`, see [0008](decisions/0008-repo-tree-layout.md)), not a
real deploy.

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

### 2. bff steel-thread

a new service backing the homelab UI/dashboard - browser talks only to
it, never to `pings` or any other internal service directly (see
[decisions pending write-up](#decisions-pending-write-up) above). proves
the first real inter-service call in this repo: a generated client
against `pings`'s internal contract, plus the call convention itself
(base URL via [0014](decisions/0014-namespacing.md)'s namespace DNS,
timeouts, tracing/logging context propagation) - none of which exists
yet. worked in its own thread when picked up, not bundled with anything
else on this list.

1. the two pending decision docs above (public/internal API visibility,
   dual listener) block this - `pings` needs its internal contract split
   out before there's anything to generate a client against.
2. `make new-service` the `bff` service; no domain/API scaffolded yet,
   same as any new service.
3. the inter-service call convention itself - base URL resolution,
   timeouts, context propagation - doesn't exist anywhere in this repo
   today, this is the first consumer.
4. the generated client against `pings`'s internal contract, wired into
   `bff`'s own domain layer.
5. demo the steel-thread end to end: browser -> `bff` -> `pings`,
   nothing else in the path yet.

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

### 4. security adversarial review of architecture + roadmap

sequenced deliberately: after initiative 3, because that's the first
point there's real homelab infrastructure to review against instead of
local-only/theoretical; before initiative 5, because Harbor introduces a
self-hosted GitHub Actions runner with network access to both GitHub and
Harbor plus real secret flow (1Password Connect + ESO) - exactly where a
security gap stops being theoretical and starts having real blast
radius. findings here should shape how initiative 5 gets built, not get
retrofitted onto it afterward.

covers the architecture and this roadmap directly, not just one
component - RBAC/namespace boundaries ([0014](decisions/0014-namespacing.md)),
secrets handling ([0011](decisions/0011-secrets-generalized.md)),
ingress/edge trust ([0013](decisions/0013-ingress-dns-tls.md),
[0015](decisions/0015-authn-authz.md)), and whether cluster-level policy
enforcement (Kyverno or similar - see
[infrastructure](#infrastructure) below) is actually needed and what it
should enforce, rather than guessed at ahead of the review.

### 5. stand up Harbor + 1Password Connect + ESO, local and CP

[0010](decisions/0010-image-registry-and-publishing.md) and
[0011](decisions/0011-secrets-generalized.md) decided the shape - Harbor
everywhere (local push/pull too, not just CP), 1Password Connect + External
Secrets Operator everywhere too, no cloud-managed stand-ins for either.
neither is built yet; both need real credentials/infrastructure a decision
doc can't supply. least concretely scoped of these initiatives on
purpose, same reason as initiative 3 - also gated behind initiative 4
now, not just under-scoped.

1. Harbor: install via its own chart in the local cluster and CP
   (`us-west-2`), sort out TLS trust for local `nerdctl` push/pull against
   it, rewire [Tiltfile](../Tiltfile)'s `build_service`/`deploy_service` to
   push/pull through it instead of the direct containerd store, add
   `imagePullSecrets` wiring to real services' values.
2. the self-hosted GitHub Actions runner that gives CI a private path to
   Harbor - provisioning, patching story, network access to both GitHub
   and Harbor.
3. 1Password Connect server + External Secrets Operator: install in the
   local cluster and CP, seed the Connect token bootstrap secret by hand,
   stand up a `ClusterSecretStore`, and give `iam-jwks` (now env-var-backed
   only, see [0011](decisions/0011-secrets-generalized.md)) a real
   `ExternalSecret` sourcing `HYRULE_IAM_JWKS_KEYS` instead of the
   hardcoded local placeholder currently in
   [deploy/values/iam-jwks/values.yaml](../deploy/values/iam-jwks/values.yaml).
4. pair ESO with a reloader (e.g. Stakater's `Reloader`) so a rotated
   secret still restarts the pod consuming it.

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
- `app` chart's `job`/`cronjob` templates - shape decided in
  [0017](decisions/0017-scheduled-workloads.md), build when a real
  service needs one; replaces the single `_pod.tpl`/`app.pod` partial
  with `app.pod.service`/`app.pod.job` (plus a shared `app.pod.container`)
  instead of threading `mode` conditionals through one partial
- `shellcheck` as a step in CI's `lint` job - decided in
  [0021](decisions/0021-shell-script-style.md), not wired in yet

### infrastructure

- Kyverno (or similar) for cluster-level policy enforcement / admission
  control - no admission control exists in the cluster today. what it
  should actually enforce (and whether Kyverno specifically is the right
  tool) is a question for initiative 4's security review, not decided
  ahead of it
- CI coverage for the k3s/Tilt/Helm path (initiative 1)
- `hyrule_app_ro` postgres role - strict `SELECT`-only, for consumers that
  should never be able to write regardless of physical topology
- nightly `pg_dump` CronJob for CP's postgres - shape decided in
  [0019](decisions/0019-postgres-backup-dr.md), blocked on two things:
  an actual off-cluster storage target (self-hosted MinIO, Backblaze B2,
  or similar - not chosen yet), and a real deploy mechanism for
  standalone infra `CronJob`s in CP at all - no ArgoCD instance exists
  yet to write a `PreSync`-hook-style manifest against
  ([0012](decisions/0012-migrations-under-gitops.md)), same gap the
  `migrate` Job has
- 1Password Connect server + External Secrets Operator in-cluster,
  feeding `deploy/helm/platform` via `additionalK8sObjects` - see
  [0007](decisions/0007-iam-jwks-key-distribution.md) and
  [0011](decisions/0011-secrets-generalized.md). sequenced behind there
  being a second real region to test replication against, not built
  ahead of that
- error registry: config-driven generator + a docs page per code -
  currently hand-written on purpose, see
  [conventions.md#api-design](conventions.md#api-design)
- Prometheus + Grafana + Alertmanager + Loki + Tempo in-cluster, local
  and CP - shape decided in
  [0016](decisions/0016-observability-backend.md), install itself
  (chart/values, local vs CP vs both, alert rules) not started;
  `platform` grows the `ServiceMonitor`/`PrometheusRule`/log-shipper
  plumbing per that decision
- migrate `go/internal/lib/tracing` off `sentry-go` onto OpenTelemetry
  (OTLP export to Tempo) - see [0016](decisions/0016-observability-backend.md);
  span-naming convention ([conventions.md#tracing](conventions.md#tracing))
  is unaffected, only the backend underneath the helpers changes

### platform

- `bff` steel-thread (initiative 2)
- Kafka/Flink/normalized eventing with versioned schemas is still the
  direction for cross-service data movement generally, based on how this
  has been built before - it needs a real service with an actual
  streaming/fan-out problem to get proven out on, not a payload this
  small; nothing currently in flight is that service yet. `iam-jwks`
  turned out not to be that service - 1Password's own replication plus
  External Secrets Operator + a reloader's restart-on-change
  ([0011](decisions/0011-secrets-generalized.md)) already solves key
  distribution, no stream needed

### product

homelab's actual day-to-day product surface isn't decided, and shouldn't
be guessed at in this doc - what homelab should actually do for you is
the one category here that needs your input before it has real items.

### homelab

- survey homelab's current real state (initiative 3, ticket 1)
- first real deploy of `pings` (initiative 3)
- document the hyrule -> homelab handoff once it's happened for real
