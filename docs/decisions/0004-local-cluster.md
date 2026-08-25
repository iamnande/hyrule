# 0004: local cluster - Rancher Desktop, not kind

supersedes 0002's "local cluster: kind" leg. with prod now targeting k3s
([0003](0003-runtime.md)), fidelity points the other way - `kind` runs
vanilla upstream Kubernetes, not k3s, so it's no longer the closer match.
Rancher Desktop bundles real k3s directly, running in the same VM as its
containerd - not a k3s lookalike, not a wrapper around `kind`.

that VM-sharing also removes a whole failure class. `kind` nodes are
separate containers each running their own nested containerd, so getting
a locally-built image into the cluster needs an explicit transfer step
(`kind load docker-image`). tried this first, on podman (this repo's
container engine): that command is hardcoded to shell out to a binary
literally named `docker`, even with `KIND_EXPERIMENTAL_PROVIDER=podman`
set - a 5-year-old open upstream gap. worked around it (`make cluster-up`
symlinking `docker` -> `podman`), then reconsidered instead of just
living with the workaround: checked k3d (podman support is equally
labeled experimental, no more mature) and Podman Desktop's Kind/Minikube
extensions (same underlying `kind` + podman pairing, just GUI-wrapped,
not a different technology). neither removes the gap.

Rancher Desktop does, structurally: images built with `nerdctl` land
directly in the containerd image store k3s's kubelet already reads from.
no load step, no shim, nothing to fall out of sync.

## mechanics

- container engine mode: **containerd**, not dockerd/moby - matches k3s's
  own embedded runtime, and it's Rancher Desktop's first-class mode
  (dockerd support exists for teams that need a `docker` CLI/socket, but
  containerd is the one that gets the shared-image-store property above).
- [Tiltfile](../../Tiltfile) uses the `ext://nerdctl` extension's
  `nerdctl_build` in place of `docker_build` - same API shape, different
  backend. this is Tilt's own documented path for Rancher Desktop in
  containerd mode, not a workaround.
- `make cluster-up`/`cluster-down` wrap `rdctl start`/`rdctl shutdown`;
  `make dev` runs `tilt up` with `~/.rd/bin` (nerdctl, Rancher Desktop's
  own `rdctl`) appended to `PATH`.
- kubeconfig context is `rancher-desktop`, created by the app itself - no
  cluster-config file to check in, unlike `kind`'s.
