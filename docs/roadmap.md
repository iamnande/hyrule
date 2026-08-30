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

## next 3 initiatives

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

### 2. iam-jwks: real domain - done

[0007](decisions/0007-iam-jwks-key-distribution.md): the source of truth
is a 1Password vault, region-local propagation is External Secrets
Operator's job (see [0011](decisions/0011-secrets-generalized.md) - off
the shelf, restarts on rotation via a paired reloader), and `iam-jwks`
itself just reads a local mount and caches in memory. no poll/push/
eventing code belongs in this service.

1. ~~domain + `KeyStore` interface + a postgres-backed implementation for
   local dev (`hyrule_app_ro`, strict `SELECT`-only) + migration for the
   keys table~~ - done.
2. ~~the real API: `GET /.well-known/jwks.json`, EdDSA/Ed25519 JWKs (RFC
   8037), oapi-codegen wiring, handlers, integration tests~~ - done.
3. ~~the secret-file-backed `KeyStore` implementation, wired in behind
   config~~ - done: `go/cmd/iam-jwks/app.Module(fileCfg)` picks
   `svc.WithFileStore()` when `HYRULE_IAM_JWKS_KEYS_FILE_PATH` is set,
   `svc.WithPostgres()` otherwise - no database wiring at all in
   file-backed mode. the 1Password Connect server + External Secrets
   Operator themselves are still separate infrastructure work, tracked
   under infrastructure below, sequenced behind there being a second real
   region to test against.

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

### 4. stand up Harbor + 1Password Connect + ESO, local and CP

[0010](decisions/0010-image-registry-and-publishing.md) and
[0011](decisions/0011-secrets-generalized.md) decided the shape - Harbor
everywhere (local push/pull too, not just CP), 1Password Connect + External
Secrets Operator everywhere too, no cloud-managed stand-ins for either.
neither is built yet; both need real credentials/infrastructure a decision
doc can't supply. least concretely scoped of the four initiatives on
purpose, same reason as initiative 3.

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

- `iam-jwks` real domain + `KeyStore` interface (initiative 2) -
  [0007](decisions/0007-iam-jwks-key-distribution.md) found this isn't
  the service that proves out eventing after all: 1Password's own
  replication plus External Secrets Operator + a reloader's
  restart-on-change ([0011](decisions/0011-secrets-generalized.md))
  already solves key distribution, no stream needed
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
