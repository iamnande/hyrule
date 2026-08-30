# 0020: hyrule versioning - manual `VERSION` bump + matching git tag; what "homelab" ends up being stays undecided

## correction: hyrule *is* homelab today - this doc's first pass assumed a split that doesn't exist

the original framing here (and the wording in
[architecture.md](../architecture.md)'s "seed for
[homelab](https://github.com/iamnande/homelab)" and
[docs/roadmap.md](../roadmap.md)'s repeated "no visibility into
homelab's current state from inside this repo") treats homelab as an
already-separate consumer repo that copies patterns out of this one.
checked against reality directly with the person who'd know: it isn't.
hyrule *is* the homelab, right now, one codebase. what it becomes later
is genuinely open - two live possibilities, neither chosen:

- homelab ends up being something else entirely (a wiki, a site) while
  hyrule stays the code, unchanged from today's shape.
- hyrule gets wholesale copied or renamed into a `homelab` identity,
  commit history and all - hyrule *becomes* homelab rather than seeding
  a separate thing.

this doc doesn't pick between them - there's no concrete need forcing
that choice yet, same "don't decide ahead of a real need" bar every
other decision this session held itself to
([0007](0007-iam-jwks-key-distribution.md)'s eventing,
[0015](0015-authn-authz.md)'s policy engine,
[0017](0017-scheduled-workloads.md)'s job templates,
[0018](0018-resource-sizing-autoscaling.md)'s autoscaling,
[0019](0019-postgres-backup-dr.md)'s PITR) - the first version of this
doc broke that pattern by designing a cross-repo pinning scheme for a
split that isn't real yet. what's left below is scoped to what's
actually decidable today: how `VERSION` itself moves, regardless of
which future this turns into.

## `VERSION` is bumped by hand, at a deliberate checkpoint - not per-commit, not CI-driven

no automated release pipeline exists or is being built here - the CI gap
on record ([docs/roadmap.md](../roadmap.md)'s initiative 1) is about
verifying deploy correctness, not cutting releases, and inventing a
tag-on-`main`-drives-a-release pipeline with nothing concrete asking for
one would be its own build-ahead-of-need mistake. bump `VERSION` by
hand, in a commit, when the state of the repo is actually worth marking
as a checkpoint. semver here is a loose human marker, not a strict
compatibility contract enforced by tooling - there's no `go.mod require`
line anywhere that could even depend on it (everything reusable lives
under `go/internal/lib/...`/`go/internal/svc/...`, and Go's `internal/`
rule means nothing outside this module could import it even if one
existed) - so there's nothing to be rigorous about beyond "bump minor
for a real pattern change, patch for a small fix, don't sweat major."

## a matching git tag is the actual pointer - `VERSION` alone isn't enough

`VERSION`'s value at any given commit is unambiguous only if you already
know which commit to look at. the durable, shareable reference is a git
tag (`vX.Y.Z`, matching `VERSION` exactly) created on the same commit as
the bump - that's the thing anyone (a future porting effort, a future
"homelab" of either shape above, this doc) can actually point at. one
source of truth (the file), one mechanical consequence of bumping it
(the tag) - not two independently-maintained markers that can drift from
each other.

## `deploy`-time tags stay git-sha, unaffected by this

[0010](0010-image-registry-and-publishing.md) already settled that a
deployable image is tagged by git sha, calling `VERSION` "a human-facing
release marker, not what a deployment pins to" - this doc doesn't
reopen that, just makes the release-marker half of that sentence
concrete.

## not decided here

- which of the two futures above (homelab-as-something-else vs.
  hyrule-becomes-homelab) actually happens, or when.
- if a genuine second, separate codebase ever does consume this one's
  patterns: how it pins to a point in time, whether it tracks `main` or
  a tag, how re-syncing an already-copied pattern later would work - all
  real questions, none of them answerable before there's a real second
  codebase to design against.
- the mechanics of the first real handoff/porting event itself (e.g.
  `pings`) - that's [docs/roadmap.md](../roadmap.md#3-first-real-homelab-deploy-of-pings)'s
  initiative 3, to be documented as it actually happens, not speculated
  on here.
