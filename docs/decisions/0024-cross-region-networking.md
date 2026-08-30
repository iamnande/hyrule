# 0024: cross-region private networking - ngrok, not a separate VPN mesh

CP (`us-west-2`) and DP (`us-east-2`) are two separate clusters entirely
(per [0009](0009-environments-and-promotion.md), a VPS each to start) -
not one cluster spanning regions. narrower scope than it first looked:
[0023](0023-postgres-replication-topology.md) decided DP doesn't accept
writes at all today, so this was never a general write-forwarding path.
two concrete needs, both DP reaching something in CP:

- [0023](0023-postgres-replication-topology.md): DP's logical-replication
  connection to CP's postgres.
- [0010](0010-image-registry-and-publishing.md): DP pulling images from
  CP's Harbor.

## reuse ngrok rather than standing up a second networking technology

[0013](0013-ingress-dns-tls.md) already put ngrok in charge of public
ingress for CP/DP - private/internal endpoints and TCP tunneling cover
both needs above without introducing a separate VPN mesh (WireGuard,
Tailscale/Headscale) as a second thing to run, secure, and reason about
independently. CP's postgres and Harbor get reached from DP through
ngrok's own access-controlled tunnels, not a raw exposed port and not a
site-to-site tunnel maintained outside ngrok entirely.

this also means whatever ngrok Kubernetes Operator work
[0013](0013-ingress-dns-tls.md) already deferred is the same
infrastructure work this decision leans on too - not two separate
rollouts to sequence, one.

## not decided here

- the actual ngrok configuration (which endpoint type, what access
  control - IP allowlisting, mTLS, or an ngrok-native mechanism) for the
  postgres replication connection and the Harbor pull path specifically -
  real infrastructure work, no cluster or ngrok setup to configure it
  against yet.
- whether ngrok's per-connection overhead/latency is acceptable for
  postgres logical replication specifically - worth a real check once
  it's running, not assumed here.
