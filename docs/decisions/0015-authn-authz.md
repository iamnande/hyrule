# 0015: authN/authZ model - iam-jwks stays verification-only, minimal claims, defer real authZ

## iam-jwks verifies, it doesn't issue

[0007](0007-iam-jwks-key-distribution.md) already shapes `iam-jwks` as a
pure verification-side service - serves `/.well-known/jwks.json`, reads a
key from config, caches it, nothing else. that doesn't change here. token
*issuance* (login, token exchange, minting a JWT signed with the private
half of the key `iam-jwks` publishes) belongs to a dedicated identity
service that doesn't exist yet -
[conventions.md#url-structure-versioning](../conventions.md#url-structure--versioning)
already frames "an identity service owning account/workspace/user/team"
as the eventual shape, and
[docs/roadmap.md](../roadmap.md) is explicit that proving RLS end to end
is "blocked on an identity service existing." this decision is about the
*model* every service will verify against - the issuer itself is real
infrastructure + a real service to scaffold, not decided further here.

## token format: EdDSA, already decided

signed with Ed25519, `alg: EdDSA` - not a new decision, just inherited
from [0007](0007-iam-jwks-key-distribution.md)'s key format (RFC 8037).

## claims: standard + `account_id`, nothing speculative beyond that

standard registered claims (`iss`, `sub`, `aud`, `exp`, `iat`, `jti`) plus
exactly one custom claim that matters structurally: `account_id`. this
isn't an arbitrary addition - it's the same value
[conventions.md#data-layer](../conventions.md#data-layer)'s RLS-readiness
work already built a slot for: `database.WithTx`'s `setGUCs` step (`SET
LOCAL app.account_id = ...`) is a no-op today specifically because
nothing populates it yet. a verified token's `account_id` claim is what
populates it - the same claim that answers "who is this" for
authorization is what scopes every query, no separate mechanism needed
for the two.

a coarse `roles` (or `scopes`) claim - a flat list of strings - goes in
too, for the minimal authorization step below. no nested permission
objects, no per-resource grants in the token itself.

## authorization: claim-based only, a real policy model isn't decided

past "does this token have role X," a real RBAC/ABAC engine isn't
designed here - there's no authorization requirement concrete enough yet
to design against; homelab's actual product surface is still openly
undecided per [docs/roadmap.md](../roadmap.md)'s product section. this
matches every other "don't build the mechanism before a real need"
decision in this repo ([0007](0007-iam-jwks-key-distribution.md)'s
eventing, [0018](0018-resource-sizing-autoscaling.md)'s autoscaling) -
inventing a policy engine against zero real authorization rules would be
guessing, not deciding.

## enforcement: both the ngrok edge and in-service middleware

ngrok's Traffic Policy validates a token's signature/expiry at the edge
for anything crossing it publicly - a first line of defense, and a cheap
one given [0013](0013-ingress-dns-tls.md) already put ngrok in charge of
that traffic anyway. it doesn't replace in-service verification:
service-to-service calls that never cross the ngrok edge still need it,
and every service needs the claims *in process* regardless of whether
the edge already checked the signature - `account_id` has to reach
`database.WithTx`'s GUC step, which happens deep inside a service, not at
the edge.

in-service verification is shared middleware
(`go/internal/lib/rest/middleware`, alongside `access_log.go`,
`context.go`, `panic_recovery.go`, `tracing.go` - a new file joins that
set) - verifies the signature against `iam-jwks`'s published keys (cached
locally, refreshed on an interval - polling `iam-jwks` at request time
would be exactly the kind of synchronous cross-service dependency
[0007](0007-iam-jwks-key-distribution.md) already avoided for key
distribution itself), and puts the verified claims on the request
context, matching the existing pattern in
[go/internal/lib/logging/context.go](../../go/internal/lib/logging/context.go)
and
[go/internal/lib/tracing/context.go](../../go/internal/lib/tracing/context.go).

## not decided here

- the identity service itself - scaffolding, login flow, token
  issuance/refresh, session model.
- a real authorization policy engine, once a concrete need for one shows
  up.
- token revocation - a compromised or logged-out token is valid until
  `exp` today, by construction; no revocation list or short-lived-token
  + refresh story decided.
- the actual middleware implementation and its key-refresh mechanism -
  real code, not written in this doc.
