# 0005: one generic chart + a homelab-plumbing chart + a wrapper, not one chart per service

homelab will eventually run many services on top of hyrule's pattern,
each needing the same boilerplate (Deployment, probes, resources, RBAC)
plus whatever homelab-specific plumbing shows up later (a secrets
backend, self-hosted CRDs like a VPN operator or a cert-manager issuer).
one chart per service means drift between services that all want the
same shape; one giant chart means homelab-specific plumbing leaks into
what should be a portable, reusable app shape. splitting the two and
gluing them with a thin wrapper avoids both.

## topology

`deploy/helm/app` (generic - no homelab-specific assumptions, could be
lifted into any k8s project as-is) + `deploy/helm/platform` (homelab
integration plumbing) + `deploy/helm/app-platform` (thin wrapper,
aliases both as dependencies). per-service values move to
`deploy/values/<service>/values.yaml`, using only the two wrapper keys
(`app:`, `platform:`).

`app-platform`'s dependencies are vendored, not resolved from a chart
repo at deploy time - `make helm-vendor` runs `helm dependency build`;
the resulting `Chart.lock` and `charts/*.tgz` are committed. re-run it
after editing `app` or `platform`, same as any other generated-and-committed
artifact in this repo.

## what's in `app` vs `platform`

`app`: workload (`mode: deployment` today - `job`/`statefulset`/etc. are
placeholders in the values shape, not templates; add one when a real
service needs it, following the shared `app.pod` partial), Service,
ServiceAccount, RBAC, probes, resources, env. nothing in here should
ever need to know it's running in homelab specifically.

`platform`: intentionally near-empty right now.
[docs/conventions.md#config](../conventions.md#config) already says
secrets are injected at invocation time, not via a k8s-native secret
backend, and homelab has no self-hosted CRDs to wrap yet - so this
chart is just an `additionalK8sObjects` escape hatch until one shows
up. when it does, it goes here, never in `app`.

## config: env vars only, no config files

`app` has no mechanism for mounting a config file into a pod - no
ConfigMap-backed settings blob, no `--config` flag, nothing. every
service's config comes in through `app`'s plain `env` map (literal env
var names, no prefix magic - `app` has no opinion on any service's
naming convention) plus `extraEnv`/`extraEnvFrom` for anything raw.
pings already works this way (`internal/lib/config`, `HYRULE_*`); this
makes it the only way, not just the default.

an earlier version of this chart carried a dynamic settings mechanism
(merge a `runtimeSettings` map, serialize it to JSON, mount it, pass
`--config`) for services whose config might not fit env vars. dropped
before any service used it - config files invite drift between what's
in the mounted file and what's actually running, in a way a rollout
triggered by an env var change doesn't.

## deliberately deferred

- `values.schema.json` + generated `README.md` (readme-generator-for-helm
  convention) - `values.yaml` is annotated ready for it, generation
  tooling isn't wired into CI yet.
- `helm-unittest` snapshot tests - chart correctness is exercised today
  by `tilt ci` + the integration suite against a live cluster; add
  per-template suites once there's a second workload mode or service to
  actually catch drift between.

both tracked in [conventions.md#known-gaps](../conventions.md#known-gaps).
