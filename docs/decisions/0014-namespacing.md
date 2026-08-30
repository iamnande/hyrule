# 0014: namespacing - one namespace per service, no role or domain buckets

every `go/internal/svc/<name>` service gets its own namespace, named after
its slug (`pings`, `iam-jwks`) - the same name as its Helm release
([Tiltfile](../../Tiltfile)'s `deploy_service`, `release_name=slug`) and
its `deploy/values/<slug>/` directory. no coarser grouping sits above
that: no `cp`/`dp` namespace pairing the regions from
[0009](0009-environments-and-promotion.md) reintroduce as a namespace
concept, and no shared `data` namespace either - namespace boundaries
track services, not roles.

## domain grouping: label it, don't merge namespaces for it

the temptation once a second `iam-*` service exists is to put every
service under one domain into one namespace (`iam`). not doing that: it's
the same role-bucket shape as `cp`/`dp`/`data` wearing a different name,
and it trades away exactly what per-service namespaces buy - RBAC
boundaries, resource quotas, and blast radius contained to one service.

domain relationship is a real thing worth being able to see and query,
just not a namespace boundary. it goes on as a label instead (e.g.
`hyrule.io/domain: iam`) applied to each service's own namespace - group
by it in `kubectl`, a dashboard, or a `NetworkPolicy`/RBAC rule scoped to
the label, without collapsing the isolation. the exact label key and
whether anything actually consumes it yet isn't decided here - there's
one `iam-*` service today, nothing to group. revisit when a second one
exists and the grouping need is real, not hypothetical - same pattern as
[0006](0006-service-scaffold.md)'s control-plane/data-plane call and
[0007](0007-iam-jwks-key-distribution.md)'s eventing deferral.

## shared, non-service infrastructure gets its own namespace too

the shared postgres and the `migrate` Job aren't a `go/internal/svc/<name>`
service - conventions.md's data layer section is explicit that every
database-backed service shares this one instance. forcing it into any one
service's namespace would make that service look like it owns the
database; it doesn't. it gets a namespace of its own: `databases` - a
category namespace for shared stateful backends generally (postgres
today, whatever comes next - redis, TiKV, anything else in that shape),
not a per-technology namespace and not a generic `data` bucket either,
which is exactly the role-bucket shape rejected above. the same goes for
whatever lands in `deploy/helm/platform` later (the 1Password Connect
server/operator from [0007](0007-iam-jwks-key-distribution.md)) - cluster-
scoped plumbing that isn't a service gets its own namespace, typically
whatever a given third-party chart already defaults to, decided per
resource as it's introduced rather than pre-declared now.

**nothing here carries the repo/project name.** the namespace is
`databases`, the postgres `Deployment` is `postgres`, its Services are
`postgres-rw`/`postgres-ro` - not `hyrule-database`. a rename or a
company kickoff shouldn't mean chasing a project name through every DNS
reference a dozen services depend on; infra identifiers describe what a
thing *is* (a namespace of databases, a postgres instance), never what
this repo happens to be called this year. `postgres-rw`/`postgres-ro`
both point at the exact same single instance today - there's no read
replica yet, this is a naming/routing seam only, so that when a real
replica shows up later, only `postgres-ro`'s selector changes and every
consuming service (already pointed at the right endpoint for its role)
needs zero changes. `iam-jwks` (the `hyrule_app_ro` role) uses
`postgres-ro`; `pings` (the `hyrule_app` role) uses `postgres-rw`. the
Postgres-internal names (`hyrule_owner`, `hyrule_app`, `hyrule_app_ro`,
the `hyrule` database name itself) are a separate, bigger surface -
untouched here, out of scope for this decision.

## what changed to wire this in

- `local/cluster/postgres.yaml`: `Deployment` renamed `hyrule-database` ->
  `postgres`, in the new `databases` namespace; one `Service` split into
  two (`postgres-rw`, `postgres-ro`), both selecting the same pod.
  `migrate-job.yaml` moved to the `databases` namespace too and now
  connects to `postgres-rw` by its short name (works because both live in
  the same namespace).
- [Tiltfile](../../Tiltfile): `local_database()` applies a `databases`
  `Namespace` object before anything else lands in it; `configmap_from_dir`
  takes an explicit `namespace` argument so both generated ConfigMaps
  (`postgres-init`, `hyrule-migrations`) land in it too. the Tilt resource
  for the database workload is now named `postgres`, matching the renamed
  `Deployment`. `deploy_service()` passes `namespace=slug` and
  `--create-namespace` to `helm_resource`, so every service's own
  namespace is created on first deploy without a manual step.
- every service now reaches postgres via its role-appropriate
  cross-namespace DNS name (`postgres-rw.databases.svc.cluster.local` or
  `postgres-ro.databases.svc.cluster.local`), not a bare short name that
  only resolves within the same namespace -
  `deploy/values/{pings,iam-jwks}/values.yaml` and the
  `local/new-service.sh` scaffold template (defaults to `postgres-rw`, the
  `hyrule_app` role a scaffolded service starts with) all updated.
  [conventions.md#local-orchestration](../conventions.md#local-orchestration)
  updated to match.
- `local/new-service.sh` and both existing services' values files were
  also still writing `HYRULE_ENVIRONMENT: dev` - stale from before
  [0009](0009-environments-and-promotion.md) removed `dev` as a valid
  `Environment` value. fixed to `local`, since this is the value Tilt's
  local cluster deploy actually uses.
- deliberately left alone: `mk/database.mk`'s standalone container (used
  by `make db-up`/CI) is still named `hyrule-database` - it's a plain
  docker/podman container name nothing else resolves by DNS or depends
  on, so renaming it wouldn't reduce any real rename pain, just add diff
  noise. the Postgres-internal role/database names (`hyrule_owner`,
  `hyrule_app`, `hyrule_app_ro`, `hyrule` itself) are the same kind of
  bigger, separate surface - not touched here.

## not decided here

- the actual domain label key/taxonomy, and whether anything consumes it -
  deferred until a second `iam-*` (or any other domain-sharing) service
  exists.
- whether a `NetworkPolicy` should explicitly allow the
  service-namespace -> `databases`-namespace path, or whether k3s's
  default CNI leaves that open by default either way - a security
  hardening question, not a namespacing one.
- this hasn't been verified against a live `tilt up`/`make dev` run yet -
  do that before trusting it fully.
