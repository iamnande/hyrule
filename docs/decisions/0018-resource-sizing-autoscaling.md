# 0018: resource sizing & autoscaling - static, conservative requests/limits per service; no HPA until real contention data exists

## already in practice, just not written down until now

both real services' values files already carry the identical block -
`resources: requests: {cpu: 50m, memory: 64Mi} limits: {memory: 128Mi}`
in [deploy/values/iam-jwks/values.yaml](../../deploy/values/iam-jwks/values.yaml)
and [deploy/values/pings/values.yaml](../../deploy/values/pings/values.yaml) -
because `local/new-service.sh` already stamps it into every new service's
`deploy/values/<slug>/values.yaml` at scaffold time. that wasn't a
decision anyone wrote down, just a starter value that happened to match
[0006](0006-service-scaffold.md)'s scaffold-is-the-only-creation-path
rule. this doc makes it deliberate instead of accidental, same move
[0016](0016-observability-backend.md) made for "sentry spans around
domain-level operations" already being true in code before it was ever
written into `conventions.md`.

## memory request + limit, CPU request only - no CPU limit

a CPU limit means CFS throttling the instant a service bursts past it,
which degrades a request in progress; a memory limit means OOM eviction,
which protects the node from one runaway pod starving everything else on
it. on homelab-scale hardware where the tradeoff is "let a burst borrow
idle CPU" vs "guarantee a runaway pod can't take the node down with it,"
the second matters and the first doesn't - so memory gets both a request
(scheduling signal) and a limit (a hard ceiling), CPU gets a request
only. this is already exactly what the scaffold's numbers encode
(`limits:` only ever lists `memory`); the QoS class that falls out is
`Burstable`, not `Guaranteed` - correct default, no service has asked for
guaranteed CPU yet.

**correction: `deploy/helm/app/values.yaml`'s own default carries these
numbers too, not `resources: {}`.** the first pass of this doc left the
chart's default empty on the theory that "`app` has no opinion on any
service's specific numbers" - but that reasoning proves too much: taken
seriously, it argues against `probes`' chart-level defaults
(`enabled: true`, real paths) existing either, and those are already
there. `app` isn't actually opinion-free today, and a safety floor that
protects the node isn't the kind of per-service opinion
[0005](0005-helm-chart-split.md#what's-in-app-vs-platform) was talking
about anyway - it's closer to "don't ship unbounded by default," the
same category probes already fall into. [0005](0005-helm-chart-split.md)
also advertises `app` as portable enough to be "lifted into any k8s
project as-is" - a consumer that isn't this repo's own scaffold gets
zero protection from an empty default, which undercuts this doc's whole
memory-limit rationale above. the chart's default is now the same
`cpu: 50m` / `memory: 64Mi` request, `memory: 128Mi` limit; a service
overrides it in `deploy/values/<slug>/values.yaml` exactly like before.
the scaffold's own copy of this block
([local/new-service.sh](../../local/new-service.sh)) and the two live
services' values files are now redundant with the chart default rather
than the only source of it - harmless as-is, and a candidate for a later
cleanup pass, not addressed further here.

## the scaffold's numbers are a starting point, not a hard rule

`50m`/`64Mi`/`128Mi` came from nowhere more rigorous than "a small Go
service probably needs about this much" - fine as a scaffold default,
not something to treat as measured. a service that demonstrably needs
more overrides it in its own `deploy/values/<slug>/values.yaml`, same
mechanism already in use; there's no policy requiring every service stay
at the default forever. [0016](0016-observability-backend.md)'s
Prometheus install is what eventually turns "probably needs about this
much" into real data - revisit specific services' numbers once it exists,
not before.

## no HPA - static sizing is the actual policy at this scale, not a placeholder for one

[0003](0003-runtime.md)'s "homelab-scale hardware, not a fleet of cloud
nodes" is a real constraint, not just an inference: a horizontal
autoscaler earns its complexity when there's enough headroom across
enough nodes for scaling out to mean something, and enough real traffic
variance to justify reacting to it automatically. neither is true yet.
static requests/limits, bumped by hand when a service needs more, is
simpler and sufficient until there's contention data saying otherwise -
same "don't build the mechanism before a real need" reasoning as
[0007](0007-iam-jwks-key-distribution.md)'s deferred eventing,
[0015](0015-authn-authz.md)'s deferred policy engine, and
[0017](0017-scheduled-workloads.md)'s deferred job/cronjob templates. no
`hpa.yaml` gets added to `app` now; if usage data from
[0016](0016-observability-backend.md)'s Prometheus install ever shows
real contention, that's a fresh decision against concrete evidence, not
something to build speculatively here.

## not decided here

- specific numeric tuning for any one service beyond the shared
  default - per-service, decided empirically once
  [0016](0016-observability-backend.md)'s metrics exist to decide it
  against.
- a vertical pod autoscaler - no evidence yet the static default is
  wrong often enough to automate adjusting it.
- `Guaranteed` QoS (requests == limits) for any workload - no service
  has a concrete need for guaranteed CPU yet.
