# 0006: cp/dp service scaffold, discovered by the Tiltfile

`make new-service` (`hack/new-service.sh`) prompts for a slug, a name, a
description, and a type - `cp`, `dp`, or `both`. control-plane and
data-plane are two flavors of the same generic shape, not two chart
topologies: a `cp` service owns state and gets `database.Module` +
`config.LoadDatabase()` wired in plus the `HYRULE_DATABASE_*` block in
its values file; a `dp` service gets neither. `both` runs the generator
twice, producing `<slug>-cp` and `<slug>-dp` as two independent services
sharing a name and description.

## what gets scaffolded, and what doesn't

every new service gets `cmd/<slug>/{main.go,app/module.go}`,
`internal/svc/<slug>/module.go` (an empty `fx.Module` - it boots, serves
`/discovery` and the probe endpoints via the shared runtime module, and
does nothing else yet), `api/<slug>/openapi.yaml` (zero paths), and
`deploy/values/<slug>/values.yaml`. a row goes into README's services
table.

domain logic, API endpoints, and (for `cp`) the repository/migration
layer are not scaffolded - there's no honest starting content to put
there. `internal/svc/pings` is the reference for adding a first
endpoint the same way pings got one: a path in `openapi.yaml`, oapi-codegen
wired into `internal/svc/<slug>/api`, a domain type, and for `cp` a
migration plus a `sql:` entry in `sqlc.yaml`.

## database config moved out of internal/lib/config

`internal/lib/config.PingsModule` used to bundle `LoadDatabase()` in with
every service's baseline config - a `dp` service has no database, so
that coupling doesn't generalize. renamed to `BaseModule` (deployment,
tracing, HTTP server only); anything that wires in `database.Module`
also provides `config.LoadDatabase()` itself, same as `cmd/<slug>/main.go`
does for `cp`.

## the Tiltfile discovers services, it doesn't list them

`deploy/values/<slug>/` is the source of truth for which services exist
locally - the Tiltfile lists that directory and builds/deploys whatever
it finds, wiring in the shared `migrate` dependency only for services
whose values declare `HYRULE_DATABASE_HOST`. `make new-service` never
touches the Tiltfile. the dev postgres (`deploy/local/postgres.yaml`) is
named `hyrule-database`, not `pings-postgres` - it's shared by every
`cp` service, not owned by pings.

each service's `k8s_resource` uses `port_forward(local_port=0, ...)` so
the local port is OS-assigned - more than one service can run locally
at once without a fixed `8000` colliding across all of them. `tilt up`'s
own UI shows the port it picked per service.
