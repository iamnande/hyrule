# Tech Spec — ArgoCD + GitOps foundation (scalable shape, proven locally)

## Overview + Design Principles

installs ArgoCD into the local cluster and proves the shape that scales to ~60 services: a
services registry drives generated `AppProject`s and a single `matrix`-generator `ApplicationSet`,
instead of hand-writing an `AppProject`+`ApplicationSet` pair per service. proven today against
the one service that exists (`pings`), sized so a second service or a second region is a data
change (a registry line, a labeled cluster Secret), not new ArgoCD YAML. the migrate Job becomes
a real, opt-in `app` chart primitive per [0025](../../decisions/0025-migrate-as-chart-primitive.md),
unifying Tilt's local path onto the same definition instead of a hand-maintained duplicate.

- **the registry is the one source of truth for "what's live"** - `local/argocd/appprojects/services.yaml`
  lists every registered service (`service`, `team`, `domain`). deliberately decoupled from
  `deploy/values/<service>/` existing: a service can be scaffolded (`make new-service`) without
  being registered, an explicit gate rather than implicit-by-directory. Tilt's own discovery
  (`ls deploy/values`) stays untouched - a different question (what's scaffolded, not what's
  live in Argo), not a duplicate mechanism to keep in sync
- **`AppProject`s are generated, never hand-written, one per service** - project identity is the
  stable axis (services rarely get renamed; team ownership changes constantly), so the project
  stays keyed to the service name. `team`/`domain` land as labels on it, not the boundary itself -
  ready to drive a generated `argocd-rbac-cm` policy mapping later (not built this pass, see PRD
  § Scope & Non-Goals) without ever restructuring a project when ownership changes
- **one `ApplicationSet` for every service, not one per service** - a `matrix` generator crosses
  the registry (`git` generator, file mode) with the existing `clusters` generator (region
  labels). the collision this requires avoiding: both a registry entry and a matched cluster can
  expose a field called `name` (service name vs. cluster name) - the registry's field is called
  `service`, not `name`, specifically so the matrix generator's merged param set never collides
- **region/environment identity is generator-derived, never committed to a values file** - per
  [0026](../../decisions/0026-region-environment-argo-derived.md), `HYRULE_REGION`/
  `HYRULE_ENVIRONMENT` are `helm.parameters` sourced from the matched cluster's own labels. the
  cluster registration Secret (`cluster-local.yaml`) is the single place either value is asserted
- **migrate moves into `deploy/helm/app` as an opt-in template, not a second Application source**
  - per [0025](../../decisions/0025-migrate-as-chart-primitive.md), narrowing the first draft's
    stated principle that it "avoids prematurely committing prototype-only shape to the chart."
    that principle held when migrate was one-off, single-service tooling; it doesn't once any of
    ~60 services needs to opt into its own migrate step. declaring a `datastores` entry
    (`flavor: postgres`) gates it, independent of `mode:` - a service stays `mode: deployment` and
    separately opts a migrate Job in by naming a datastore, not by flipping a standalone flag
- **a `datastores` list is the one opt-in surface, generalized past migrate on purpose** - each
  entry is `slug` + `flavor` + flavor-specific keys (`schema_management` for `flavor: postgres`,
  defaulted to `enabled`). only `flavor: postgres` does anything this pass; other flavors are
  shape-only, so a second one (a `key-value` cache, say) lands additively later instead of a
  values-schema rewrite. a service engineer opts into schema management by declaring the datastore
  itself - no separate mechanism to learn
- **postgres connection facts are a single chart default, never a per-service value** - `host`,
  `port`, the owner role and its credentials live once, in `deploy/helm/app/values.yaml`. a
  platform engineer repoints every opted-in service at a new instance with one edit to one file;
  no service's own values ever carries a connection string, and no generated code sits between the
  edit and every service picking it up. same mechanism [0026](../../decisions/0026-region-environment-argo-derived.md)
  already uses for region/environment - a chart default, not a second place the same fact could
  drift from. the owner credentials stay plaintext for this pass (same risk profile as today's
  already-committed `hyrule_owner`/`hyrule_owner`, just centralized) -
  [0011](../../decisions/0011-secrets-generalized.md) is where that's actually solved, not here
- **migrate targets a per-service logical database on the one shared postgres instance** -
  enough to give the chart primitive a real per-service target without deciding
  [0023](../../decisions/0023-postgres-replication-topology.md)'s broader physical-topology
  question. the migrate template builds its own `-database=` DSN from the chart-level postgres
  default crossed with the datastore entry's own `slug` (the database name) - a service supplies
  neither a connection string nor a ConfigMap name (chart-derived, `{{ include "app.fullname" . }}-migrations`),
  so per-service RLS/RBAC work still lands additively later instead of as a rewrite
- **the `hyrule-migrations`-style ConfigMap still isn't re-managed by this pass** - now scoped
  per-service-namespace instead of `databases` only, same accepted limitation as the first draft:
  populating it (a Kustomize generator, `configmap_from_dir`'s equivalent under ArgoCD) is new
  mechanism for zero proof value this pass cares about
- **postgres itself stays in `databases`; only the app moves** - unchanged from the first draft.
  each service's migrate Job now runs *in that service's own namespace* (not `databases`),
  reaching postgres over the network the same way the app's own DB connection already does - so
  it no longer needs a second, explicit-namespace resource the way the first draft's shared
  cross-namespace hook Job did

---

## Data Model

N/A — no new application data model. the new artifacts are ArgoCD's own CRDs (`ApplicationSet`,
`AppProject`, `Application`), a plain YAML registry (`services.yaml`, not a CRD), a small local
Helm chart that renders `AppProject`s, one `Secret` registering the local cluster, and a new
`migrate.yaml` template in `deploy/helm/app` - configuration objects, not data this repo's
services read or write. `pings` gains one new logical database on the existing shared postgres
instance (§ Core, § Files Changed) - a deployment-topology fact, not a new entity.

