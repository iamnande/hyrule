# 0025: migrate becomes a real `app` chart primitive - narrows 0017

[0017](0017-scheduled-workloads.md) has a section, "not the migrate Job's
pattern," ruling migrate out of `app` entirely: "one-off infra bootstrap
tooling that lives outside both `app` and `platform`... don't conflate the
two when this finally gets built." that held when migrate was a single
hand-written Job for one shared local postgres. it doesn't hold once
[0002](../proposals/0002-argocd-gitops-foundation/prd.md) needs any of
~60 services to optionally run its own migrate step against its own
logical database - that needs a chart-native, opt-in mechanism, not a
bespoke ArgoCD-side file per service.

this narrows that one section only. everything else 0017 decided - the
`job`/`cronjob` primitive split, `_pod-container.tpl`/`_pod-service.tpl`/
`_pod-job.tpl`, `mode: job`/`cronjob` staying unbuilt until a real service
needs one - is unaffected.

## `migrate` is its own primitive, not `job`/`cronjob` mode

`app` gains a `migrate.yaml` template, gated by `migrate.enabled` (default
`false`), independent of `mode:` - a service stays `mode: deployment` and
separately opts a migrate Job in alongside it. reasons it's not `job`/
`cronjob` mode:

- 0017's `job`/`cronjob` mode is for a *service* owning a recurring or
  one-off unit of its own business work, triggered by `schedule` or a
  one-off apply. migrate is schema bootstrap, triggered by ArgoCD's
  PreSync hook lifecycle (or Tilt's `resource_deps`) - never a `schedule`.
  keeping it a separate primitive keeps `job`/`cronjob` mode's meaning
  exactly what 0017 intended.
- the existing `app.pod`/`_pod.tpl` partial hardcodes an `http` port and
  all three probes - wrong for a job, per 0017's own diagnosis. rather
  than wait on 0017's `_pod-container.tpl` split (still unbuilt, per
  `docs/roadmap.md#devx`) as a prerequisite, `migrate.yaml` carries its
  own small, self-contained pod spec. not wired into `app.pod.container`
  when that split eventually lands - reconsider then, not guessed at now.

a service opts in by declaring a `datastores` entry
(`flavor: postgres`), not a standalone `migrate.enabled` flag - see
[0002](../proposals/0002-argocd-gitops-foundation/tech-spec.md)'s
`datastores` shape for the full values design and why. `migrate.yaml`
ranges over `.Values.datastores`, emitting one Job per entry with
`flavor: postgres` and `schema_management` not `disabled` (the default).
the Job's `-database=` DSN is built entirely from chart-level connection
defaults (`.Values.postgres`, a single file - see 0002) crossed with the
entry's own `slug` as the database name - a service never types a
connection string. the migrations `ConfigMap` name is chart-derived
(`{{ include "app.fullname" . }}-migrations`), also never a value. what's
still a chart default, not per-datastore: `migrate.image.repository`/
`image.tag`, `migrate.backoffLimit` - job mechanics, not connection
identity. carries `argocd.argoproj.io/hook: PreSync` +
`hook-delete-policy: BeforeHookCreation` unconditionally - inert metadata
under plain `kubectl apply`, meaningful only under ArgoCD.

a service with more than one `flavor: postgres` datastore entry isn't
handled - today's naming (`{{ app.fullname }}-migrate`/`-migrations`,
no `slug` in either) assumes exactly one per service, matching every
real case so far. revisit by weaving `slug` into both names if a second
one ever actually shows up, not guessed at now.

## Tilt unifies onto it too

`local/cluster/migrate-job.yaml` is retired. Tilt's `local_migrate()`
renders `app`'s own `migrate.yaml` (via `helm template`, `migrate.enabled:
true` set in local values) instead of a hand-maintained duplicate. Tilt is
orchestration - what triggers a sync, what depends on what - not a second
place that defines what a migrate Job looks like. the chart is the one
definition, consumed by both control planes.

## not decided here

- whether `migrationsConfigMap` population ever gets automated (a
  Kustomize generator, `configmap_from_dir` equivalent) - still the same
  accepted limitation [0002](../proposals/0002-argocd-gitops-foundation/tech-spec.md)
  already carries forward, now scoped per-service instead of `databases`
  only.
- `app.pod.container`/the rest of 0017's split - still unbuilt, still
  tracked in `docs/roadmap.md#devx`, not a prerequisite of this decision.
