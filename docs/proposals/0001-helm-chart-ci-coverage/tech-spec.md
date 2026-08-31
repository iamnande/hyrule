# Tech Spec — Helm chart CI coverage

## Overview + Design Principles

adds a `helm` job to CI that statically verifies all three charts
(`app`, `platform`, `app-platform`) - `helm lint`, `helm template`
against real values, and `helm-unittest` suites - with no live cluster
involved. a missing required value now fails template rendering loudly
instead of producing a silently broken manifest.

- static-only, no cluster: `kind` stays banned everywhere including CI
  (PRD scope decision) - every check here operates on rendered YAML text
- `app-platform` renders against the real per-service files under
  `deploy/values/*/values.yaml`, never synthetic test values - its own
  `values.yaml` is an intentionally empty placeholder (`app: {}` /
  `platform: {}`) and proves nothing on its own
- one guard point: the `required` check lives in `app`'s `app.image`
  helper and fires identically whether `app` renders standalone or as
  `app-platform`'s aliased dependency - Helm scopes subchart values
  automatically, so no duplicate guard is needed in the wrapper
- new services need zero CI wiring: the job discovers
  `deploy/values/*/values.yaml` the same way the Tiltfile's
  `SERVICE_SLUGS` does (`ls deploy/values`), not a hardcoded list
- independent job, no `needs:` chaining - matches `lint`/`unit`/
  `integration`/`smoke`
- every external reference this job introduces - the two GitHub Actions
  and the `helm-unittest` plugin - is pinned to a commit SHA, not a
  version tag, so a tag can't be moved out from under the job after the
  fact (A08). scoped to what this PRD introduces only: the existing
  four jobs' tag-pinned actions (`actions/checkout@v6`, etc.) are
  untouched - a broader retrofit is out of scope here, tracked as its
  own follow-on if wanted

---

## Data Model

N/A — no data model. the only new on-disk artifacts are `helm-unittest`
fixture files at `deploy/helm/<chart>/tests/*_test.yaml` and a pinned
plugin version.

---

## Interfaces

**Makefile** (`mk/test.mk`) - one new target, same shape as the existing
four:

```makefile
.PHONY: test-helm
test-helm: ## test: lint, template, and unittest all helm charts
	@echo $(PROJECT_LOG_FMT) "testing helm charts"
	@for chart in app platform; do \
		helm lint deploy/helm/$$chart; \
		helm template deploy/helm/$$chart > /dev/null; \
	done
	@for values in deploy/values/*/values.yaml; do \
		helm lint deploy/helm/app-platform -f $$values; \
		helm template deploy/helm/app-platform -f $$values > /dev/null; \
	done
	@for chart in app platform app-platform; do \
		helm unittest deploy/helm/$$chart; \
	done
```

**bootstrap** (`mk/bootstrap.mk`) - `helm-unittest` is a helm plugin, not
a mise-managed binary, so provisioning it is a separate, idempotent
step pinned the same way `mise.toml` pins everything else:

```makefile
HELM_UNITTEST_REF := 33c48cac798e465deda9a66c8e6c07c0973cf53d # v1.1.2

.PHONY: bootstrap
bootstrap: ## setup: install mise (if missing) and provision pinned tool versions
	@command -v mise >/dev/null 2>&1 || curl -fsSL https://mise.run | sh
	@mise install
	@helm plugin list 2>/dev/null | grep -q '^unittest\b' \
		|| helm plugin install https://github.com/helm-unittest/helm-unittest \
			--version $(HELM_UNITTEST_REF)
	@echo $(PROJECT_LOG_FMT) "toolchain ready - if this is a new shell, run: eval \"\$$(mise activate <your-shell>)\""
```

pinned by commit SHA rather than the `v1.1.2` tag - `helm plugin
install --version` accepts any git ref, tag or SHA, and passes it
straight to `git checkout`. bumping the pin later needs
`helm plugin uninstall unittest` first - the idempotency check only
looks for presence, not version, same tradeoff the tag-based check
would have had.

**CI** (`.github/workflows/ci.yml`) - new job, independent, same
skeleton as `lint`/`unit`, actions pinned by SHA:

```yaml
  helm:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803 # v6
      - uses: jdx/mise-action@c2a87611a18de5b3828c5652fe268e992400cb5c # v4
      - run: make bootstrap
      - run: make test-helm
```

---

## Validation

- **template render (`app.image` helper, `_helpers.tpl`):**
  `.Values.image.repository` has no safe default (`""` today) - guarded
  with Helm's `required` function. this is the only value across all
  three charts with no safe default; everything else in `values.yaml`
  either has one or is genuinely optional (`extraEnv`, `additionalContainers`,
  etc.)
- **`helm lint`:** chart metadata and template syntax, Helm's own
  built-in checks - no repo-specific rules layered on top (see
  [0022](../../decisions/0022-helm-yaml-style.md))
- **`helm-unittest` structural assertions:** all four labels present
  (`app.kubernetes.io/name`, `app.kubernetes.io/instance`,
  `app.kubernetes.io/managed-by`, `helm.sh/chart`), Service `selector`
  equals the Deployment's pod template labels, and each enabled probe's
  `path` matches its configured value

---

## Core

**1. required guard** (`deploy/helm/app/templates/_helpers.tpl`,
`app.image`):

