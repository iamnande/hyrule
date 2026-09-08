# PRD - ArgoCD + GitOps foundation (scalable shape, proven locally)

## Problem

`docs/roadmap.md`'s "pre-ArgoCD gap review" list (0009-0022) is fully decided. no ArgoCD
instance, manifest, or CLI exists anywhere in the repo. decisions
[0009](../../decisions/0009-environments-and-promotion.md),
[0012](../../decisions/0012-migrations-under-gitops.md),
[0017](../../decisions/0017-scheduled-workloads.md), and
[0019](../../decisions/0019-postgres-backup-dr.md) all defer their ArgoCD-shaped details to
"when ArgoCD exists." nothing today takes a green `main` and rolls it anywhere beyond a
laptop's Tilt loop. that's the actual foundation gap standing between here and a real product
service.

a first draft of this proposal wrote that shape as one hand-authored `AppProject` +
`ApplicationSet` pair per service - fine for proving the mechanic against `pings` alone, but a
real liability heading toward the ~60 services this repo expects to carry: that shape is
`O(services)` hand-maintained ArgoCD YAML, exactly the kind of per-unit config growth this
proposal should be closing off before it exists, not walking into deliberately. this revision
settles the shape that actually scales - a services registry driving generated `AppProject`s and
a single generator-based `ApplicationSet` - and proves it locally against the one service that
exists today, the same way the rest of this proposal proves ArgoCD's mechanics without real VPS
spend or homelab hardware.

