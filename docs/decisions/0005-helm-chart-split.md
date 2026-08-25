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
ServiceAccount, RBAC, probes, resources, env, and the dynamic settings
mechanism below. nothing in here should ever need to know it's running
in homelab specifically.

`platform`: intentionally near-empty right now.
[docs/conventions.md#config](../conventions.md#config) already says
secrets are injected at invocation time, not via a k8s-native secret
backend, and homelab has no self-hosted CRDs to wrap yet - so this
chart is just an `additionalK8sObjects` escape hatch until one shows
up. when it does, it goes here, never in `app`.

## dynamic settings injection

`app`'s `runtimeSettings` (plus `global.runtimeSettings` for
environment-wide defaults) gets merged and dumped as one JSON blob into a
ConfigMap, mounted at `/app/conf/<fullname>.yml`, wired in as `--config`
only when non-empty - and a `config-hash-<fullname>` pod annotation
forces a rollout on change. the chart never parses or validates the
blob; that's each service's own config-loading code's job.

pings doesn't use this - its config is pure env vars
(`internal/lib/config`, `HYRULE_*`), wired through `app`'s plain `env`
map instead (literal env var names, no prefix magic - `app` has no
opinion on any service's naming convention). the settings mechanism
exists for a future service whose config doesn't fit env vars, without
needing a chart change to onboard it.

skips re-running individual settings leaves back through Helm's `tpl`
function before serializing - that only pays off once a service's
settings need a value resolved at render time (a namespace, a release
name), and nothing does yet. add it the same way `runtimeSettings`
itself was added, when a real service needs it.

## deliberately deferred

- `values.schema.json` + generated `README.md` (readme-generator-for-helm
  convention) - `values.yaml` is annotated ready for it, generation
  tooling isn't wired into CI yet.
- `helm-unittest` snapshot tests - chart correctness is exercised today
  by `tilt ci` + the integration suite against a live cluster; add
  per-template suites once there's a second workload mode or service to
  actually catch drift between.

both tracked in [conventions.md#known-gaps](../conventions.md#known-gaps).
