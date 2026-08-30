# 0009: environments & promotion model - local + prod (cp/dp regions), no staging, lockstep

three environments, not more: `local` (laptop/PC, the existing Rancher
Desktop/Tilt loop - unchanged) and `prod`, which is two regions rather
than one cluster - a control-plane region (`us-west-2`, closest real
region code to Oregon) and a first data-plane region (`us-east-2`), a VPS
to start for both, homelab hardware later. `Region` and `Environment` are
already separate axes in
[go/internal/lib/config](../../go/internal/lib/config) (`deployment.go`);
this decision is what actually populates them - see below.

CP/DP here is a deployment-topology fact (which physical region a
release runs in), not a code fork - [0006](0006-service-scaffold.md)
already ruled out control-plane/data-plane as a generator concept, and
nothing here reopens that. `pings` and `iam-jwks` don't need to know
which region they're running in beyond the `Region` value they already
carry for logging/discovery.

## no staging tier

local dev is the only gate before prod - CI (`lint`/`unit`/`integration`/
`smoke`, [conventions.md#ci](../conventions.md#ci)) plus whatever
[0012](0012-migrations-under-gitops.md) and initiative 1 on
[docs/roadmap.md](../roadmap.md) add for the Helm/Tilt path is what has
to go green. matches the repo's existing pattern of not building a
mechanism ahead of a real need for it ([0007](0007-iam-jwks-key-distribution.md)'s
eventing, [0018](0018-resource-sizing-autoscaling.md)'s autoscaling) -
there's no team or customer traffic yet to protect with a pre-prod gate.
revisit when there is one.

## cp and dp promote in lockstep

one promotion event (merge to `main`, green CI) updates both regions to
the same version together. the alternative - independent per-region
promotion - is more flexible (canary a change in one region first) but
means the two regions can silently run different versions of the same
service by design, which is the wrong default for something meant to
double as a company-launch foundation: "what version is running where"
should be one answer. sequencing *within* a lockstep promotion (parallel
vs. cp-then-dp) is an ArgoCD/rollout-strategy detail, decided when that
work happens, not here.

## values shape

`deploy/values/<service>/values.yaml` keeps meaning exactly what it means
today - the base/local defaults Tilt already reads, unchanged. each real
environment adds a values overlay, applied on top via Helm's own `-f`
layering (multiple `-f` flags, last one wins):

```
deploy/values/<service>/
  values.yaml              # base + local (Tilt reads this alone)
  values.us-west-2.yaml    # cp overlay - only what differs from base
  values.us-east-2.yaml    # dp overlay - only what differs from base
```

an overlay holds only the delta for that region (replica count,
resources, `HYRULE_REGION`, anything else region-specific) - not a full
copy of the base file. a service with no overlay file for a given region
simply doesn't run there yet (see "not decided here").

## code changes made alongside this decision

`go/internal/lib/config`'s `Region`/`Environment` enums predated this
decision and didn't match it once it was made:

- `Region` had `PrimaryRegion = "us-east-2"` / `SecondaryRegion =
  "eu-central-1"` - `eu-central-1` had no decision behind it anywhere in
  this repo and no consumer; dropped rather than kept as an unmotivated
  placeholder. renamed to `USWest2Region` / `USEast2Region` (location-based,
  not primary/secondary - a CP moving hardware later shouldn't require
  renaming a Go constant, and "secondary" doesn't scale once a second DP
  exists).
- `Environment` had `DevEnvironment = "dev"`, unreferenced anywhere except
  its own validation test - removed now that "no staging tier" is decided,
  rather than carried as unused surface.

## not decided here

- which services actually run in `us-west-2` only vs. also `us-east-2` -
  nothing forces this split yet with two small services; a service simply
  gets an overlay file for a region when it actually needs to run there.
  revisit when a real service has an actual reason to be region-specific.
- the ArgoCD Application/AppProject shape that consumes this layout, and
  promotion sequencing within a lockstep release - both depend on
  [0012](0012-migrations-under-gitops.md) and the rest of the pre-ArgoCD
  decision list in [docs/roadmap.md](../roadmap.md).
- how `local/new-service.sh` should scaffold the new overlay files (today
  it only writes one `values.yaml`) - a `make new-service` / scaffold
  change, not a decision, tracked as follow-up work once this shape is
  actually used.
