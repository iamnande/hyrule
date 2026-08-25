# 0005: one generic chart + a platform chart + a wrapper, not one chart per service

adapted from an external convention doc the user provided (app/platform
split with dynamic settings injection) - not hyrule-specific in origin,
but the shape fits: homelab will eventually run many services, each
needing the same boilerplate (Deployment, probes, resources, RBAC) plus
whatever org-specific plumbing shows up later (secrets, internal CRDs).
one chart per service means drift; one giant chart means org plumbing
leaks into what should be portable.

## topology

`deploy/helm/app` (generic, zero org assumptions) +
`deploy/helm/platform` (org-internal plumbing) +
`deploy/helm/app-platform` (thin wrapper, aliases both as dependencies).
per-service values move to `deploy/values/<service>/values.yaml`, using
only the two wrapper keys (`app:`, `platform:`).

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
mechanism below. everything here should be safe to hand to an external
party with zero hyrule-specific assumptions leaking in.

`platform`: intentionally near-empty right now.
[docs/conventions.md#config](../conventions.md#config) already says
secrets are injected at invocation time, not via a k8s-native secret
backend, and there are no internal CRDs yet - so this chart is just an
`additionalK8sObjects` escape hatch until a real integration exists.
when one does, it goes here, never in `app`.

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

simplified from the source doc: skipped the per-leaf `tpl` re-templating
step (source doc's version re-runs specific map values through Helm's
template engine before serializing) since it assumes a `services`
sub-map schema hyrule doesn't have. add it back the same way if a real
service's settings need render-time templating.

## deliberately deferred

- `values.schema.json` + generated `README.md` (readme-generator-for-helm
  convention) - `values.yaml` is annotated ready for it, generation
  tooling isn't wired into CI yet.
- `helm-unittest` snapshot tests - chart correctness is exercised today
  by `tilt ci` + the integration suite against a live cluster; add
  per-template suites once there's a second workload mode or service to
  actually catch drift between.

both tracked in [conventions.md#known-gaps](../conventions.md#known-gaps).