---

## Interfaces

**`local/argocd/appprojects/services.yaml`** (new) - the registry:

```yaml
- service: pings
  team: platform
  domain: core
```

`service`, not `name` - avoids colliding with the `clusters` generator's own `{{name}}` param
once both feed the same `matrix` generator (see Overview). lives inside the `appprojects` chart
directory (not `local/argocd/` directly, as first drafted) - see that chart's own entry below for
why: it's the one file both the chart and the services `ApplicationSet`'s git generator read, and
each needs it in a different shape.

**`local/argocd/cluster-local.yaml`** (updated) - gains an `environment` label alongside
`region`, per [0026](../../decisions/0026-region-environment-argo-derived.md):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cluster-local
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: cluster
    region: us-west-2
    environment: local
type: Opaque
stringData:
  name: local
  server: https://kubernetes.default.svc
  config: '{"tlsClientConfig":{"insecure":false}}'
```

`region: us-west-2` rather than a new `local`/`region: n/a` value - unchanged reasoning from the
first draft (proves the generator's selector is the literal one a real CP cluster registration
will carry later). `environment: local` is not a stand-in the same way - `local` is one of
[0009](../../decisions/0009-environments-and-promotion.md)'s three real environment values, so
this is the actual, correct label for this cluster, not a placeholder.

**`local/argocd/appprojects/`** (new) - a small chart, no org-specific logic beyond "one
`AppProject` per registry entry":

```yaml
# Chart.yaml
apiVersion: v2
name: appprojects
description: generates one AppProject per entry in the services registry
type: application
version: 0.1.0
```

```yaml
# templates/appproject.yaml
{{- $services := .Files.Get "services.yaml" | fromYamlArray }}
{{- range $services }}
---
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: {{ .service }}
  namespace: argocd
  labels:
    team: {{ .team }}
    domain: {{ .domain }}
spec:
  sourceRepos:
    - https://github.com/iamnande/hyrule
  destinations:
    - namespace: {{ .service }}
      server: '*'
  clusterResourceWhitelist: []
  namespaceResourceWhitelist:
    - group: '*'
      kind: '*'
{{- end }}
```

reads its own `services.yaml` via `.Files.Get` + `fromYamlArray`, not `.Values` - discovered live
against a real ArgoCD sync (execution phase, not caught by `helm template` alone during planning):
Helm's values loader requires a values file's root to be a map, but the services `ApplicationSet`'s
git generator (below) requires the *same physical file* to be a top-level array to generate one
param set per entry (R4) - passing it as `helm.valueFiles` fails outright
(`cannot unmarshal array into Go value of type map[string]interface {}`). moving the registry
inside the chart and reading it via `.Files.Get` keeps one source of truth in its required array
shape for both consumers, without a second copy or a reshaping step. `destinations.server: '*'` -
cluster scoping is already enforced by the services `ApplicationSet`'s own `clusters` generator
selector; the project only needs to fix *which namespace*, not which cluster.

**`local/argocd/application-projects.yaml`** (new) - the one Application that renders every
`AppProject` from the registry, applied directly by `argocd-up` (same as `cluster-local.yaml`),
not itself generated:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: argocd-projects
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/iamnande/hyrule
    targetRevision: prototype/0002-argocd-gitops
    path: local/argocd/appprojects
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy: {}
```

`project: default` - ArgoCD's built-in project, unchanged. this Application is platform bootstrap
infra, the same tier as registering the cluster itself; it can't reasonably belong to a
per-service project it's the thing generating. a single `source`, not the first draft's two - the
chart no longer needs an externally injected values file (see `appprojects/` above), so there's
nothing for a second `ref: values` source to supply.

**`local/argocd/applicationset-services.yaml`** (new) - the one `ApplicationSet` for every
service, `matrix` generator, replaces the first draft's per-service `appproject-pings.yaml` +
`applicationset-pings.yaml` pair entirely:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: services
  namespace: argocd
spec:
  syncPolicy:
    preserveResourcesOnDeletion: true
  generators:
    - matrix:
        generators:
          - git:
              repoURL: https://github.com/iamnande/hyrule
              revision: prototype/0002-argocd-gitops
              files:
                - path: local/argocd/appprojects/services.yaml
          - clusters:
              selector:
                matchLabels:
                  region: us-west-2
  template:
    metadata:
      name: '{{service}}-{{name}}'
    spec:
      project: '{{service}}'
      sources:
        - repoURL: https://github.com/iamnande/hyrule
          targetRevision: prototype/0002-argocd-gitops
          path: deploy/helm/app-platform
          helm:
            releaseName: '{{service}}'
            valueFiles:
              - $values/deploy/values/{{service}}/values.yaml
            parameters:
              - name: app.env.HYRULE_REGION
                value: '{{metadata.labels.region}}'
              - name: app.env.HYRULE_ENVIRONMENT
                value: '{{metadata.labels.environment}}'
        - repoURL: https://github.com/iamnande/hyrule
          targetRevision: prototype/0002-argocd-gitops
          ref: values
      destination:
        server: '{{server}}'
        namespace: '{{service}}'
      syncPolicy: {}
