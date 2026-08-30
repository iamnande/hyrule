# 0017: recurring/scheduled workloads - `app` gains real `job`/`cronjob` templates when first needed, sharing `app.pod` but not its deployment-shaped assumptions

still not built - [0005](0005-helm-chart-split.md#topology) is explicit
that `mode: job`/`cronjob` stay values-shape placeholders until a real
service needs one, and nothing does yet. what this decides is the shape,
so whoever adds the first one isn't guessing.

## `job` is the primitive, `cronjob` wraps it with a schedule

`deploy/helm/app/templates/job.yaml` and `cronjob.yaml`'s `jobTemplate`
share an identical pod shape - nothing about the pod itself differs
between a one-off `job` and a recurring `cronjob`, only what triggers it.
`cronjob.yaml` is that same shape plus `schedule` and
`concurrencyPolicy`.

values gained under a `job:`/`cronjob:` block: `schedule`,
`concurrencyPolicy` (cronjob only), `backoffLimit`, `restartPolicy`
(`Never` or `OnFailure` - never `Always`, that's a Deployment-only
value), `activeDeadlineSeconds`, `ttlSecondsAfterFinished`,
`successfulJobsHistoryLimit`/`failedJobsHistoryLimit` (cronjob only).

## purpose-built pod partials per workload shape, not one shared partial with mode conditionals

checked the existing `app.pod` partial
([_pod.tpl](../../deploy/helm/app/templates/_pod.tpl)) directly: it
hardcodes an `http` container port and all three probes
(`startupProbe`/`livenessProbe`/`readinessProbe`) unconditionally keyed
off `.Values.service.port` - correct for a Deployment serving HTTP,
meaningless for a batch job that runs to completion and exits. rather
than retrofit that partial with `if .Values.mode` branches sprinkled
through it, split by workload archetype instead - each partial only ever
means one thing, no conditionals to trace through later:

- **`_pod-container.tpl`** (`app.pod.container`) - the slice that
  genuinely doesn't differ by workload shape: image, env/envFrom,
  resources, volumeMounts, additionalContainers. pulled out once so the
  two pod partials below don't duplicate it.
- **`_pod-service.tpl`** (`app.pod.service`) - full pod spec for
  `deployment` mode: labels, serviceAccountName, imagePullSecrets,
  initContainers, volumes, and the container via
  `include "app.pod.container"` plus the `http` port and all three
  probes, unconditionally.
- **`_pod-job.tpl`** (`app.pod.job`) - same skeleton, no port, no
  probes. used by both `job.yaml` and `cronjob.yaml`'s `jobTemplate`.

`deployment.yaml` calls `app.pod.service`; `job.yaml`/`cronjob.yaml` call
`app.pod.job`. `_pod.tpl`/`app.pod` goes away, replaced by these three.

## not the migrate Job's pattern

the existing `migrate` Job
([local/cluster/migrate-job.yaml](../../local/cluster/migrate-job.yaml),
and the ArgoCD `PreSync` hook [0012](0012-migrations-under-gitops.md)
describes for CP/DP) looks like prior art but isn't - it's one-off infra
bootstrap tooling that lives outside both `app` and `platform` entirely,
triggered by Tilt locally and by ArgoCD's sync lifecycle in a real
environment, not by a `schedule`. `app`'s `job`/`cronjob` mode is for a
*service* owning a recurring or one-off unit of its own work (a cleanup
task, a scheduled report) - a different thing that happens to compile to
the same Kubernetes primitive. don't conflate the two when this finally
gets built.

## observability: no new `platform` plumbing needed

[0016](0016-observability-backend.md)'s `ServiceMonitor`-in-`platform`
pattern doesn't apply here - a job pod that runs for seconds and exits is
too short-lived for Prometheus to scrape meaningfully. job/cronjob
success, failure, and duration are already exposed cluster-wide by
`kube-state-metrics` (`kube_job_status_*`, `kube_cronjob_*` - bundled
with the `kube-prometheus-stack` chart [0016](0016-observability-backend.md)
explicitly names), so alerting on a failed job rides on that existing
signal, not new per-job scrape config. logs are unchanged - stdout to
Loki, same as any other pod.

## deliberately not built yet

matches [0005](0005-helm-chart-split.md#topology)'s own stance -
templates and values land when a real service needs one, not
speculatively. tracked in
[docs/roadmap.md](../roadmap.md#devx).

## not decided here

- retry/backoff specifics for any particular job - that's per-workload,
  decided by whoever adds the real one, not generically here.
- RBAC scoping beyond what `deployment` mode's `ServiceAccount`/`rbac.yaml`
  already provide - no evidence yet a job needs anything narrower or
  broader.
- `schedule` timezone handling (`CronJob.spec.timeZone` vs UTC-only) -
  no concrete schedule exists yet to decide against.
