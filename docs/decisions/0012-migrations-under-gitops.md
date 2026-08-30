# 0012: migrations under GitOps - a PreSync hook, run independently per region

the migrate `Job` is immutable by design once created, and
[conventions.md#local-orchestration](../conventions.md#local-orchestration)
is explicit that re-running it after drift is a manual `tilt trigger
migrate` click - "verified directly, this doesn't happen automatically."
that's fine with a human driving Tilt. it has no answer under ArgoCD: no
Tilt UI, no human to click the button, and ArgoCD's default behavior
(diff the rendered manifest, apply whatever changed) doesn't naturally
re-run a Job just because `migrations/` changed.

**an Argo sync-hook Job** - annotated `argocd.argoproj.io/hook: PreSync`
with `argocd.argoproj.io/hook-delete-policy: BeforeHookCreation` - is the
shape: ArgoCD creates it and waits before syncing the rest of the
manifests, and the delete policy means each sync gets a fresh Job instead
of erroring on the old immutable one, the same problem
[conventions.md](../conventions.md#local-orchestration) already flags for
the Tilt path today.

## runs against each region's own postgres, independently

[0023](0023-postgres-replication-topology.md) settles the question this
doc originally left open: CP and DP are two independently-migrated
databases, not one shared instance, because DP's logical-replication
subscriptions need their own matching schema (DDL doesn't replicate).
so the migrate Job isn't a single sync-wide step - it's one PreSync hook
per region's Application, each applying the identical
[migrations](../../migrations) content against that region's own
postgres. [0009](0009-environments-and-promotion.md)'s lockstep
promotion is what keeps this safe: both regions sync together, so both
migrate Jobs run before either region's new application code (which
might depend on the new schema) starts serving traffic.

## not decided here

- the actual ArgoCD Application/AppProject manifests - no ArgoCD instance
  exists yet to write them against.
- DDL-vs-replication-lag ordering between the two regions' migrate
  Jobs - [0023](0023-postgres-replication-topology.md) already flags this
  as a known sharp edge, not solved preemptively.
