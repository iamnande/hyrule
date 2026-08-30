# hyrule

a reference implementation - a playbook for solving real problems, built as
working code. see [docs/architecture.md](docs/architecture.md) for the why.

## services

| service | entrypoint | what it does |
|---|---|---|
| `pings` | `go/cmd/pings` | a lightweight registry of homelab apps/services/hosts - a thing self-reports by pinging |
| `iam-jwks` | `go/cmd/iam-jwks` | JSON Web Key Set used to verify JWTs |

## running it

```
make bootstrap   # install mise (if missing) and provision pinned tool versions
make doctor      # verify your toolchain matches what's pinned
make build       # compile pings for the local machine
make run         # go run pings locally
make image-build # build the pings container image
make image-run   # run the built image locally
make new-service # scaffold a new service (see docs/decisions/0006-service-scaffold.md)
make help        # everything else
```

`SERVICE_NAME` selects which service under `go/cmd/` a target targets,
defaults to `pings`: `make run SERVICE_NAME=other-service`.

### running against a real local cluster

```
make cluster-up  # start Rancher Desktop (kubernetes + containerd)
make dev         # tilt up - builds and deploys every service under deploy/values/
make cluster-status
make cluster-down
```

see [docs/decisions/0003](docs/decisions/0003-runtime.md),
[0004](docs/decisions/0004-local-cluster.md), and
[0005](docs/decisions/0005-helm-chart-split.md) for why this shape.

## endpoints

every service gets these for free via `runtime.NewModule`:

| path | what |
|---|---|
| `GET /discovery` | service name, version, commit, region, environment |
| `GET /startupz` | startup probe |
| `GET /livez` | liveness probe |
| `GET /readyz` | readiness probe |
| `GET /healthz` | dependency diagnostics |

`pings`:

| path | what |
|---|---|
| `POST /pings` | record a ping - registers the name on first call, bumps last-seen after |
| `GET /pings` | list everything registered, with derived state (`up`/`stale`) |

`iam-jwks`:

| path | what |
|---|---|
| `GET /.well-known/jwks.json` | the current key set - see [docs/decisions/0007](docs/decisions/0007-iam-jwks-key-distribution.md) |

## requirements

- `make bootstrap` handles the rest (installs [mise](https://mise.jdx.dev/), provisions Go/Helm/Tilt at the versions pinned in `mise.toml`)
- `docker` or `podman` for `make db-*` (either works - `make` picks whichever is on `PATH`, preferring `docker`)
- [Rancher Desktop](https://rancherdesktop.io/) (containerd mode) for `make cluster-*`/`make dev` - see [docs/decisions/0004](docs/decisions/0004-local-cluster.md)

## configuration

no `.env` file, ever - config comes from real environment variables, and
local defaults already cover the common case (see any `envDefault` tag under
`go/internal/lib/config`). local-only values (like the standalone test
database's credentials) are hardcoded directly where they're used, since
they're not secrets. for anything that actually is one, inject it at
invocation time (e.g. `op run -- make run`) - nothing secret should ever
touch disk.

## tests

```
make test-unit         # ./go/internal/...
make test-integration  # ./go/tests/...
make test-lint         # golangci-lint
make test-smoke        # build+run the real binary, curl it, stop it
```

`test-integration` and `test-smoke` need a reachable postgres - `make db-up`
starts one standalone (no cluster required), `make db-down` stops it.
