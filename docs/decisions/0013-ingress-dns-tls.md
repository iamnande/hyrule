# 0013: ingress, DNS, TLS - ngrok for CP/DP, Traefik + step-ca locally

## CP and DP: ngrok manages domains and TLS, not cert-manager

domain routing and TLS cert issuance/rotation for anything CP/DP exposes
publicly goes through ngrok - their own pipeline handles Let's Encrypt
issuance and rotation behind the scenes, so there's no separate
cert-manager install or DNS-01 solver to run or maintain in either
region's cluster. `morethq.com` is already registered (Porkbun); ngrok
manages the domains/records it needs directly rather than this repo
driving Porkbun's API itself.

starting point is whatever's simplest to stand up for real against the
two actual services that exist (`pings`, `iam-jwks`) - the [ngrok
Kubernetes
Operator](https://ngrok.com/docs/k8s/) is the intended end state
(implements the standard `networking.k8s.io/Ingress` API with its own
`IngressClass`, so this fits the same generic Ingress shape described
below without a bespoke CRD), but adopting it is real infrastructure
work sequenced behind there being a cluster to install it into - same
"not decided further here" bucket as Harbor and ESO
([0010](0010-image-registry-and-publishing.md),
[0011](0011-secrets-generalized.md)).

## local dev: k3s's bundled Traefik + step-ca, deliberately split

local dev doesn't touch ngrok, real DNS, or real Let's Encrypt at all -
matches the split called out explicitly when parity was first discussed:
a cert issuer is exactly the kind of environment-specific concern worth
diverging on, and ngrok's whole value here is managing *public* domains,
which a laptop's local cluster has no need to be part of.

- **ingress controller**: Traefik, already bundled with k3s
  ([0003](0003-runtime.md)) - zero extra install, no reason to swap it
  for something else with nothing concrete driving that.
- **TLS**: `cert-manager` + `step-ca` as a self-hosted, local-only ACME
  issuer (step-ca implements its own ACME endpoint, so `cert-manager`
  talks to it exactly like it would talk to Let's Encrypt). this is the
  one place `cert-manager` exists in this repo at all - CP/DP never run
  it, since ngrok absorbs that role there entirely. an intentional
  asymmetry: local dev carries infrastructure (`cert-manager`, `step-ca`)
  that prod doesn't need, the inverse of the usual "prod has more moving
  parts" shape, and worth remembering as such rather than assuming
  parity always means "local mirrors prod one-for-one."

## one generic Ingress template, environment-specific values

`deploy/helm/app` gets one Ingress template (off by default,
`ingress.enabled`), parameterized the same way every other `app` field
already is - `ingress.className`, `ingress.host`, `ingress.tls`. the
difference between local and CP/DP is entirely in which values get set
per environment (`ingressClassName: traefik` + a `step-ca` issuer
reference locally, `ingressClassName: ngrok` in CP/DP), via the
per-region values overlays [0009](0009-environments-and-promotion.md)
already established - not two different templates or a homelab-specific
fork of the chart.

## not decided here

- actually writing the Ingress template, and the values wiring for both
  the local (`step-ca`/`cert-manager`) and CP/DP (`ngrok`) paths - real
  chart work, not done in this doc.
- standing up `step-ca` locally, and the ngrok Kubernetes Operator in
  CP/DP - both real infrastructure work, no cluster running either yet.
- whether ngrok ends up relevant to
  [0024](0024-cross-region-networking.md)'s still-open question - ngrok
  is a public-ingress product, and 0024 is about CP<->DP private
  connectivity (replication, Harbor pulls), a different problem;
  worth a second look once ngrok's actually running, not assumed here.
