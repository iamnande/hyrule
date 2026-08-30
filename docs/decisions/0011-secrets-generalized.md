# 0011: secrets, generalized - External Secrets Operator, 1Password Connect as backend

**supersedes the distribution mechanism in
[0007](0007-iam-jwks-key-distribution.md)** - 1Password stays the source
of truth (self-hosted Connect server, already decided there), but
`onepassword-operator` is out: too little production mileage to trust as
load-bearing cluster infrastructure, versus the alternative below, which
is a widely-deployed CNCF project most people touching Kubernetes secrets
have actually run. this also turns 0007's single-secret answer (the
iam-jwks signing key) into the general one for any future secret - a
database password, a third-party API key - not just that one.

## External Secrets Operator (ESO) + a self-hosted 1Password Connect server

ESO's own 1Password Connect provider talks to the same self-hosted
Connect server 0007 already called for - nothing about the backend or
"1Password is the source of truth" changes. what changes is who turns a
vault item into a real Kubernetes `Secret`:

- a cluster-scoped `ClusterSecretStore` points ESO at the Connect server -
  installed once, cluster-singleton infrastructure, the same category as
  Harbor ([0010](0010-image-registry-and-publishing.md)) and the Connect
  server itself. not wrapped by `app`/`platform`'s per-service model.
- each service that needs a secret declares its own `ExternalSecret`
  (which vault item, which fields, which K8s `Secret` name to write) in
  its own namespace ([0014](0014-namespacing.md)) - this is the
  per-service piece, and it's exactly the shape 0007 already carved out
  for `OnePasswordItem`: a per-service custom resource plugging into an
  already-running cluster controller, landing via `platform`'s
  `additionalK8sObjects` escape hatch
  ([0005](0005-helm-chart-split.md#what's-in-app-vs-platform)). only the
  CRD name changes.
- the resulting `Secret` is consumed exactly the way `app`'s chart
  already supports today, no chart change needed: `extraEnvFrom` (see
  [deploy/helm/app/values.yaml](../../deploy/helm/app/values.yaml)) points
  a service's pod at it as env vars, matching
  [conventions.md#config](../conventions.md#config)'s "env vars, full
  stop" rule. **an env var, not a mounted file, is the default for every
  secret** - ESO writing a `Secret` doesn't require a volume mount, and
  nothing about this pipeline needs one.

**local dev runs the same pipeline, not a stand-in for it** - matches the
standing parity principle from
[0010](0010-image-registry-and-publishing.md): local and prod diverge
only where there's a concrete, environment-specific reason to, and there
isn't one for secrets delivery. a local 1Password Connect server + ESO
+ `ClusterSecretStore` exist in the local cluster too, sourced from the
same 1Password account (a local-only vault/item, not prod's).

## rotation still restarts the pod, without 1Password's operator doing it

what made the original operator attractive was "rotation propagation is
free - no staleness window to reason about beyond how long until it
notices." that property is worth keeping, just built from more
broadly-proven parts: pair ESO with a lightweight reloader (e.g.
Stakater's `Reloader`) that watches a `Secret` for changes and rolls the
Deployment consuming it. two independently well-adopted tools composed
together, instead of one single-vendor tool that bundles sync-and-restart
as one opinionated unit.

## `iam-jwks`'s `KeyStore` collapses to one implementation

0007's two `KeyStore` implementations (postgres-backed for local dev,
secret-file-backed for a real deployment) existed because local dev had
no reason to stand up real secrets infrastructure just to run `iam-jwks`.
that reason is gone now that 1Password Connect + ESO run locally too
(above) - keeping two implementations around after that would just be
two things to maintain and a chance to trust the wrong one.

collapsing to **one implementation, reading the signing key from an env
var** - not a mounted file. the "secret-file-backed" name and the
mounted-`Secret`-volume design in 0007 predates this decision and was
never actually required by anything: Ed25519 key material is small
enough to fit an env var trivially, `extraEnvFrom` already gets it there
(above), and reading it as a file was quietly inconsistent with
[0005](0005-helm-chart-split.md#config-env-vars-only-no-config-files)'s
"env vars only, no config files" rule the rest of this repo already
follows. the postgres-backed `KeyStore`, its repository package, and the
`iam_jwks_keys` migration all go away with it - `iam-jwks` stops using
postgres entirely. `hyrule_app_ro` itself stays: conventions.md is
explicit `iam-jwks` was only ever "the first to use it," not the only
intended consumer - it's a general role for the next read-only service,
not something scoped to `iam-jwks` specifically. this is real code
change, tracked as follow-up work alongside standing up the local
Connect server/ESO pipeline itself, not done in this doc.

## not decided here

- how ESO itself and the `ClusterSecretStore` get installed, in both the
  local cluster and CP - real infrastructure work, same deferral as
  Harbor in [0010](0010-image-registry-and-publishing.md).
- the Connect server's own token has to be seeded into the cluster
  before ESO can use it - a manual bootstrap step (`kubectl create
  secret` once, by hand, during cluster setup) that can't itself be
  managed by the system it's bootstrapping. standard chicken-and-egg for
  this kind of setup, not solved further here.
- ESO also supports Vault, cloud-provider secret managers, and others
  behind the same `SecretStore` abstraction - not a reason to hedge on
  1Password now, just worth knowing the escape hatch exists if that ever
  needs to change: only the `SecretStore` config would move, not every
  consuming service.
