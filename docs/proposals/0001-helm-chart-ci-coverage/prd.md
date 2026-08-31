# PRD - Helm chart CI coverage

## Problem

nothing about `deploy/helm` (`app`, `platform`, `app-platform`) is verified
by CI today - only `integration`/`smoke` run, and both only need a
standalone Postgres (`make db-up`), never touching a chart. a broken
template, a values field that's silently unset, or a rendered manifest
missing an expected label/selector only surfaces at `tilt up` time locally
- or not at all, if nobody happens to render that exact path before
merging.

this is [roadmap initiative 1](../../roadmap.md#1-ci-coverage-for-the-k3stilthelm-path),
called out as the blocker everything downstream builds on: initiative 3
(first real homelab deploy of `pings`) is about to trust this path in
production, and initiative 5 rewrites the Tiltfile's build/deploy wiring
directly - exactly the code nothing verifies today.

## Solution

1. `helm lint` + `helm template` for all three charts (`app`, `platform`,
   `app-platform`), rendered against their real values files, as a new CI
   job.
2. add `required` guards to values that must be set - today a missing
   value renders silently (empty string, `null`, or a broken selector)
   instead of failing loudly.
3. `helm-unittest` suites per chart: successful rendering with valid
   values, failure when a required value is missing, and structural
   assertions on rendered output (labels, Service selector matches pod
   labels, probe paths present).
4. wire the job into [.github/workflows/ci.yml](../../../.github/workflows/ci.yml)
   as its own independent job - no `needs:` chaining, same pattern as the
   existing four.

## Configuration

no new user-facing configuration. `helm-unittest` test files
(`deploy/helm/<chart>/tests/*_test.yaml`) are the only new inputs, and
they're test fixtures, not runtime config.

## States

N/A — this is a build-time check (CI job), not a stateful component.

## Behavior

**before:** a broken chart template or an unset required value is
invisible until someone runs `tilt up`/`helm install` against it locally,
or isn't caught at all.

**after:** a PR fails CI if any chart fails to lint, fails to render, has
a required value left unset, or fails a `helm-unittest` structural
assertion. no live cluster involved - `helm lint`/`helm template`/
`helm-unittest` all operate on rendered YAML text.

## API

N/A — no external API surface.

## Lifecycle

N/A — no runtime objects created or destroyed; this is a static check
against files already in the repo.

## Edge Cases

- `platform` is intentionally near-empty ([0005](../../decisions/0005-helm-chart-split.md))
  - it still gets lint/template/`helm-unittest` coverage now, so the
    harness is already in place when it grows real content.
- `app-platform` is a thin wrapper vendoring `app`+`platform` as
  dependencies - it must render against a real `deploy/values/<service>/values.yaml`
  file, not synthetic test values, to mean anything.
- `app`'s `job`/`cronjob` modes aren't built yet (placeholders in the
  values shape only) - coverage is for `mode: deployment` as it exists
  today; extending it is follow-on work when those modes land.
- a live-cluster-only failure class (an RBAC misconfiguration, a
  `helm install` hook ordering issue) stays uncovered - deliberately out
  of scope, see Scope & Non-Goals.

## Observability

none beyond standard CI job pass/fail - no new logging, metrics, or
alerting surface.

## Scope & Non-Goals

**in scope:**
- `helm lint` + `helm template` for `app`, `platform`, `app-platform`
- `required` guards on values that must be set
- `helm-unittest` suites per chart (valid render, missing-required-value
  failure, structural assertions)
- wiring the above into `.github/workflows/ci.yml` as a new job

**out of scope:**
- a live cluster deploy in CI (`kind` or otherwise) - rejected in
  grounding: the target failure modes (invalid templates, missing
  values) don't need a live API server, and `kind` would require forking
  the Tiltfile's `nerdctl_build` step (Rancher-Desktop-specific, doesn't
  load into `kind`'s per-node containerd) for a benefit not asked for
  here. explicit decision: **no `kind` anywhere, including CI.**
- verifying the Tiltfile's own orchestration (image build/injection,
  `trigger_mode`, the `migrate` Job lifecycle) - stays unverified by CI,
  same gap as today, just narrower than initiative 1's original framing
- `values.schema.json` + generated `README.md` (readme-generator-for-helm)
  - separate roadmap devx item, unrelated tooling
- `job`/`cronjob` chart modes - not built yet

## Design Decisions Summary

| decision                  | options considered                                                                                                  | chosen               | rationale                                                                                                                                                                                                                |
|---------------------------|---------------------------------------------------------------------------------------------------------------------|----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| CI verification mechanism | (a) real `tilt ci` against a `kind` cluster in Actions; (b) `helm lint`/`helm template`/`helm-unittest`, no cluster | (b) static rendering | (a) needs a Tiltfile-forked build path to get images into `kind` at all - real added cost not justified by the failure modes in scope (invalid templates, missing values), which (b) already covers directly and cheaply |
| chart scope               | `app` only; all three charts                                                                                        | all three            | `app-platform` is what's actually deployed via `helm_resource` in the Tiltfile; `platform` will grow real content later - cheaper to have the harness in place now than retrofit it                                      |
| live-cluster coverage     | build it now; defer it                                                                                              | defer                | initiative 3 (first real homelab deploy) is next up and exercises this by hand shortly; initiative 5 reworks the Tiltfile's build path for Harbor - anything `kind`-specific built now likely gets thrown away then      |

## Requirements

- **R1:** CI fails if `helm lint` fails for any of `app`, `platform`,
  `app-platform`.
- **R2:** CI fails if `helm template` cannot render any chart against its
  real values file(s).
- **R3:** a required value left unset fails template rendering loudly
  (via Helm's `required` function), not silently - covered by a
  `helm-unittest` case that asserts the failure.
- **R4:** `helm-unittest` suites exist per chart, covering at minimum:
  successful rendering with valid values, and structural assertions
  (labels, Service selector matches pod labels, probe paths present) on
  rendered output.
- **R5:** the check runs as an independent job in
  `.github/workflows/ci.yml`, no `needs:` chaining, consistent with the
  existing four jobs.
