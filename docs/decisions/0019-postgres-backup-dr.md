# 0019: postgres backup & DR - nightly logical dump off-cluster, CP only, PITR deferred

## why this can't be skipped: `local-path-provisioner` isn't redundant storage

[0003](0003-runtime.md)'s `local-path-provisioner` ties a PVC to one
node's local disk - no synchronous replica, no network-attached
redundancy underneath it. that's true in a real CP deployment the same
way it's true locally: homelab hardware, not a cloud disk with its own
durability guarantees. so backup isn't optional insurance on top of an
already-durable storage layer, it's the only durability CP's data has -
if that node's disk goes, whatever wasn't backed up elsewhere is gone,
full stop.

## nightly logical `pg_dump`, off-cluster, CP only

the actual mechanism: a nightly `pg_dump` against CP's primary, shipped
to storage outside the cluster. logical dump, not WAL archiving/PITR -
see below for why that's deferred, not an oversight. CP only - DP is a
selective logical subscriber to a subset of published tables, not a
mirror ([0023](0023-postgres-replication-topology.md)), so it isn't
authoritative for anything and doesn't need its own backup: losing DP
means resubscribing once it's back, not data loss.

this is operational infrastructure, not a service's own workload - the
same bucket [0017](0017-scheduled-workloads.md) puts the existing
`migrate` Job in, not `app`'s `job`/`cronjob` mode (which is for a real
service's own recurring work). it lives alongside
[local/cluster/migrate-job.yaml](../../local/cluster/migrate-job.yaml)
as freestanding infra tooling - a raw `CronJob` manifest today, whatever
[0012](0012-migrations-under-gitops.md)'s ArgoCD story becomes for real
infra manifests once ArgoCD exists.

## WAL archiving / PITR (pgBackRest, wal-g) - deferred

point-in-time recovery buys back the gap between "lost everything since
last night's dump" and "lost everything since the last committed
transaction." that gap only matters if a concrete RPO requirement says a
day of data loss is unacceptable - nothing in this repo has said that
yet, homelab's own product surface isn't even decided
([docs/roadmap.md](../roadmap.md)'s product section). standing up
`pgBackRest`'s repo/stanza machinery (or `wal-g`) against a real RPO
number that doesn't exist yet would be exactly the kind of
build-ahead-of-need this repo consistently avoids
([0007](0007-iam-jwks-key-distribution.md)'s eventing,
[0015](0015-authn-authz.md)'s policy engine,
[0017](0017-scheduled-workloads.md)'s job templates,
[0018](0018-resource-sizing-autoscaling.md)'s autoscaling). revisit
against a real RPO number if one ever shows up, not preemptively.

**"nightly" itself is a prioritization call, not a derived number - worth
being honest about that.** this doc argues PITR isn't justified because
no real RPO exists, then turns around and picks a specific ~24h data-loss
window anyway. that's not a contradiction so much as an admission: a
recovery story has to start somewhere, "nightly" is that starting point,
and it's chosen on the same "good enough until proven otherwise" basis
as the scaffold's resource numbers in
[0018](0018-resource-sizing-autoscaling.md) - not rigorously derived,
not meant to be the last word. recoverability matters enough that this
gets more attention as real requirements surface: finer-grained RPO
needs, retention depth, and a rehearsed restore procedure are real
follow-up decisions to flesh out deliberately, not gaps this doc is
pretending don't exist.

## if CNPG ever gets adopted, it likely supersedes this - not decided here, not blocking on it either

[0023](0023-postgres-replication-topology.md) already leaves "whether a
Postgres-replication-aware operator like CloudNativePG is worth adopting
over hand-rolled SQL" as its own open question. CNPG bundles Barman Cloud
backups as a first-class feature, so adopting it later would likely
replace the plain `pg_dump` CronJob above rather than run alongside it.
this decision isn't waiting on that one: today's postgres is a plain
Deployment running raw SQL, and it needs a backup story that works
against *that* shape now, not one gated on an operator-adoption decision
with no timeline.

## not decided here

- the actual off-cluster storage target (self-hosted MinIO, Backblaze B2,
  an off-site NAS) and its credentials - real infrastructure that
  doesn't exist yet, same "real infra work, not decided further here"
  bucket as Harbor ([0010](0010-image-registry-and-publishing.md)) and
  1Password Connect ([0011](0011-secrets-generalized.md)).
- retention policy (how many nightly dumps to keep) and encryption of
  the dump at rest/in transit.
- the restore procedure itself - not written as a runbook, never
  rehearsed. "restore" today means "spin up postgres, run the migrate
  Job, load the latest dump" in outline only.
- alerting on a failed/missed backup run - depends on
  [0016](0016-observability-backend.md)'s stack existing to alert
  through.
- CP failover more broadly - [0023](0023-postgres-replication-topology.md)
  already owns "what happens if CP's primary is unreachable" as an open
  question; this doc is about recovering lost data, not staying up
  through an outage.
