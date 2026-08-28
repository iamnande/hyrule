# 0006: service scaffold, discovered by the Tiltfile

`make new-service` (`hack/new-service.sh`) prompts for a slug, a name, a
description, and whether the service owns a database. that's the only
structural fork: a database-backed service gets `database.Module` +
`config.LoadDatabase()` wired in plus the `HYRULE_DATABASE_*` block in
its values file; one that doesn't gets neither.

control-plane and data-plane are not generator concepts. a service that
genuinely splits into two different pieces of software - a policy
authoring service and a policy-enforcement sidecar with actually
different code - is two ordinary `make new-service` runs with their own
real names and descriptions, not a generated pair. most services don't
split at all: `iam-jwks` was the test case for this - loading a signing
key from storage, caching it, serving reads is identical code wherever
it runs, so it's one service, not a control-plane/data-plane pair. "runs
in more than one place" is a deployment concern (a second Helm release,
later, when there's a second place to run it), not a reason to fork the
codebase.

## what gets scaffolded, and what doesn't

every new service gets `cmd/<slug>/{main.go,app/module.go}`,
`internal/svc/<slug>/module.go` (an empty `fx.Module` - it boots, serves
`/discovery` and the probe endpoints via the shared runtime module, and
does nothing else yet), `api/<slug>/openapi.yaml` (zero paths), and
`deploy/values/<slug>/values.yaml`. a row goes into README's services
table.

domain logic, API endpoints, and (for a database-backed service) the
repository/migration layer are not scaffolded - there's no honest
starting content to put there. `internal/svc/pings` is the reference for
adding a first endpoint the same way pings got one: a path in
`openapi.yaml`, oapi-codegen wired into `internal/svc/<slug>/api`, a
domain type, and (if it owns a database) a migration plus a `sql:` entry
in `sqlc.yaml`.

## database config moved out of internal/lib/config

`internal/lib/config.PingsModule` used to bundle `LoadDatabase()` in with
every service's baseline config - a database-less service doesn't need
that coupling. renamed to `BaseModule` (deployment, tracing, HTTP server
only); anything that wires in `database.Module` also provides
`config.LoadDatabase()` itself, same as `cmd/<slug>/main.go` does for a
database-backed service.

## the Tiltfile discovers services, it doesn't list them

`deploy/values/<slug>/` is the source of truth for which services exist
locally - the Tiltfile lists that directory and builds/deploys whatever
it finds, wiring in the shared `migrate` dependency only for services
whose values declare `HYRULE_DATABASE_HOST`. `make new-service` never
touches the Tiltfile. the dev postgres (`deploy/local/postgres.yaml`) is
named `hyrule-database`, not after any one service - every
database-backed service shares it.

each service's `k8s_resource` uses `port_forward(local_port=0, ...)` so
the local port is OS-assigned - more than one service can run locally
at once without a fixed `8000` colliding across all of them. `tilt up`'s
own UI shows the port it picked per service.
