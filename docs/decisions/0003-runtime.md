# 0003: prod runtime - k3s, not vanilla Kubernetes

homelab runs on homelab-scale hardware, not a fleet of cloud nodes. a full
vanilla control plane (etcd, multiple heavier system pods, no opinions on
storage/ingress/load-balancing out of the box) is built for a scale and an
uptime bar this doesn't have. k3s is still a CNCF-certified, fully
conformant Kubernetes distribution - same API, same manifests, same
`kubectl`/Helm - just single-binary, lower footprint, and batteries
included (embedded containerd, `local-path-provisioner`, a lightweight
`ServiceLB`, Traefik ingress by default). nothing about hyrule's shape
(one `internal/svc/<name>` + one Helm chart per service) depends on which
distribution runs it.

this changes what "fidelity to prod" means for local tooling - see
[0004](0004-local-cluster.md).