no homelab hardware exists yet. that's a separate gap from ArgoCD itself - 0009 already splits
`prod` into CP/DP (control-plane/data-plane) regions on VPS ("VPS to start... homelab hardware
later"), independent of physical homelab existing. this proposal targets the **local** Rancher
Desktop/k3s cluster only. it proves ArgoCD's mechanics - Application sync, the PreSync migrate
hook, the generator shape - without real VPS spend or homelab hardware.

## Solution

1. install ArgoCD via its own Helm chart into the local cluster, in its own `argocd` namespace,
   via new `mk/argocd.mk` (`argocd-up`/`argocd-down`). same pattern as `mk/cluster.mk`.
2. add a services registry, `local/argocd/appprojects/services.yaml` - one entry per service (`service`,
   `team`, `domain`), the single source of truth for "what's live in ArgoCD, owned by whom."
   decoupled from `deploy/values/<service>/` existing: a service can be scaffolded without being
   registered - an explicit deploy gate, not implicit-by-directory (Tilt's own directory-listing
   discovery stays exactly as-is, a deliberate difference - see § Configuration).
3. generate one `AppProject` per registry entry - never hand-written - via a single bootstrap
   `Application` sourcing a small local chart that ranges over the registry (§ Interfaces in the
   tech spec). project scope stays keyed to the service, not team/domain: team ownership changes
   far more often than a service does, so project identity (name, destinations, RBAC scoping)
   stays stable while team/domain live as labels, ready to drive a generated RBAC-policy mapping
   later without ever renaming or restructuring a project (see § Scope & Non-Goals).
4. write a single `ApplicationSet` for all services, using a `matrix` generator crossing the
   registry (`git` generator, file mode) with a `clusters` generator matched on region labels
   (`region: us-west-2` / `region: us-east-2`, per 0009's naming) - one generated `Application`
   per service × matched cluster. today that's exactly one match (`pings` × the local cluster),
   sized so both a new service and a new region are additive (a registry line, a labeled cluster
   Secret) rather than hand-written Application manifests. region and environment identity
   (`HYRULE_REGION`/`HYRULE_ENVIRONMENT`) are injected into each generated Application from the
   matched cluster's own labels, never hardcoded in a service's committed values file (per
   [0026](../../decisions/0026-region-environment-argo-derived.md) - narrows 0009's overlay-file
   shape for these two keys specifically). sync the generated Application into `pings` - the same
   namespace Tilt targets today - matching how ArgoCD actually runs in every non-local
   environment, standalone with no Tilt in the picture at all. locally that means Tilt's `pings`
   resource is stopped before a sync, so there's exactly one controller writing to that namespace
   at a time, not two sharing it.
5. move the migrate Job into `deploy/helm/app` itself as an opt-in, PreSync-hook-annotated
   template, per [0025](../../decisions/0025-migrate-as-chart-primitive.md) - narrows 0017's
   "migrate lives outside app/platform" framing, which was reasoning about a single
   shared-postgres one-off, not "any of ~60 services can opt into owning its own migration step."
   the opt-in is declaring a `datastores` entry (`flavor: postgres`) in a service's own values, not
   a standalone flag - a service engineer never supplies a connection string or a ConfigMap name;
   postgres connection facts live once, as a chart default, and the Job derives everything else
   (database name, ConfigMap name) from the entry itself and chart-native naming. each opted-in
   service targets its own logical database on the one shared postgres instance for now (physical
   topology/replication stays [0023](../../decisions/0023-postgres-replication-topology.md)'s
   still-open question, not decided here - see § Scope & Non-Goals). Tilt unifies onto the same
   template (`local_migrate()` renders it via `helm template` instead of a hand-maintained
   duplicate); `local/cluster/migrate-job.yaml` is retired.
6. pull images from the local containerd store Tilt's `nerdctl_build` already populates
   (`imagePullPolicy: IfNotPresent`). this is an explicit local-only stand-in for Harbor
   (initiative 5, not built yet, gated behind initiative 4).
7. add "ArgoCD + GitOps foundation" as its own numbered initiative on `docs/roadmap.md`. today
   it's referenced only as a dependency gate under `infrastructure`, not a first-class
   initiative like 1-5.

## Configuration

new manifests live under `local/argocd/` (the services registry, the generated-`AppProject`
chart + its wrapper Application, the services `ApplicationSet`), per
[0008](../../decisions/0008-repo-tree-layout.md)'s `deploy/` vs `local/` split - the generator
shape itself is dev-loop-only plumbing, not yet exercised against real CP/DP clusters. the one
deliberate exception: the migrate Job template moves *into* `deploy/helm/app`
([0025](../../decisions/0025-migrate-as-chart-primitive.md)) - it's a real chart primitive any
service can opt into everywhere, not a prototype-only artifact, so it belongs where the rest of
the chart lives, not under `local/`.

`local/argocd/appprojects/services.yaml` and Tilt's own service discovery (`ls deploy/values`, unchanged)
answer different questions on purpose - Tilt asks "what's scaffolded," the registry asks "what's
live in ArgoCD, owned by whom." a service can exist in `deploy/values/` without a registry entry;
Tilt still deploys it locally, ArgoCD just doesn't know about it yet.

`pings`' own `deploy/values/pings/values.yaml` loses its `HYRULE_REGION`/`HYRULE_ENVIRONMENT`
keys - both become Argo-derived (0026) - and its `migrate:` block becomes a `datastores` entry
(`slug: pings`, `flavor: postgres`), with no connection string of any kind. `iam-jwks` is
unchanged (still unregistered, see § Scope & Non-Goals).

postgres connection facts (host, port, owner role/credentials) move to a single chart default in
`deploy/helm/app/values.yaml` - the one file a platform engineer edits to repoint every
`datastores`-opted-in service at once, no per-service values touched, no generated code. same
mechanism 0026 already established for region/environment (a chart default an Argo-managed
Application can override, that a service's own committed values never sets).

## States

- **Application sync status:** `OutOfSync` → `Syncing` → `Synced`
- **Application health:** `Progressing` → `Healthy` / `Degraded`
- **PreSync hook Job:** `Pending` → `Running` → `Succeeded`, deleted and recreated fresh on
  every sync (`BeforeHookCreation`). same problem
  [conventions.md#local-orchestration](../../conventions.md#local-orchestration) already flags
  for the Tilt path (Jobs are immutable once created)
- **service registration:** unregistered (no `services.yaml` entry) → registered, no `AppProject`
  yet → `AppProject` generated (bootstrap Application synced) → `Application` generated (services
  `ApplicationSet` matched it against a labeled cluster) → `OutOfSync` until first manual sync

## Behavior

**before:** only Tilt's manual-trigger loop deploys locally, owning the `pings` namespace. no
GitOps mechanism exists anywhere in the repo.

**after:** Tilt's `pings` resource is stopped. `pings` has a `local/argocd/appprojects/services.yaml` entry,
which generates its `AppProject` and, crossed with the local cluster's `region` label, its
`Application`. pushing a values change to the prototype's feature branch and triggering a manual
sync rolls the change into the same `pings` namespace Tilt used, now owned by ArgoCD instead. the
migrate hook - now `app`'s own opt-in template - reruns as a PreSync step before the Deployment
updates, targeting `pings`' own logical database. adding a second service means one more
`services.yaml` line and a `deploy/values/<service>/` directory - no new ArgoCD YAML.

## API

N/A - no new external API surface. ArgoCD's own UI/API is port-forwarded locally for observing
sync/health state. it is not a new hyrule-owned service.

## Lifecycle

1. `make argocd-up` installs ArgoCD's chart into the `argocd` namespace, applies
   `cluster-local.yaml` (labels the in-cluster destination `region: us-west-2`,
   `environment: local`), the bootstrap Application that generates `AppProject`s from
   `services.yaml`, and the services `ApplicationSet`.
2. syncing the bootstrap Application (manual) creates `pings`' `AppProject` from its
   `services.yaml` entry.
3. the services `ApplicationSet` has already generated `pings`' `Application` (matrix generator,
   registry × the local cluster's label match) - it sits `OutOfSync` until `pings`' `AppProject`
   exists (step 2) and a manual sync is triggered (R7).
4. Tilt's `pings` resource is stopped, then a sync (manually triggered) runs the PreSync migrate
   hook and applies the `app-platform` release into the existing `pings` namespace.
5. `make argocd-down` tears ArgoCD down wholesale. Tilt's `pings` namespace and resources stay
   untouched throughout.

## Edge Cases

- the migrate hook reruns on every sync even when its migrations content is unchanged. accepted
  as a known no-op cost, consistent with 0012's per-sync hook design, not solved differently
  here.
- CP/DP dual-region migrate ordering (0023's already-flagged sharp edge) is **not** exercised by
  this pass. only one Application, no second region.
- a service's generated `Application` can exist (via the `ApplicationSet`) before its `AppProject`
  does (via the separately-synced bootstrap Application) - ArgoCD accepts the `Application`
  object either way but leaves it in a validation error state until the referenced project
  exists. bootstrap order matters at sync time, not apply time (R2/R3, see tech spec § Core).
- Tilt and ArgoCD are never pointed at `pings` at the same time - Tilt's `pings` resource is
  stopped before a sync, so there's one controller at a time, not simultaneous ownership
  survived under load. matches the real target shape (ArgoCD standalone, no Tilt in non-local
  environments) instead of proving a coexistence scenario that won't exist for real.
- first sync adopting `pings`' existing, Tilt-created resources (Tilt's own `helm_resource` call
  leaves a Helm release secret behind in the namespace) is unverified going in - this pass is
  what proves whether ArgoCD's own apply cleanly takes ownership or needs the namespace wiped
  first.
- image freshness: ArgoCD doesn't rebuild images here, so a stale local image tag won't reflect
  new code. this pass proves manifest/config sync and the hook mechanic, not a full
  build+deploy loop.
- the PreSync hook is a provisional shape, not a settled one. a full Application sync always
  runs it, but a selective/resource-targeted sync (`argocd app sync pings --resource ...`) skips
  hooks entirely - 0012 didn't account for that. not solved here.

## Observability

ArgoCD's own UI/API (port-forwarded) shows sync and health status. no new logging, metrics, or
alerting is added to hyrule itself in this pass.

## Scope & Non-Goals

**in scope:**
- ArgoCD installed locally via its own Helm chart + new `mk/argocd.mk` targets
- `local/argocd/appprojects/services.yaml`, a services registry (service, team, domain) driving everything
  below - registering a service is a registry line, not new ArgoCD YAML
- generated (not hand-written) per-service `AppProject`s from the registry
- a single services `ApplicationSet`, `matrix` generator (registry × `clusters` generator on
  region labels), one generated `Application` per service × matched cluster
- region/environment identity Argo-derived per [0026](../../decisions/0026-region-environment-argo-derived.md),
  never committed to a service's values file
- the migrate Job converted to an opt-in `app` chart primitive per
  [0025](../../decisions/0025-migrate-as-chart-primitive.md), unified with Tilt's own local
  migrate path, targeting each opted-in service's own logical database
- a `datastores` values shape (`slug`, `flavor`, flavor-specific keys) a service declares to opt
  into schema management - `flavor: postgres` is the only one built this pass; other flavors
  (e.g. `key-value`) are shape-only, additive later, not exercised here
- postgres connection facts (host, port, owner role/credentials) as a single chart default, never
  a per-service value - a service supplies no DSN, no ConfigMap name, no connection detail of any
  kind
- a new roadmap initiative entry

**out of scope:**
- `iam-jwks` coverage/registration. no new mechanics to prove beyond `pings`. picked up as
  trivial follow-up once the pattern holds - registering it is one `services.yaml` line
- registering real CP/DP clusters with the `clusters` generator. no VPS clusters exist to
  register yet - the generator's shape doesn't need to change when they do, but that
  registration is a separate initiative once VPS provisioning happens, not guessed at here
- generating an RBAC-policy mapping (`argocd-rbac-cm`) from the registry's `team`/`domain`
  labels. the labels exist so this is additive later, but there's exactly one operator today -
  no multi-team access boundary to enforce yet. not built this pass
- RLS, and postgres RBAC/role design beyond "migrate connects with the existing shared owner
  role, against a per-service logical database name." table-stakes eventually, but designing it
  now would mean designing it without an identity service to test it against - the one
  forward-looking constraint taken now is that the migrate template's database name is a
  per-service value (the `datastores` entry's own `slug`), not hardcoded, so this lands
  additively rather than as a rewrite
- real secret management for the postgres owner credentials. they stay a plaintext chart default
  this pass - the same risk profile as today's already-committed `hyrule_owner`/`hyrule_owner`,
  just centralized to one file instead of duplicated per service - not a regression, but not a
  fix either. [0011](../../decisions/0011-secrets-generalized.md) is where that's actually solved.
  TLS on the database connection itself, including locally, is a related direction raised during
  this pass's own design review but not decided or built here - see `docs/roadmap.md`'s TLS
  next-step entry
- physical-vs-logical postgres topology, replication, and per-service instance-vs-shared-server
  decisions - [0023](../../decisions/0023-postgres-replication-topology.md)/`docs/roadmap.md`'s
  still-open "per-service databases" direction. this pass assumes logical per-service databases
  on one shared instance as the minimum needed to design the migrate primitive, without settling
  that broader direction
- region/resource values overlays (`values.us-west-2.yaml` etc.) - 0009's shape stays available
  for genuinely per-region tuning, but no second region exists yet to need one
- Harbor/registry integration. local containerd image reuse is an explicit, called-out stand-in
- automated sync policy / self-heal. manual sync only, to keep each step of the prototype
  observable
- running Tilt and ArgoCD against `pings` at the same time. not the real target shape (ArgoCD
  runs standalone in non-local environments), so not exercised here - Tilt's `pings` resource is
  stopped before a sync, full stop
- an `argocd` CLI pin in `mise.toml`. `kubectl apply` + the port-forwarded UI are enough for
  this pass
- building out 0017's `app.pod.container`/`_pod-service.tpl`/`_pod-job.tpl` split. migrate's own
  pod spec is self-contained per 0025, not a dependency of this pass

## Design Decisions Summary

| decision | options considered | chosen | rationale |
|---|---|---|---|
| git source | `main` directly; a dedicated feature branch | feature branch | avoids forcing every prototype iteration through `main`. whether it tracks `main` for real is decided once the pattern's proven |
| namespace strategy | same namespace as Tilt (`pings`); a separate namespace (`pings-gitops`) | same (`pings`) | ArgoCD runs standalone against `pings` in every non-local environment - no Tilt there to collide with. proving a synthetic separate-namespace shape locally would validate a topology that never exists for real; the actual safety property is never running Tilt and ArgoCD against `pings` at once, not namespace isolation |
| image source | wait for Harbor; reuse the local containerd image | reuse local image, `IfNotPresent` | Harbor (initiative 5) is gated behind initiative 4 and isn't a blocker for proving local sync mechanics |
| migrate hook's postgres | a dedicated second instance; reuse Tilt's `local_database()` | reuse Tilt's | idempotent, and a second instance adds no coverage of the interesting mechanic (hook execution ordering) |
| manifest location | `deploy/argocd/` (real shape); `local/argocd/` (prototype) | `local/argocd/` for the generator/registry; `deploy/helm/app` for migrate itself | matches 0008's split for the ArgoCD-side plumbing; migrate is a real chart primitive per 0025, not prototype-only, so it belongs with the rest of the chart |
| service scope | `pings` + `iam-jwks`; `pings` only | `pings` only | `pings` alone exercises the new mechanics that matter; `iam-jwks` adds no new proof for this pass |
| topology mechanism | one `AppProject`+`ApplicationSet` per service (hand-written); a services registry driving a single `matrix`-generator `ApplicationSet` | registry + `matrix` generator | one hand-written pair per service is `O(services)` ArgoCD YAML - a real liability at ~60 services. a registry + generator makes onboarding a service a data change, not a manifest-authoring exercise |
| `AppProject` creation | hand-written per service; generated from the registry via N bootstrap Applications (one per service); generated via one Application ranging over the whole registry | one Application, ranges over the registry | same end result (one `AppProject` per service) as the N-Application option, without the extra bootstrap-ordering problem of each of those Applications needing a project of its own first |
| `AppProject` granularity | one global project; one per team/domain; one per service | one per service | project identity should be the stable axis - team ownership changes far more often than a service does. team/domain stay as labels on the (stable) per-service project, ready to drive a generated RBAC-policy mapping later without ever restructuring projects |
| region/environment source | 0009's committed overlay files (`values.<region>.yaml`); `helm.parameters` from the `clusters` generator's own labels | generator-derived (`helm.parameters`) | narrows 0009 per [0026](../../decisions/0026-region-environment-argo-derived.md) - which cluster an Application lands on is already the generator's own fact; writing it into a second, selected-by-matching file is one truth expressed twice, and today's `pings` values already prove it can drift (`HYRULE_REGION: us-east-2` vs the cluster label's `us-west-2`) |
| migrate location | stays under `local/argocd/` as a second Application source (as drafted first); becomes an opt-in `deploy/helm/app` template | chart-native, opt-in (`migrate.enabled`) | narrows 0017 per [0025](../../decisions/0025-migrate-as-chart-primitive.md) - any of ~60 services needs to be able to opt into its own migrate step without a bespoke ArgoCD-side file per service; also unifies Tilt's local path onto the same definition instead of a hand-maintained duplicate |
| database topology (informing migrate only) | migrate keeps targeting the single shared `hyrule` database; a per-service logical database on the one shared postgres instance | per-service logical database, one shared instance for now | unblocks a chart-native migrate design (each service's Job needs its own target) without deciding 0023's broader physical-topology/replication question, which stays open |
| migrate opt-in shape | a standalone `migrate.enabled` boolean; a `datastores` list (`slug`, `flavor`, flavor-specific keys) a service declares to opt in | `datastores` list | a service engineer shouldn't need to understand migrate's mechanics to opt in - declaring the datastore itself is the opt-in. also generalizes past postgres (`flavor: key-value`, etc.) without a rewrite once a second flavor actually needs schema/config of its own |
| postgres connection facts | a full DSN typed into each service's own values (as first drafted); a single chart-level default in `deploy/helm/app/values.yaml` | chart-level default | a platform engineer should make one edit, in one file, and have it apply to every opted-in service implicitly - not hand-edit N services' values or generate per-service manifests. same mechanism 0026 already uses for region/environment; a service supplies no DSN, ever |
| migrate ConfigMap naming | a per-service value (`migrationsConfigMap`, typed by the service engineer, as first drafted); chart-derived from `app.fullname` | chart-derived | one less thing a service engineer types or can get wrong; every scaffolded service already sets `fullnameOverride: <slug>` (0006), so the derived name matches today's hand-typed one exactly |

## Requirements

- **R1:** ArgoCD is installed into the local cluster via its own Helm chart, brought up via
  `make argocd-up` and torn down via `make argocd-down`.
- **R2:** `local/argocd/appprojects/services.yaml` exists as the single registry (service, team, domain) that
  drives everything below - onboarding a service is a registry line plus its existing
  `deploy/values/<service>/`, never new hand-written ArgoCD YAML.
- **R3:** one `AppProject` per registered service is generated (never hand-written) from the
  registry, scoped to that service's own namespace.
- **R4:** a single `ApplicationSet` uses a `matrix` generator (the registry × a `clusters`
  generator matched on region labels) to template one `Application` per service × matched
  cluster, sourced from a dedicated feature branch of this repo, targeting
  `deploy/helm/app-platform` + `deploy/values/<service>/values.yaml`.
- **R5:** `HYRULE_REGION`/`HYRULE_ENVIRONMENT` are injected into each generated `Application` from
  its matched cluster's own labels, never hardcoded in a service's committed values file.
- **R6:** the migrate Job is an opt-in `deploy/helm/app` template, opted into by declaring a
  `datastores` entry (`flavor: postgres`) rather than a standalone flag, carries the PreSync hook
  annotations (per 0012's designed shape), demonstrably runs before the rest of a fresh
  Application sync, and is the same definition Tilt's local dev loop now renders -
  `local/cluster/migrate-job.yaml` is retired. a service supplies no connection string and no
  ConfigMap name - both derive from a chart-level default and chart-native naming.
- **R7:** a values change pushed to that branch and manually synced in ArgoCD is reflected in
  the running `pings` deployment.
- **R8:** Tilt's `pings` resource is stopped, then ArgoCD's first sync cleanly takes ownership of
  the existing `pings` namespace - no orphaned or conflicting resources left over from Tilt's
  prior management of it.
- **R9:** `docs/roadmap.md` has a new numbered initiative entry for this work.
