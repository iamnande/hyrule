# 0002: local orchestration - kind + Helm + Tilt

phase two (per the roadmap) is standing up a real dev loop against a local
cluster, once there's a service worth orchestrating (there is now - pings
has a real endpoint and a real postgres dependency). three separate
choices bundled under one decision since they only make sense together.

## local cluster: kind, not k3d/minikube/Docker Desktop

`kind` runs actual upstream Kubernetes in Docker; `k3d` runs `k3s`
(Rancher's stripped-down distribution) in Docker - faster and lighter, but
not the same Kubernetes underneath (different defaults for storage,
networking, some APIs). hyrule's whole point is being the reference
implementation homelab's real services copy - fidelity to what a real
cluster actually does matters more here than the speed difference, and
`kind` is the vanilla, first-party (k8s SIG), most-documented option.
minikube and Docker Desktop's built-in cluster are heavier and less
scriptable for CI-adjacent use than either.

## chart format: Helm, not Kustomize

both are legitimate; the honest tradeoff is templating (Helm) vs. patching
(Kustomize). Kustomize wins for a small, mostly-static set of manifests
where environments only differ by replica count/resources/namespace -
adding a templating engine for that is over-engineering. that's not this
repo's shape: hyrule is a *pattern* that homelab's services repeat one
service at a time (`homelab/services/<name>`), each needing its own
parameterized instance of the same shape (image, tag, resources, probes,
env). that's exactly what Helm's templating is for, and most third-party
infra (databases, monitoring agents, anything homelab pulls in that isn't
first-party) already ships as a Helm chart - staying in one templating
ecosystem instead of mixing patch-based overlays with vendor Helm charts
is one less thing to context-switch on.

## dev loop: Tilt

watches source, rebuilds the image, redeploys (or live-updates a running
container) - actively maintained, integrates with `kind` cleanly, and has
a first-class `helm()` Tiltfile function that renders the chart being
built here rather than requiring a separate manifest path for dev vs.
deploy.

## not decided here

whether `homelab` itself runs a GitOps controller (Argo CD/Flux) on top of
these charts - that's homelab's own concern, out of scope for what hyrule
needs to prove locally.

## addendum: kind's podman provider

implementing this surfaced a real, 5-year-old open upstream gap: `kind
load docker-image` (what Tilt uses to push locally-built images into the
cluster) is hardcoded to shell out to a binary literally named `docker`,
even with `KIND_EXPERIMENTAL_PROVIDER=podman` set. this repo's container
engine is podman (no `docker` binary installed), so this bit immediately.

reconsidered the choice rather than just working around it - checked
k3d (podman support is equally labeled experimental, not more mature)
and Podman Desktop's Kind/Minikube extensions (same underlying `kind` +
podman pairing, just GUI-wrapped, not a different technology; podman's
own `kube play` isn't a real cluster - no load balancing, PVs, or network
policies). none of those alternatives remove the gap, so `kind` stands.
fix in place: `make cluster-up` symlinks `docker` -> `podman` into a
repo-local `.cluster/bin`, the same workaround kind's own maintainers
point to. see [conventions.md#local-orchestration](../conventions.md#local-orchestration)
for the full mechanics.
