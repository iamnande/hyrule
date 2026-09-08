# 0026: region/environment identity is Argo-derived, not values-file-layered - narrows 0009

[0009](0009-environments-and-promotion.md)'s "values shape" section
decided region-specific config lands in a committed overlay file per
region (`values.us-west-2.yaml`, `values.us-east-2.yaml`, delta-only,
layered on top of the base file via Helm's `-f`), and explicitly left "the
ArgoCD Application/AppProject shape that consumes this layout" undecided,
deferred to [0002](../proposals/0002-argocd-gitops-foundation/prd.md).

this decides that question, for `HYRULE_REGION`/`HYRULE_ENVIRONMENT`
specifically: they're injected as `helm.parameters` on the ApplicationSet
template that generates each service's Application, sourced from the
`clusters` generator's own matched-cluster labels
(`{{metadata.labels.region}}`, `{{metadata.labels.environment}}`) - never
committed to any service's values file, base or overlay. a service simply
never sets these two keys; Argo always supplies them.

## why not the overlay-file shape for these two

which cluster an Application lands on is exactly what the `clusters`
generator already knows - that's its whole job. writing the same fact
into a second place (a committed overlay file, selected by matching the
generator's own selector) is one source of truth expressed twice, and the
two can drift: today's `deploy/values/pings/values.yaml` already hardcodes
`HYRULE_REGION: us-east-2` while `local/argocd/cluster-local.yaml`'s label
says `us-west-2` - proof this isn't hypothetical. deriving both from the
generator removes the second copy entirely.

`local/argocd/cluster-local.yaml` gains an `environment: local` label
alongside `region:` - real CP/DP cluster registrations (not built yet,
per [0002](../proposals/0002-argocd-gitops-foundation/prd.md)'s
non-goals) carry `environment: prod` the same way.

Tilt's own local dev loop doesn't go through ArgoCD's generator at all, so
`deploy/helm/app`'s chart defaults carry `HYRULE_ENVIRONMENT: local` (no
default `HYRULE_REGION` - a single local cluster has no second region to
distinguish). ArgoCD's `helm.parameters` override both for any
Argo-managed Application, chart defaults are what Tilt actually runs with.

## overlay files aren't retired

0009's `values.<region>.yaml` shape stays exactly as decided for anything
that's genuinely per-region tuning the `clusters` generator can't supply
on its own - replica counts, resource sizing, and 0009's own examples.
narrowed to exclude region/environment identity specifically, not
replaced wholesale. no overlay file exists yet regardless - no second
region exists yet to need one.

## not decided here

- how replica/resource overlays get consumed once a second region exists
  and one is actually written - `helm.valueFiles` on the same
  ApplicationSet template can carry it (`[base, overlay]`, last wins,
  same as 0009's own `-f` layering), sketched here as the likely shape,
  not committed to until a real overlay file exists to prove it against.
