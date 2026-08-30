# 0016: observability backend - Prometheus+Grafana+Alertmanager (metrics/alerts), Loki (logs), Tempo (tracing), all self-hosted in `platform`; Sentry dropped; per-customer signals deferred

four pillars, each doing the one thing it's actually for - no tool
covering two out of a mistaken sense of economy. Sentry conflated
tracing with error-tracking; splitting them is the point of this
revision.

## tracing: Tempo, not Sentry

Sentry's tracing (`go/internal/lib/tracing`'s `sentry.StartSpan`/
`sentry.CaptureException`/`sentryhttp` today) is being dropped, not kept
alongside a new metrics/logs stack - two separate observability
vendors/backends for one org is exactly the kind of redundancy this repo
avoids elsewhere (one registry, one secrets backend). tracing's whole
value is triaging and visualizing calls across services - real
distributed tracing, not a side effect of an error-tracking product.
Tempo joins the same Grafana instance Loki/Prometheus already sit in
(0016's original metrics/logs half, unchanged below) - one pane of glass
for all three pillars instead of a fourth, separate SaaS console.

**this isn't a live migration with a coverage gap - Sentry isn't
actually delivering anything today.** checked directly:
[go/internal/lib/config/tracing.go](../../go/internal/lib/config/tracing.go)'s
`IngestionURL` (the DSN) has no default and nothing sets
`HYRULE_TRACING_INGESTION_URL` anywhere in this repo - `sentry.Init`
runs with an empty `Dsn`, which the SDK treats as a no-op, nothing ever
sent. the SDK is wired but has been net-zero value the whole time. that
changes what "dropping it" means: there's no working capability to lose
in the gap before Tempo exists, just dead code to remove.

this is a real code change, not just a backend swap: `go/internal/lib/tracing`
moves off the `sentry-go` SDK onto OpenTelemetry instrumentation (OTLP
export to Tempo). [conventions.md#tracing](../conventions.md#tracing)'s
span-naming rule (`<component>.<action>[.<qualifier]`, dots only) and the
three-tier helper shape (`Start`/`StartNamedf`/`StartNamed`) carry over
unchanged - only what's underneath the helpers changes. not built as
part of this decision doc; tracked in
[docs/roadmap.md](../roadmap.md#infrastructure) alongside actually
installing Tempo.

## errors: logging is the error-tracking surface, no separate issue-tracking product

no Sentry, no self-hosted substitute (GlitchTip, self-hosted Sentry) -
"what must be raised as a meaningful error" is `slog`'s job
([go/internal/lib/runtime/logging.go](../../go/internal/lib/runtime/logging.go)),
not a dedicated error-tracker's. an error worth surfacing gets logged
with enough structure to alert on (matches
[conventions.md](../conventions.md)'s existing "errors wrapped with `%w`
and enough context to debug without a stack trace" convention) and, once
Loki's in place, correlated to its trace via `trace_id`/`span_id` on the
log line - the same request's story lives in both places, joined by that
id, without a third product doing deduplication/grouping on top. if
issue-grouping/dedup becomes a real pain point later, that's a fresh
decision against a concrete complaint, not something to build speculatively
here.

## metrics: Prometheus + Grafana, self-hosted in-cluster - unchanged

not Sentry (APM/error tracking, not a general infra metrics store) and
not a cloud-managed stand-in - matches the standing bias in this repo of
running real infra rather than substituting a managed service where a
self-hosted one does the job ([0010](0010-image-registry-and-publishing.md)'s
Harbor, [0011](0011-secrets-generalized.md)'s 1Password Connect + External
Secrets Operator both chose the self-hosted path over a cloud
alternative). Prometheus + Grafana is the standard k3s/homelab pairing,
well-trodden enough that there's no real alternative worth spiking.

**the install vehicle is named explicitly: the `kube-prometheus-stack`
Helm chart** (`prometheus-community/kube-prometheus-stack`), not
standalone Prometheus/Grafana/Alertmanager charts glued together by
hand. it's the de facto standard for exactly this pairing on k3s/homelab
clusters, and it bundles three things this doc already wants anyway:
Alertmanager (below), `kube-state-metrics` (what
[0017](0017-scheduled-workloads.md)'s job/cronjob observability section
depends on existing), and `node-exporter` (node-level CPU/memory/disk,
not decided elsewhere but free with this chart and generically useful
on homelab hardware). naming the chart here closes the gap 0017 would
otherwise have left open - "kube-state-metrics exists" stops being an
assumption about how 0016 gets built and becomes part of what it
decided.

## alerting: Alertmanager, wired to Prometheus - now in scope

metrics existing without anything acting on them isn't the goal - "track
system KPIs/health" means something pages or notifies when they cross a
threshold. Alertmanager (Prometheus's own alerting component) is in
scope as of this revision, superseding the original doc's "no alerting
story decided" punt. concrete alert rules and notification routing
(Slack? email? something else) are content, not shape, and stay
undecided here - same distinction as "dashboards - content, not shape"
below.

## logs: Loki, paired with Grafana - unchanged

every service already emits structured JSON via `slog` to stdout - Loki
(with Promtail or Grafana Alloy as the shipper) scrapes container stdout
directly, so this needs zero application-code changes beyond whatever
`trace_id`/`span_id` correlation the Tempo migration adds, only
in-cluster plumbing.

## signals (per-customer/business metrics, e.g. Superset) - deferred

a fourth, distinct concern from the three pillars above: per-customer
KPIs and business-facing analytics, not system health. explicitly not
decided here - same reasoning [docs/roadmap.md](../roadmap.md)'s product
section already gives for not guessing at homelab's product surface:
there's no real customer-facing product yet to build per-customer
metrics against. revisit once one exists.

## `platform` grows scrape-config/agent plumbing, same shape as 0011

[0005](0005-helm-chart-split.md#what's-in-app-vs-platform) already
defined `platform` as the homelab-plumbing chart, and
[0011](0011-secrets-generalized.md) already set the pattern for how a
CRD-driven piece of infra plumbing attaches to a service: a
`ServiceMonitor`, `PrometheusRule`, or log-shipper equivalent is
homelab-specific plumbing, not part of `app`'s generic shape, so it goes
into `platform` via the `additionalK8sObjects` escape hatch - never a
template added directly to `app`. this is the same reasoning 0011 used
for the `ExternalSecret` CRD, not a new pattern.

## not decided here

- actually installing Prometheus/Grafana/Loki/Tempo/Alertmanager (which
  charts, which values, local cluster vs CP vs both) and migrating
  `go/internal/lib/tracing` off `sentry-go` onto OpenTelemetry - real
  infrastructure and code work, not something a decision doc can supply,
  same reasoning [0011](0011-secrets-generalized.md) and
  [0019](0019-postgres-backup-dr.md) used to punt their own installs.
  tracked in [docs/roadmap.md](../roadmap.md#infrastructure).
- retention/storage sizing for Prometheus, Loki, or Tempo.
- concrete alert rules and notification routing.
- per-customer signals/analytics (see above) - needs a real product
  surface first.
- dashboards - content, not shape.
