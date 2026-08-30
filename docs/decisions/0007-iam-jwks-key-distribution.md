# 0007: iam-jwks key distribution - 1Password + the platform chart, not a sync protocol

**superseded by [0011](0011-secrets-generalized.md)** in two ways:
1Password (via a self-hosted Connect server) stays the source of truth,
but the operator syncing it into a K8s `Secret` is External Secrets
Operator now, not `onepassword-operator`; and the two `KeyStore`
implementations below (postgres-backed, secret-file-backed) collapse
into one env-var-backed implementation, since local dev now runs the
same secrets pipeline as prod instead of needing a lighter-weight
stand-in. the `KeyStore` interface itself, and 1Password as the source of
truth, are unaffected.

the signing key's source of truth is a 1Password vault, not postgres and
not a custom store. registering the key as a vault item is the rotation
event - everything downstream reacts to that, nothing polls or pushes on
`iam-jwks`'s own initiative.

each region's cluster runs a 1Password Connect server and the
`onepassword-operator` - off the shelf, nothing built here. a
`OnePasswordItem` CRD points at the vault item and syncs it into a native
Kubernetes `Secret` in that cluster; the operator restarts `iam-jwks`'s
deployment when the item changes. rotation propagation is free - no
custom code, no staleness window to reason about beyond "how long until
the operator notices."

this is exactly what `deploy/helm/platform` exists for
([0005](0005-helm-chart-split.md), [0006](0006-service-scaffold.md)): a
secrets backend integration goes in `platform`, never `app`. the
`OnePasswordItem` lands there via `additionalK8sObjects`.

`iam-jwks` mounts the synced `Secret` as a volume, loads it once at
startup, serves from the in-memory cache. no live 1Password API calls at
request time, no cross-region network calls, no poll/push/eventing code
in the service at all.

## KeyStore: two implementations, not one

the domain defines a narrow `KeyStore` interface (see
[docs/style.md](../style.md) - consumer defines, producer returns
concrete types). two implementations satisfy it:

- **postgres-backed**, for local dev - no real key material involved,
  reuses the existing `hyrule-database` pattern every other
  database-backed service already uses. reads through a `hyrule_app_ro`
  role (strict `SELECT`, no write grants), so the read path can't write
  even in dev.
- **secret-file-backed**, for a real deployment - reads the mounted
  Secret volume at startup, nothing else.

which one gets wired in is a composition-root concern, not a runtime
branch inside the domain: `go/cmd/iam-jwks/app.Module(fileCfg)` picks
`svc.WithFileStore()` when `HYRULE_IAM_JWKS_KEYS_FILE_PATH` is set,
`svc.WithPostgres()` (plus `database.Module`) otherwise - the file path
being set is the only signal, no separate mode flag to keep in sync
with it. file-backed mode never includes `database.Module` at all, so
there's no postgres connection sitting unused in a real deployment.

## what this ruled out

poll, push, and custom eventing all solve "how does `iam-jwks` replicate
data across regions" - that turned out not to be `iam-jwks`'s problem.
1Password's own replication plus the operator's restart-on-change already
solves it. `iam-jwks` doesn't become the service that proves out
Kafka/Flink/eventing - see the roadmap update alongside this decision.

## not decided here

the 1Password Connect server + operator themselves aren't stood up yet -
that's real infrastructure work, tracked in
[docs/roadmap.md](../roadmap.md), sequenced behind there being a second
real region to test replication against.