```

`{{service}}` from the git-files side, `{{name}}`/`{{server}}`/`{{metadata.labels.*}}` from the
`clusters` side - the matrix generator merges both parameter sets, no collision (see Overview).
the ApplicationSet's own `syncPolicy.preserveResourcesOnDeletion: true` - discovered live on
`argocd-down`, not something a sync-only test surfaces: without it, the ApplicationSet controller
auto-adds `resources-finalizer.argocd.argoproj.io` to every Application it generates, which exists
specifically to cascade-delete that Application's live resources when the Application itself is
deleted. `helm uninstall`ing ArgoCD deletes the application-controller pod that would process that
finalizer before `kubectl delete namespace argocd` gets to `pings-local`, so the delete hangs
forever instead of either completing or (worse) actually cascading into `pings`. this is the
opposite of R8/the PRD's explicit "registered service namespaces untouched" - `preserveResourcesOnDeletion`
is what makes that safety deliberate instead of an accidental deadlock.

`helm.releaseName: '{{service}}'` - discovered live, on the real adoption attempt R8 asks for:
without it, ArgoCD defaults `.Release.Name` to the Application's own name (`pings-local`), which
flows into `app.selectorLabels` (`app.kubernetes.io/name`/`app.kubernetes.io/instance`) and
mismatches the `app.kubernetes.io/name: pings`/`instance: pings` Tilt's own `helm_resource` release
(named `pings`) already set on the running Deployment - `spec.selector` is immutable, so the sync
fails outright trying to change it. pinning the release name to the service name is what makes
Argo's render match what's already running, letting adoption succeed.
`project: '{{service}}'` - the `AppProject` `application-projects.yaml` generates. `syncPolicy: {}`
- unchanged, still manual-sync-only.

**`deploy/helm/app/values.yaml`** (updated) - `datastores` replaces the first draft's `migrate:`
block as the opt-in surface; postgres connection facts land as their own top-level default,
`migrate:` narrows to job mechanics only (image, `backoffLimit` - no `enabled`, no `database`, no
`migrationsConfigMap`):

```yaml
datastores: []

postgres:
  host: postgres-rw.databases.svc.cluster.local
  port: 5432
  ownerUser: hyrule_owner
  ownerPassword: hyrule_owner
  sslMode: disable

migrate:
  image:
    repository: docker.io/migrate/migrate
    tag: v4.19.1
  backoffLimit: 0

env:
  HYRULE_ENVIRONMENT: local
