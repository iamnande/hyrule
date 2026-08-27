# hyrule

a reference implementation - a playbook for solving real problems, built as
working code. see [docs/architecture.md](docs/architecture.md) for the why.

## services

| service | entrypoint | what it does |
|---|---|---|
| `pings` | `cmd/pings` | a lightweight registry of homelab apps/services/hosts - a thing self-reports by pinging |

## running it

```
make bootstrap   # install mise (if missing) and provision pinned tool versions
make doctor      # verify your toolchain matches what's pinned
make build       # compile pings for the local machine
make run         # go run pings locally
make image-build # build the pings container image
make image-run   # run the built image locally
make stack-up    # start the local dependency stack (see stack/compose.yml)
make stack-down  # stop it
make new-service # scaffold a new cp/dp service (see docs/decisions/0006-service-scaffold.md)
make help        # everything else
```

`SERVICE_NAME` selects which service under `cmd/` a target targets, defaults
to `pings`: `make run SERVICE_NAME=other-service`.

## endpoints (pings)

| path | what |
|---|---|
| `GET /discovery` | service name, version, commit, region, environment |
| `GET /startupz` | startup probe |
| `GET /livez` | liveness probe |
| `GET /readyz` | readiness probe |
| `GET /healthz` | dependency diagnostics |
| `POST /pings` | record a ping - registers the name on first call, bumps last-seen after |
| `GET /pings` | list everything registered, with derived state (`up`/`stale`) |

## requirements

- `make bootstrap` handles the rest (installs [mise](https://mise.jdx.dev/), provisions Go at the version pinned in `mise.toml`)
- `docker` or `podman` (either works - `make` picks whichever is on `PATH`, preferring `docker`)

## configuration

no `.env` file, ever - config comes from real environment variables, and
local defaults already cover the common case (see any `envDefault` tag under
`internal/lib/config`). local-only values (like the compose stack's database
credentials) are hardcoded directly where they're used, since they're not
secrets. for anything that actually is one, inject it at invocation time
(e.g. `op run -- make run`) - nothing secret should ever touch disk.

## tests

```
make test-unit         # ./internal/...
make test-integration  # ./tests/...
make test-lint         # golangci-lint
make test-smoke        # build+run the real binary, curl it, stop it
```
