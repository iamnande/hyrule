# 0002: local orchestration - Helm + Tilt

phase two (per the roadmap) is standing up a real dev loop against a local
cluster, once there's a service worth orchestrating (there is now - pings
has a real endpoint and a real postgres dependency).

**local cluster choice moved to [0003](0003-runtime.md) /
[0004](0004-local-cluster.md)** - originally decided as `kind` here, since
superseded now that homelab's prod runtime is k3s, not vanilla k8s.

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
container) - actively maintained, integrates cleanly with any standard
kubeconfig context (see [0004](0004-local-cluster.md) for which one),
and has a first-class `helm()` Tiltfile function that renders the chart
being built here rather than requiring a separate manifest path for dev
vs. deploy.

## not decided here

whether `homelab` itself runs a GitOps controller (Argo CD/Flux) on top of
these charts - that's homelab's own concern, out of scope for what hyrule
needs to prove locally.