```

`postgres:` is the one file a platform engineer edits to repoint every opted-in service - see
§ Overview. the chart-level `HYRULE_ENVIRONMENT: local` default is what Tilt actually runs with -
it never goes through the `ApplicationSet`'s `helm.parameters` (see § Core). ArgoCD-managed
Applications always override it per-sync.

**`deploy/helm/app/templates/migrate.yaml`** (new) - self-contained pod spec per
[0025](../../decisions/0025-migrate-as-chart-primitive.md), not wired into `app.pod`/`_pod.tpl`
(hardcodes an `http` port + probes wrong for a Job) or 0017's still-unbuilt
`app.pod.container`/`_pod-job.tpl` split. ranges over `.Values.datastores`, rendering a Job per
`flavor: postgres` entry not opted out via `schema_management: disabled`:

```yaml
{{- range .Values.datastores }}
{{- if and (eq .flavor "postgres") (ne .schema_management "disabled") }}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "app.fullname" $ }}-migrate
  labels:
    {{- include "app.selectorLabels" $ | nindent 4 }}
  annotations:
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
spec:
  backoffLimit: {{ $.Values.migrate.backoffLimit }}
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: migrate
          image: "{{ $.Values.migrate.image.repository }}:{{ $.Values.migrate.image.tag }}"
          args:
            - -path=/migrations
            - -database=postgres://{{ $.Values.postgres.ownerUser }}:{{ $.Values.postgres.ownerPassword }}@{{ $.Values.postgres.host }}:{{ $.Values.postgres.port }}/{{ .slug }}?sslmode={{ $.Values.postgres.sslMode }}
            - up
          volumeMounts:
            - name: migrations
              mountPath: /migrations
      volumes:
        - name: migrations
          configMap:
            name: {{ include "app.fullname" $ }}-migrations
{{- end }}
{{- end }}
```

`.slug` is the only per-entry fact the DSN needs - it doubles as the logical database name, same
as `pings`' database is named `pings` today. `$` reaches the root context from inside `range` for
`.Values.postgres`/`.Values.migrate` and the `app.fullname`/`app.selectorLabels` helpers. naming
(`{{ app.fullname }}-migrate`/`-migrations`, no `slug` woven in) assumes exactly one
`flavor: postgres` entry per service - true everywhere today; a second one would collide and isn't
handled (see [0025](../../decisions/0025-migrate-as-chart-primitive.md)). the two
`argocd.argoproj.io/*` annotations are inert metadata under plain `kubectl apply`/Tilt's
`helm template` path - meaningful only once ArgoCD owns the sync.

**`mk/argocd.mk`** (new) - bootstrap order gains the projects Application and the services
`ApplicationSet`:

```makefile
ARGOCD_NAMESPACE := argocd
ARGOCD_CHART_REF := 10.8.1 # argo-cd, appVersion v3.5.2

.PHONY: argocd-up
argocd-up: ## argocd: install argocd + bootstrap the registry-driven project/service generators
	@echo $(PROJECT_LOG_FMT) "installing argocd"
	@helm repo add argo https://argoproj.github.io/argo-helm --force-update
	@helm upgrade --install argocd argo/argo-cd \
		--version $(ARGOCD_CHART_REF) \
		--namespace $(ARGOCD_NAMESPACE) --create-namespace
	@kubectl apply -f local/argocd/cluster-local.yaml
	@kubectl apply -f local/argocd/application-projects.yaml
	@kubectl apply -f local/argocd/applicationset-services.yaml

.PHONY: argocd-down
argocd-down: ## argocd: tear down argocd wholesale (registered service namespaces untouched)
	@echo $(PROJECT_LOG_FMT) "tearing down argocd"
	@helm uninstall argocd --namespace $(ARGOCD_NAMESPACE) --ignore-not-found
	@kubectl delete namespace $(ARGOCD_NAMESPACE) --ignore-not-found

.PHONY: argocd-ui
argocd-ui: ## argocd: port-forward the argocd UI/API to localhost:8080
	@kubectl port-forward svc/argocd-server -n $(ARGOCD_NAMESPACE) 8080:443
```

**`Tiltfile`** (updated) - `local_migrate()` becomes per-service, sourced from the chart instead
of a hand-maintained duplicate:

```python
def has_postgres_datastore(datastores):
    return any(
        d.get('flavor') == 'postgres' and d.get('schema_management', 'enabled') != 'disabled'
        for d in datastores
    )


def local_migrate(slug):
    values_path = 'deploy/values/%s/values.yaml' % slug
    values = read_yaml(values_path)
    datastores = values.get('app', {}).get('datastores', [])
    if not has_postgres_datastore(datastores):
        return
    migrations_configmap = '%s-migrations' % slug  # matches app.fullname-derived naming
    k8s_yaml(configmap_from_dir(
        migrations_configmap, 'migrations/%s' % slug, namespace=slug,
    ))
    rendered = local(
        'helm template %s deploy/helm/app-platform --values=%s '
        '--show-only charts/app/templates/migrate.yaml' % (slug, values_path),
        quiet=True,
    )
    k8s_yaml(blob(str(rendered)))
    k8s_resource(
        '%s-migrate' % slug,
        objects=['%s:configmap' % migrations_configmap],
        resource_deps=['postgres'],
        labels=['local-only'],
    )


def deploy_service(slug):
    values_path = 'deploy/values/%s/values.yaml' % slug
    values = read_yaml(values_path)
    needs_migrate = has_postgres_datastore(values.get('app', {}).get('datastores', []))

    helm_resource(
        name=slug,
        chart='deploy/helm/app-platform',
        release_name=slug,
        namespace=slug,
        flags=['--values=%s' % values_path, '--create-namespace', '--set-json=app.datastores=[]'],
        image_deps=['%s/%s' % (ORG_NAME, slug)],
        image_keys=[('app.image.repository', 'app.image.tag')],
        resource_deps=['%s-migrate' % slug] if needs_migrate else [],
        port_forwards=[port_forward(local_port=0, container_port=8000)],
        labels=['deploy'],
    )
    # ... k8s_resource(...) unchanged
```

`--set-json=app.datastores=[]` replaces the first draft's `--set=app.migrate.enabled=false` -
blanking the whole list is what suppresses every migrate Job the service's own release would
otherwise render, now that opt-in is per-datastore rather than one flag. `local_migrate()`'s
separate `helm template --show-only` extraction (full values, `datastores` intact) is what
actually renders and runs it, same division of responsibility as the first draft.

```python
local_database()
for slug in SERVICE_SLUGS:
    local_migrate(slug)
for slug in SERVICE_SLUGS:
    build_service(slug)
for slug in SERVICE_SLUGS:
    deploy_service(slug)
```

`deploy_service`'s own `helm_resource` call forces `app.datastores=[]` unconditionally -
under Tilt, migrate stays a separate, manually-re-triggerable resource (same reason
[0012](../../decisions/0012-migrations-under-gitops.md) needed an ArgoCD-side PreSync hook in the
first place: a Job is immutable once created, and `helm upgrade` re-running on every code change
would otherwise fail against it). under ArgoCD, `datastores` from the service's own committed
values file applies as-is - the PreSync hook's `BeforeHookCreation` delete policy is what makes a
fresh Job on every sync safe there, a problem Tilt's separate-resource model sidesteps
differently, not by adopting the same mechanism. one chart definition, two different
orchestration strategies around it - matches "Tilt is orchestration, the chart defines what
should exist."

**`local/postgres/init/03-pings-database.sql`** (new) - `pings` gets its own logical database on
the one shared instance:

```sql
CREATE DATABASE pings OWNER hyrule_owner;
GRANT CONNECT ON DATABASE pings TO hyrule_app;
GRANT CONNECT ON DATABASE pings TO hyrule_app_ro;
```

table/sequence grants inside `pings` itself land via the existing `ALTER DEFAULT PRIVILEGES...`
pattern in `01-app-role.sql`/`02-app-ro-role.sql`, scoped per-database by Postgres already -
`hyrule_owner` running `pings`' migrations inside the `pings` database is what triggers those
default privileges there, same as it already does for `hyrule`. automating this per registry
entry (an init script generated from `services.yaml`, matching the "generated not hand-written"
approach used for `AppProject`s) is explicitly not built this pass - one entry, `pings`, hand-added
- see PRD § Scope & Non-Goals.

**`deploy/values/pings/values.yaml`** (updated):

```yaml
app:
  fullnameOverride: pings

  image:
    repository: iamnande/pings
    tag: latest

  service:
    port: 8000

  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      memory: 128Mi

  env:
    HYRULE_HTTP_SERVER_ADDR: ":8000"
    HYRULE_DATABASE_HOST: postgres-rw.databases.svc.cluster.local
    HYRULE_DATABASE_PORT: "5432"
    HYRULE_DATABASE_USER: hyrule_app
    HYRULE_DATABASE_PASSWORD: hyrule_app
    HYRULE_DATABASE_NAME: pings
    HYRULE_DATABASE_SSL_MODE: disable
    HYRULE_PINGS_STALE_AFTER: 5m

  datastores:
    - slug: pings
      flavor: postgres

platform: {}
```

`HYRULE_REGION`/`HYRULE_ENVIRONMENT` are gone - Argo-derived (0026), chart-default under Tilt.
`HYRULE_DATABASE_NAME` changes from the shared `hyrule` to `pings`'s own logical database -
`iam-jwks` is unaffected, it never set `HYRULE_DATABASE_HOST` and isn't registered.
`schema_management` is omitted - defaults to `enabled`, per § Interfaces - so this entry is the
whole opt-in: no `migrate.enabled`, no connection string, no ConfigMap name. `HYRULE_DATABASE_*`
above is `pings`' own runtime config for its own connection (the chart has no opinion on a
service's env var naming, per 0005) - a separate fact from `datastores`, which only feeds the
migrate Job the chart itself owns; the two happen to point at the same database today because
that's what `pings` needs, not because one derives from the other.

**`migrations/`** → **`migrations/pings/`** (moved, content unchanged) - the existing
`000001_create_pings.up.sql`/`.down.sql` already only ever concerned `pings`; this just makes
that explicit in the path instead of a repo-root directory implying a shared schema that no
longer exists once `pings` has its own logical database.

**`local/cluster/migrate-job.yaml`** - deleted. superseded by `deploy/helm/app/templates/migrate.yaml`
per [0025](../../decisions/0025-migrate-as-chart-primitive.md).

---

## Validation

- **`helm template` render-ability** - `deploy/helm/app-platform` + `deploy/values/pings/values.yaml`
  renders cleanly with a `flavor: postgres` `datastores` entry and with `datastores: []`; Tilt's
  own `helm template --show-only` extraction (§ Interfaces) is itself a render-ability check on
  every `tilt up`
- **`AppProject`/`ApplicationSet` generation correctness** - `argocd app diff argocd-projects`
  (or the UI) after a sync shows exactly one `AppProject` per `services.yaml` entry; the services
  `ApplicationSet`'s own status shows exactly one generated `Application` per service × matched
  cluster (today: `pings-local`, one match)
- **`AppProject` destination/repo scoping** - a sync targeting any namespace other than the
  project's own service name, or any repo other than this one, is rejected by ArgoCD before it
  reaches the cluster - `namespaceResourceWhitelist`/`sourceRepos` are the enforcement point
- **PreSync hook ordering** - unchanged from the first draft: ArgoCD's own hook contract
  guarantees the migrate Job reaches `Succeeded` before the rest of the sync's resources are
  applied; a failed hook blocks the sync outright
- **generator param collision** - `local/argocd/applicationset-services.yaml`'s rendered
  `Application`s (`argocd appset generate`, or the live `ApplicationSet` status) show `{{service}}`
  resolving to the registry's service name and `{{name}}`/`{{server}}` resolving to the cluster's
  own identity, not a collision/overwrite between the two
- **region/environment never committed** - `deploy/values/pings/values.yaml` has no
  `HYRULE_REGION`/`HYRULE_ENVIRONMENT` keys; the generated `Application`'s rendered manifest
  (`argocd app manifests pings-local`) shows both set correctly from `cluster-local.yaml`'s labels
- **adoption correctness** - no automated check here; R7's manual verification (first sync leaves
  no orphaned/conflicting resources) is the validation, per the PRD's own framing that this is
  what's unverified going in
- **bootstrap ordering** - syncing `argocd-projects` before `pings-local` is a manual sequencing
  step this pass relies on a human to get right (§ Core) - not itself enforced by ArgoCD; a
  service's generated `Application` sits in a validation-error state (not a crash, not data loss)
  if synced first, and resolves itself once `argocd-projects` is synced and its project exists

---

## Core

**1. bootstrap order** (`make argocd-up`): install the chart -> apply `cluster-local.yaml`
(labels the in-cluster destination `region`/`environment`) -> apply `application-projects.yaml`
(not yet synced - `OutOfSync`) -> apply `applicationset-services.yaml`, which immediately
generates `pings-local`'s `Application` (matrix match against the registry's one entry × the
local cluster) - also `OutOfSync`, and additionally in a project-validation-error state until
step 2 below.

**2. first sync order matters:** `argocd app sync argocd-projects` first - this is what actually
creates `pings`' `AppProject` from its `services.yaml` entry. only after that does
`argocd app sync pings-local` become a normal, syncable Application instead of one referencing a
nonexistent project. this ordering isn't enforced by ArgoCD itself (§ Validation) - it's a
one-time manual step per cluster bootstrap, not a per-service one (adding a second service later
never needs a re-sync of `argocd-projects` to matter in the same blocking way, since the project it
needs is generated the moment its registry entry exists and `argocd-projects` next syncs on any
trigger).

**3. stopping Tilt before a sync:** revised from the first draft's assumption - `pings` (per the
PRD) is stopped, so there's exactly one controller writing to that namespace at a time. `postgres`
keeps running; it's what both the app and the migrate hook connect to. Tilt's `<slug>-migrate`
resource (now per-service) does **not** implicitly stop with it, though - discovered live, not
caught by inspection alone: `resource_deps` (`deploy_service()`'s dependency on `<slug>-migrate`)
only orders builds, it doesn't propagate enable/disable state, and both Tilt's rendered migrate
Job and ArgoCD's PreSync-hook Job are the exact same object (same `{{ app.fullname }}-migrate`
name, same namespace). `<slug>-migrate` needs stopping explicitly alongside `<slug>` - the first
draft's global `migrate` resource needed the same explicit stop, this per-service split didn't
change that.

**4. first sync / adoption (R7):** `argocd app sync pings-local` triggers, in order: the PreSync
hook Job (now `pings`' own `deploy/helm/app` migrate template, rendered as part of the
`app-platform` release) runs to completion against `pings`' own logical database and is deleted
per `BeforeHookCreation` -> the `app-platform` release's rendered manifests are applied against
the existing `pings` namespace. same adoption-correctness caveats as the first draft (ArgoCD's
diff/apply is live-manifest-based, not Helm-release-secret-based - Tilt's leftover release secret
is inert clutter, not a conflict); what's unverified per R7 is the same "does every field ArgoCD's
server-side apply wants to own conflict with a stale Tilt-set field" question, resolved by
observing the sync result directly.

**5. values change / re-sync (R7):** pushing a `deploy/values/pings/values.yaml` change to the
feature branch, then a manual sync, re-renders the Helm source (base values + the
`ApplicationSet`'s own `helm.parameters` for region/environment) and applies the diff. the
PreSync hook still runs first (fresh Job, `migrate` is idempotent).

**6. registering a second service:** add a `deploy/values/<service>/` (already how `make
new-service` scaffolds today) plus one line to `local/argocd/appprojects/services.yaml`. the next sync of
`argocd-projects` generates its `AppProject`; the services `ApplicationSet` immediately generates
its `Application` against every already-matched cluster (today: just the local one). no new
ArgoCD YAML written by hand - the whole point of this revision.

**7. `argocd-down`:** unchanged from the first draft - uninstalling the chart and deleting the
`argocd` namespace removes every ArgoCD-owned object (the registry-generated `AppProject`s and
`Application`s included) and ArgoCD itself. does **not** touch `pings`/`databases` - un-owned
plain cluster objects at that point. restarting Tilt's `pings` resource resumes normal local dev.

---

## Query Layer

N/A — no new query layer. the migrate hook Job's SQL access is unchanged in mechanism (the
existing `migrate` CLI image), now targeting `pings`' own logical database instead of the shared
`hyrule` one.

---

## API Layer

N/A — no new hyrule-owned API. ArgoCD's own UI/API is port-forwarded via `make argocd-ui`, used
for observing sync/health state only, per the PRD's Observability section.

---

## Files Changed

| file | change |
|---|---|
| `mk/argocd.mk` | new — `argocd-up`/`argocd-down`/`argocd-ui`, bootstrap order includes the projects Application + services `ApplicationSet` |
| `local/argocd/appprojects/services.yaml` | new — the services registry (service, team, domain) |
| `local/argocd/cluster-local.yaml` | new — labels the in-cluster destination `region`/`environment` |
| `local/argocd/appprojects/Chart.yaml`, `local/argocd/appprojects/templates/appproject.yaml` | new — chart ranging over the registry, one `AppProject` per entry |
| `local/argocd/application-projects.yaml` | new — the one Application applying the `appprojects` chart |
| `local/argocd/applicationset-services.yaml` | new — the one `ApplicationSet`, `matrix` generator, replaces per-service `AppProject`/`ApplicationSet` authoring |
| `deploy/helm/app/values.yaml` | updated — `datastores: []` opt-in list, `postgres:` chart-default connection facts, `migrate:` narrowed to job mechanics only, `env.HYRULE_ENVIRONMENT: local` chart default |
| `deploy/helm/app/templates/migrate.yaml` | new — opt-in PreSync-hook Job template per `datastores` entry, self-contained pod spec |
| `deploy/values/pings/values.yaml` | updated — drops `HYRULE_REGION`/`HYRULE_ENVIRONMENT`, `HYRULE_DATABASE_NAME` → `pings`, adds a `datastores` entry (`slug: pings`, `flavor: postgres`) - no connection string, no ConfigMap name |
| `migrations/` → `migrations/pings/` | moved — makes the existing pings-only content's scope explicit |
| `local/postgres/init/03-pings-database.sql` | new — creates `pings`' own logical database + grants |
| `local/cluster/migrate-job.yaml` | deleted — superseded by the chart-native template |
| `Tiltfile` | updated — `local_migrate()` per-service, gated on a `flavor: postgres` `datastores` entry, sourced from the chart via `helm template`; `deploy_service()` forces `app.datastores=[]` on its own release, depends on the separate per-service migrate resource instead |
| `docs/decisions/0025-migrate-as-chart-primitive.md` | new — narrows 0017 |
| `docs/decisions/0026-region-environment-argo-derived.md` | new — narrows 0009 |
| `docs/roadmap.md` | add "ArgoCD + GitOps foundation" as its own numbered initiative |

---

## Integration Tests

no automated suite - this is a manual local prototype, per the PRD's own scope. verification
steps, in order:

1. `make argocd-up` - confirm the `argocd` namespace comes up healthy, `cluster-local.yaml` is
   present with `region: us-west-2`/`environment: local`, and `argocd-projects` + `services`
   (the `ApplicationSet`) both exist
2. confirm the services `ApplicationSet` generated exactly one `Application` (`pings-local`),
   `OutOfSync` and in a project-validation-error state (R3, R4 - the project doesn't exist yet)
3. `argocd app sync argocd-projects` - confirm `pings`' `AppProject` now exists, and `pings-local`
   is no longer in a project-validation-error state (still `OutOfSync`/not auto-syncing) (R2, R3)
4. stop Tilt's `pings` **and** `pings-migrate` resources; confirm `postgres` keeps running and
   `pings`' migrate Job is not a standalone Tilt resource independent of `pings` anymore
5. trigger a manual sync of `pings-local` - confirm the PreSync hook Job runs against `pings`' own
   logical database and reaches `Succeeded` before the `app-platform` resources apply, and that
   the existing `pings` namespace shows no orphaned or conflicting resources afterward (R6, R7,
   R8)
6. confirm the generated `Application`'s rendered manifest carries `HYRULE_REGION: us-west-2`/
   `HYRULE_ENVIRONMENT: local`, and that `deploy/values/pings/values.yaml` itself has neither key
   (R5)
7. edit `deploy/values/pings/values.yaml` on the feature branch, push, manually sync again -
   confirm the change (e.g. a resource limit bump) is reflected in the running Deployment (R7)
8. add a second `services.yaml` entry pointed at a `deploy/values/` directory that doesn't exist
   yet (or a trivial second service) - confirm the services `ApplicationSet` generates its
   `Application` without any new ArgoCD YAML being hand-written (R2, R4)
9. `make argocd-down` - confirm `pings`/`databases` and their resources are untouched, then
   restart Tilt's `pings` resource and confirm normal local dev resumes, including the per-service
   migrate resource re-appearing under Tilt
10. confirm `docs/roadmap.md` carries the new initiative entry (R9)

---

## Resolved Questions

| question | resolution |
|---|---|
| does Tilt's `migrate` resource need stopping too, not just `pings`? | yes, still - discovered live: per-service didn't change this from the first draft's global `migrate` resource. `resource_deps` only orders builds, it doesn't propagate enable/disable state, and Tilt's rendered migrate Job is the same object (name + namespace) the PreSync hook owns - `<slug>-migrate` needs stopping explicitly alongside `<slug>` |
| who manages the per-service migrations ConfigMap once a service is Argo-managed? | nobody, this pass - same accepted limitation as the first draft, now scoped per-service-namespace instead of `databases` only |
| does the hook Job move into the `pings` namespace? | yes, now - it's a template inside `pings`' own `app-platform` release, running in `pings`' own namespace, connecting to postgres over the network the same way the app itself does. this reverses the first draft's `databases`-namespace placement, since migrate is no longer a second, explicit-namespace-only Application source (see 0025) |
| how does the `clusters` generator match anything locally, given ArgoCD's in-cluster destination has no labels by default? | unchanged - `cluster-local.yaml`, applied directly by `make argocd-up`, now also carries `environment: local` |
| why `region: us-west-2` for the local cluster and not a new `local` value? | unchanged - proves the generator's selector is the literal one a real CP registration will carry later |
| how do 60 services avoid 60 hand-written `AppProject`/`ApplicationSet` pairs? | a registry (`services.yaml`) drives both: one Application ranges over it to generate every `AppProject`, and one `matrix`-generator `ApplicationSet` crosses it with the `clusters` generator to generate every service's `Application`. registering a service is a data change, not a manifest-authoring exercise |
| why not key `AppProject`s by team/domain instead of service? | team ownership changes far more often than a service does - keying the stable object (the project) to the volatile fact (team) means every reorg touches ArgoCD structure. `team`/`domain` stay as labels on the per-service project instead, ready to drive an RBAC-policy mapping generated from the same registry later (not built this pass) |
| why generate `AppProject`s via one Application instead of one generated Application per service? | same end result (one `AppProject` per service); one Application avoids a bootstrap-ordering problem the per-service-Application option would introduce (each of those Applications would itself need a project before its own `AppProject` exists) |
| why `helm.parameters` for region/environment instead of 0009's committed overlay-file shape? | narrows 0009 per [0026](../../decisions/0026-region-environment-argo-derived.md) - which cluster an Application lands on is already the `clusters` generator's own fact; a second, matched-and-selected file saying the same thing is one truth expressed twice, and today's `pings` values already prove it can drift |
| why does migrate move into `deploy/helm/app` instead of staying a second Application source under `local/argocd/`? | narrows 0017 per [0025](../../decisions/0025-migrate-as-chart-primitive.md) - any of ~60 services needs to opt into its own migrate step without a bespoke ArgoCD-side file each time; also lets Tilt render the same definition instead of maintaining a duplicate |
| does migrate reuse 0017's still-unbuilt `app.pod.container`/job-mode split? | no - self-contained pod spec, per 0025. that split stays tracked in `docs/roadmap.md#devx`, unaffected and not a prerequisite |
| argo-cd chart version? | `10.8.1` (`appVersion v3.5.2`), unchanged from the first draft |
| why `datastores` instead of a `migrate.enabled` flag? | a service engineer opts into schema management without needing to understand migrate's mechanics - declaring the datastore itself is the opt-in. also generalizes past postgres (`flavor: key-value`, etc.) for later without a values-schema rewrite; only `flavor: postgres` does anything this pass |
| where do postgres connection facts live, if not in a service's own values? | a single chart default, `deploy/helm/app/values.yaml`'s `postgres:` block - one file a platform engineer edits to repoint every opted-in service, same mechanism 0026 already uses for region/environment. a service supplies no DSN, ever; the migrate Job builds one internally from the chart default + the datastore entry's own `slug` |
| are the postgres owner credentials in that chart default a real secret? | no, plaintext - same risk profile as today's already-committed `hyrule_owner`/`hyrule_owner`, just centralized to one file instead of duplicated per service. [0011](../../decisions/0011-secrets-generalized.md) is where real secret management gets solved, not this pass. TLS on the connection itself (including locally) is a related direction, not decided or built here - see `docs/roadmap.md` |
| what if a service needs more than one `flavor: postgres` datastore? | not handled this pass - `migrate.yaml`'s Job/ConfigMap naming (`{{ app.fullname }}-migrate`/`-migrations`) doesn't weave in `slug`, so a second entry would collide. every real service today has exactly one; revisit by adding `slug` to both names if a second one actually shows up |
| can `services.yaml` be fed to the `appprojects` chart as `helm.valueFiles`, as first drafted? | no - discovered live during execution, not caught by `helm template` alone. Helm's values loader requires a values file's root to be a map; the services `ApplicationSet`'s git generator requires the same physical file to be a top-level array to generate one param set per entry (R4). the two requirements are incompatible for one file passed via `-f`. fixed by moving `services.yaml` inside the `appprojects` chart directory and reading it via `.Files.Get "services.yaml" | fromYamlArray` instead - one source of truth, in the array shape the generator needs, with no external values injection required |
| does adoption (R8) work without pinning the Helm release name? | no - discovered live on the actual adoption sync, not something `helm template` alone surfaces (it needs two different release names rendered and diffed to show up). ArgoCD defaults a Helm source's `.Release.Name` to the Application's own name; naming the Application `{{service}}-{{name}}` (`pings-local`) means the rendered `app.selectorLabels` no longer match what Tilt's own `helm_resource` release (named plainly `pings`) already set on the live Deployment, and `spec.selector` is immutable - the sync fails patching it. `helm.releaseName: '{{service}}'` on the ApplicationSet template pins it to match |
| does `argocd-down` tear down cleanly without touching `pings`/`databases`? | not by default - discovered live running the actual teardown, not a sync-time concern. the ApplicationSet controller auto-adds `resources-finalizer.argocd.argoproj.io` to every Application it generates; `helm uninstall`'s pod deletion races ahead of `kubectl delete namespace argocd`, so nothing is left to process that finalizer and the delete hangs indefinitely instead of completing (or, if the controller had still been running, it would have cascaded into deleting `pings`' own resources - the opposite of what R8/the PRD require). the ApplicationSet's own `syncPolicy.preserveResourcesOnDeletion: true` stops it adding that finalizer at all |

---

## Requirements Coverage

| requirement | covered by |
|---|---|
| R1 — ArgoCD installed via its own Helm chart, `make argocd-up`/`argocd-down` | § Interfaces (`mk/argocd.mk`), § Core (1, 7) |
| R2 — `local/argocd/appprojects/services.yaml` registry drives everything, no hand-written per-service ArgoCD YAML | § Interfaces (`services.yaml`), § Core (6), § Integration Tests (8) |
| R3 — one `AppProject` per registered service, generated not hand-written | § Interfaces (`appprojects/`, `application-projects.yaml`), § Core (1, 2), § Integration Tests (2, 3) |
| R4 — one `ApplicationSet`, `matrix` generator (registry × `clusters`), one Application per service × matched cluster | § Interfaces (`applicationset-services.yaml`), § Core (1), § Integration Tests (2, 8) |
| R5 — region/environment injected from the matched cluster's labels, never committed to values | § Interfaces (`cluster-local.yaml`, `applicationset-services.yaml`'s `helm.parameters`), § Validation, § Integration Tests (6) |
| R6 — migrate is an opt-in `app` chart primitive, PreSync-hook-annotated, unified with Tilt | § Interfaces (`migrate.yaml`, Tiltfile), § Core (3, 4), § Integration Tests (5) |
| R7 — a values change pushed and manually synced reflects in the running deployment; first sync cleanly adopts `pings` | § Core (4, 5), § Integration Tests (5, 7) |
| R8 — Tilt's `pings` stopped, first sync cleanly adopts the existing namespace, no orphaned/conflicting resources | § Core (3, 4), § Validation (adoption correctness), § Integration Tests (4, 5) |
| R9 — `docs/roadmap.md` gets a new numbered initiative entry | § Files Changed, § Integration Tests (10) |
