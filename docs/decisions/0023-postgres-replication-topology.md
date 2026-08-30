# 0023: postgres replication topology - logical replication, CP publishes, DP subscribes

CP (`us-west-2`) is the authoritative primary for every write, always -
DP (`us-east-2`) is not a full mirror, it's a selective logical
subscriber to whichever published tables actually benefit from
region-local reads. this is a deliberately different shape than a
byte-for-byte streaming replica: logical replication ships row-level
changes for specific tables in a `PUBLICATION`, not the whole cluster,
and - the part that reshapes 0012 - it does **not** carry DDL. a
subscriber needs a matching table structure of its own before it can
subscribe to changes for it.

## which tables get published: opt-in, not automatic

a table joins CP's publication when a real service in DP actually needs
region-local reads for it - not by default, and not "everything,
just in case." matches the repo's standing pattern (0007's eventing,
0018's autoscaling): the mechanism gets built when a concrete need
exercises it, not ahead of one. `pings` and `iam-jwks` have no such need
today - neither publishes anything yet.

## this settles 0012: migrations run independently in both regions

since DDL doesn't replicate, DP's postgres needs the same schema CP has
for any table it might ever subscribe to - which in practice means the
same migrate Job runs in both regions, against each region's own
postgres, using the identical `migrations/` directory content. not one
shared database migrated once; two independently-migrated databases that
happen to share a schema by construction. [0009](0009-environments-and-promotion.md)'s
lockstep CP/DP promotion already assumes this: one promotion event, both
regions move together, so both regions' migrate Jobs run as part of the
same sync.

**DDL ordering across replication is a known sharp edge, not solved
here.** adding a column to a published table means both regions need
that column before either side's application code depends on it - safe
under lockstep promotion's normal case (both migrate Jobs run as part of
the same sync, ahead of the new application code), but a genuine
ordering guarantee between "CP's migration landed" and "DP's migration
landed" isn't built. no protocol added for this preemptively - revisit
if a real incident actually surfaces the gap, same bar the rest of this
repo holds new mechanism to.

## publication/subscription setup isn't a migration

`CREATE PUBLICATION` (on CP) and `CREATE SUBSCRIPTION` (on DP,
referencing CP's connection info and publication name) are inherently
asymmetric - CP and DP would need to run *different* SQL, which breaks
the assumption every migration in [migrations](../../migrations) applies
identically everywhere. rather than special-case `golang-migrate` for
one region-aware exception, replication wiring stays separate,
region-aware operational tooling - the same "real infrastructure work,
not decided further here" bucket as Harbor and ESO
([0010](0010-image-registry-and-publishing.md),
[0011](0011-secrets-generalized.md)). `migrations/` stays exactly what it
is: table/schema shape, identical everywhere.

## DP doesn't accept writes - not yet, and not as a stopgap either

this isn't just a consequence of the replication shape, it's a decision
in its own right: no DP-resident service gets write access to anything,
full stop. no `hyrule_app` (write) role in DP, no write path to CP's
primary provisioned for DP services to fall back on - a service that
needs to write simply doesn't run in DP today. this narrows
[0024](0024-cross-region-networking.md) considerably: the only
cross-region traffic that actually needs a network path right now is
DP's logical-replication connection to CP's publisher and DP pulling
images from CP's Harbor ([0010](0010-image-registry-and-publishing.md)) -
not a general write-forwarding path, which was never a real requirement
to begin with.

## not decided here

- replication slot monitoring - a logical slot lives on CP (the
  publisher), one per subscription; a disconnected or lagging DP
  subscriber makes CP retain WAL indefinitely, a real disk-exhaustion
  risk with no alerting today. depends on
  [0016](0016-observability-backend.md) landing first.
- how publication/subscription bootstrap actually gets scripted or
  operator-managed, and whether a Postgres-replication-aware Kubernetes
  operator (e.g. CloudNativePG) is worth adopting over hand-rolled SQL -
  real infrastructure work, no cluster running this yet to decide it
  against.
- failover: what happens if CP's primary is actually unreachable. DP
  can't take over writes today by design (above), so this stays an open
  question, not a gap accidentally left by the replication shape.
- if a DP write path is ever actually needed later (not decided to
  happen, just not ruled out) - it isn't going to be "DP writes to its
  own postgres and hopes," it'd need real design. not speculated on
  further here.
