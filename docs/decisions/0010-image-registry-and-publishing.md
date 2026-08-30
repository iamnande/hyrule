# 0010: image registry & publishing - self-hosted Harbor, everywhere, no exceptions

Harbor, self-hosted - not a cloud-managed registry (ECR, GHCR, or
anything else billed and operated by a cloud provider). matches a
standing preference, not just this one choice: as much of this stack as
possible shouldn't depend on a specific cloud provider's proprietary
services, and a registry is exactly the kind of dependency that's easy to
accidentally wire deep into CI/CD if the first answer reached for is
"whatever's built into wherever CI happens to run."

**local dev pushes and pulls through Harbor too - not just prod.** the
standing principle going forward: local and prod stay as close to each
other as possible, and infrastructure only diverges where there's a
concrete, environment-specific reason to (a cert issuer being the clean
example - see [0013](0013-ingress-dns-tls.md)). there isn't one here, so
Harbor isn't one either.

this knowingly trades away part of what
[0004](0004-local-cluster.md) picked Rancher Desktop for: its whole
selling point was `nerdctl_build` landing an image directly in the
containerd store k3s already reads from, no push/pull round-trip needed.
routing local builds through Harbor reintroduces that round-trip on
purpose - parity with prod's actual path is worth more here than the
inner-loop speed 0004 optimized for. 0004's reasoning about *why*
Rancher Desktop over `kind` still holds (the shared containerd runtime,
no separate image-load step problem) - this only changes what happens
after the build, not the build/runtime pairing itself.

## regions

one Harbor instance in the CP region (`us-west-2`, per
[0009](0009-environments-and-promotion.md)), one more for local dev.
DP (`us-east-2`) pulls across the WAN at deploy time; replicating Harbor
to DP is real, revisit if cross-region pull latency or availability
actually becomes a problem for a deploy, not before.

## CI reaches it via a self-hosted runner, not a public endpoint

GitHub Actions' hosted runners are cloud-hosted and can't reach a private
Harbor on the VPS/homelab network directly. two ways to close that gap:
expose Harbor publicly (pulls forward the still-undecided
[0013](0013-ingress-dns-tls.md) ingress/DNS/TLS work as a dependency of
this one, and puts a registry on the open internet), or run a
self-hosted GitHub Actions runner on the same private network Harbor
lives on, so the push never leaves it.

going with the **self-hosted runner**. it costs a machine to keep
patched and running, but it keeps Harbor off the public internet
entirely - no ingress/TLS decision has to land first, and "not dependent
on a cloud provider" extends naturally to "not dependent on exposing
infrastructure to the public internet just to satisfy a cloud-hosted
CI runner." the runner needs outbound access to GitHub (to receive jobs)
and a network path to Harbor and to the build tooling it needs
(`nerdctl`/buildkit, matching [0004](0004-local-cluster.md)'s local build
path) - real infrastructure work, not decided further here.

## tag scheme: git sha, not `latest`

the deployable tag is the short git commit sha - `PROJECT_COMMIT` is
already computed this way in both the [Makefile](../../Makefile) and the
[Tiltfile](../../Tiltfile), and it's what `/discovery` already reports
per service. `VERSION`'s semver value (see
[0020](0020-hyrule-homelab-versioning.md)) is a human-facing release
marker, not what a deployment pins to - a deployment always names an
exact, immutable build, and a sha is that. this applies to local dev
too, once local builds actually push to Harbor - no more `tag: latest`
placeholder anywhere, local included.

## not decided here (real infrastructure work, tracked as follow-up)

- **standing up Harbor for real, in both the local cluster and CP** -
  this is a genuinely sized piece of work on its own, not a Tiltfile
  tweak: Harbor's own chart, its own storage backend, and - the part
  that actually blocks the local push/pull path - local image pulls need
  the container runtime (containerd, via `nerdctl`) to trust Harbor's
  TLS cert or have it explicitly allowed as insecure. worth its own
  pass, not bundled into this decision doc.
- rewiring [Tiltfile](../../Tiltfile)'s `build_service`/`deploy_service`
  to push to and pull from local Harbor instead of relying on the
  shared containerd store directly, and giving every service's chart the
  `imagePullSecrets` it'll need against a real registry (`app`'s
  values already has the field: `global.imagePullSecrets`).
- how Harbor itself gets installed and tracked in this repo at all -
  same "cluster-singleton infrastructure, no established install path
  yet" question as [0011](0011-secrets-generalized.md)'s ESO.
- the self-hosted runner's own provisioning, and the CI job that
  actually builds and pushes images - both blocked on Harbor existing
  for real first.
- Harbor's own auth/RBAC model (robot accounts for CI vs. real users)
  and image retention/GC policy.
