# 0022: Helm/YAML style - folds into initiative 1, no separate style doc

resolves the same way [0021](0021-shell-script-style.md) did: not worth
its own doc. checked directly - there's no `helm lint`, no `yamllint`,
no `helm-unittest` wired in anywhere today (`Makefile`, `mk/*.mk`, and
`.github/workflows/ci.yml` all come up empty), confirming
[docs/roadmap.md](../roadmap.md)'s own claim that nothing about
`deploy/helm`/the Tiltfile is verified by CI yet. that's exactly the
territory [docs/roadmap.md](../roadmap.md#1-ci-coverage-for-the-k3stilthelm-path)'s
initiative 1 is about to spike (`helm lint` + `helm template` + rendered-
output assertions vs a real `tilt ci` run against `kind`) - deciding a
hand-written style guide here first would just get overtaken by whatever
that spike lands on.

what already exists is consistent without anyone having written a rule
down for it - two-space indentation (YAML's own syntax leaves nothing
else to decide), `values.yaml` grouped into blank-line-separated blocks
with a comment only where a value's valid options aren't obvious (`#
deployment | statefulset | daemonset | job | cronjob` above `mode:`) -
same "worth continuing, not a new requirement" framing
[conventions.md#shell-scripts](../conventions.md#shell-scripts) used for
`local/new-service.sh`. no repo-specific YAML rule diverges from
anything a linter would already catch.

if a real divergence ever shows up (something `helm lint`/`yamllint`
wouldn't catch but this repo cares about anyway), it goes directly into
[conventions.md](../conventions.md) as a short section, same as
[0021](0021-shell-script-style.md) did for shell - not litigated
speculatively here.