```gotemplate
{{- define "app.image" -}}
{{- $registry := "" }}
{{- if (((.Values.global).image).registry) }}
{{- $registry = .Values.global.image.registry }}
{{- else if .Values.image.registry }}
{{- $registry = .Values.image.registry }}
{{- end }}
{{- $repository := required "image.repository is required" .Values.image.repository }}
{{- $tag := .Values.image.tag | default "latest" }}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- else }}
{{- printf "%s:%s" $repository $tag }}
{{- end }}
{{- end -}}
```

**2. lint/template coverage:** `app` and `platform` render against
their own committed `values.yaml` defaults - that already is each
chart's "real" values file, nothing to render against a specific
service for. `app-platform` renders once per file under
`deploy/values/*/values.yaml` (currently `pings`, `iam-jwks`) -
discovered by glob, not hardcoded, so a new service under `deploy/values`
gets covered automatically with no CI change.

**3. `helm-unittest` suites**, one file per chart under
`deploy/helm/<chart>/tests/`:

- `deploy/helm/app/tests/deployment_test.yaml` - renders with default
  values; asserts all four labels, Service selector == pod labels, all
  three probe paths (`/startupz`, `/livez`, `/readyz`); a second case
  unsets `image.repository` and asserts the `required` failure message
- `deploy/helm/platform/tests/additional_test.yaml` - renders empty
  with defaults (no `additionalK8sObjects`); a second case sets one
  object and asserts it's emitted verbatim
- `deploy/helm/app-platform/tests/render_test.yaml` - one case per file
  under `deploy/values/*/values.yaml`, loaded via `helm-unittest`'s
  `values:` key (paths relative to the chart root, e.g.
  `../../../values/pings/values.yaml`); asserts the aliased `app` and
  `platform` subcharts both render and the same label/selector/probe
  assertions hold end to end through the wrapper

---

## Query Layer

N/A — no database involved.

---

## API Layer

N/A — no runtime API surface. the only entry points are `make test-helm`
locally and the `helm` CI job.

---

## Files Changed

| file | change |
|---|---|
| `deploy/helm/app/templates/_helpers.tpl` | add `required` guard to `app.image` for `image.repository` |
| `deploy/helm/app/tests/deployment_test.yaml` | new — valid render + structural assertions + required-value failure case |
| `deploy/helm/platform/tests/additional_test.yaml` | new — empty render + populated `additionalK8sObjects` case |
| `deploy/helm/app-platform/tests/render_test.yaml` | new — one case per real `deploy/values/*/values.yaml` |
| `mk/test.mk` | add `test-helm` target |
| `mk/bootstrap.mk` | add idempotent `helm-unittest` plugin install (pinned by commit SHA) to `bootstrap` |
| `.github/workflows/ci.yml` | add independent `helm` job, actions pinned by commit SHA |

---

## Integration Tests

no new automated integration suite - this change *is* the test
coverage. manual verification before merge:

1. `make bootstrap` on a clean checkout - confirm the `helm-unittest`
   plugin installs (or is already present) without touching mise-managed
   tools
2. `make test-helm` locally - confirm it passes against current chart
   state
3. temporarily blank `image.repository` in `deploy/helm/app/values.yaml`
   - confirm `helm template deploy/helm/app` fails with the `required`
     message, then revert
4. delete a label from `app.selectorLabels` locally - confirm the
   `helm-unittest` structural assertion fails, then revert
5. push the branch, confirm the new `helm` CI job runs independently of
   `lint`/`unit`/`integration`/`smoke` and reflects the same pass/fail
   results as steps 2-4

---

## Resolved Questions

| question | resolution |
|---|---|
| does `app-platform`'s own `values.yaml` get rendered/tested too? | no - it's an intentionally empty placeholder (PRD edge cases); only the real `deploy/values/*` files are meaningful renders |
| how is `helm-unittest` provisioned given it's not a mise-managed binary? | pinned helm plugin, installed idempotently in `bootstrap`/`doctor`, version `1.1.2` |
| does the required guard need duplicating in `app-platform` for the aliased dependency? | no - Helm scopes subchart values automatically, so the single guard in `app`'s `app.image` helper fires the same way through the wrapper |
| how does CI cover a service added after this lands? | it doesn't need to - `app-platform` render/lint/test loops glob `deploy/values/*/values.yaml`, so a new service is covered with zero CI changes |
| tag or SHA for the new job's external references? | commit SHA (with the tag as a trailing comment) for `actions/checkout`, `jdx/mise-action`, and the `helm-unittest` plugin install - scoped to what this PRD introduces; the existing four jobs' tag-pinned actions are untouched, a broader retrofit is a separate follow-on |

---

## Requirements Coverage

| requirement | covered by |
|---|---|
| R1 — `helm lint` fails CI for any of the three charts | § Core (2), § Files Changed (`mk/test.mk`, `ci.yml`) |
| R2 — `helm template` fails CI if any chart can't render against real values | § Core (2) |
| R3 — an unset required value fails loudly, covered by a test case | § Validation, § Core (1, 3) |
| R4 — `helm-unittest` suites per chart, valid render + structural assertions | § Core (3), § Files Changed |
| R5 — independent CI job, no `needs:` chaining | § Interfaces (CI), § Overview design principles |
