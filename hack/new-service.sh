#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

ORG_NAME="iamnande"

read -rp "slug (lowercase, hyphens ok, e.g. widgets): " SLUG
read -rp "name (human-readable, e.g. Widgets): " NAME
read -rp "description (one line): " DESCRIPTION
read -rp "type [cp/dp/both]: " TYPE
TYPE="$(echo "$TYPE" | tr '[:upper:]' '[:lower:]')"

if [[ ! "$SLUG" =~ ^[a-z][a-z0-9-]*$ ]]; then
	echo "slug must start with a letter and contain only lowercase letters, digits, hyphens" >&2
	exit 1
fi

case "$TYPE" in
cp) KINDS=(cp) ;;
dp) KINDS=(dp) ;;
both) KINDS=(cp dp) ;;
*)
	echo "type must be cp, dp, or both" >&2
	exit 1
	;;
esac

add_readme_row() {
	local full_slug="$1" entrypoint="$2" what="$3"
	local row="| \`${full_slug}\` | \`${entrypoint}\` | ${what} |"
	awk -v newrow="$row" '
		/^## services/ { in_table = 1 }
		/^## / && !/^## services/ { in_table = 0 }
		in_table && /^\| `/ { last = NR }
		{ lines[NR] = $0 }
		END {
			for (i = 1; i <= NR; i++) {
				print lines[i]
				if (i == last) print newrow
			}
		}
	' README.md >README.md.tmp && mv README.md.tmp README.md
}

scaffold_service() {
	local kind="$1"
	local full_slug="${SLUG}-${kind}"
	local pkg
	pkg="$(echo "$full_slug" | tr -d '-')"
	local role
	if [[ "$kind" == "cp" ]]; then role="control plane"; else role="data plane"; fi

	if [[ -d "cmd/${full_slug}" || -d "internal/svc/${full_slug}" ]]; then
		echo "already exists: ${full_slug}" >&2
		exit 1
	fi

	mkdir -p "cmd/${full_slug}/app"
	mkdir -p "internal/svc/${full_slug}"
	mkdir -p "api/${full_slug}"
	mkdir -p "deploy/values/${full_slug}"

	cat >"internal/svc/${full_slug}/module.go" <<EOF
package ${pkg}

import "go.uber.org/fx"

var Module = fx.Module("${full_slug}")
EOF

	if [[ "$kind" == "cp" ]]; then
		cat >"cmd/${full_slug}/app/module.go" <<EOF
package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/database"
	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/internal/svc/${full_slug}"
)

const Name = "${full_slug}"

var Module = fx.Module(Name,
	fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	}),
	database.Module,
	svc.Module,
)
EOF
		cat >"cmd/${full_slug}/main.go" <<EOF
package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/cmd/${full_slug}/app"
	"github.com/iamnande/hyrule/internal/lib/config"
	"github.com/iamnande/hyrule/internal/lib/runtime"
)

func main() {
	fx.New(
		runtime.NewModule(app.Name),
		config.BaseModule,
		fx.Provide(config.LoadDatabase()),
		app.Module,
	).Run()
}
EOF
		cat >"deploy/values/${full_slug}/values.yaml" <<EOF
app:
  fullnameOverride: ${full_slug}

  image:
    repository: ${ORG_NAME}/${full_slug}
    tag: latest

  service:
    port: 8000

  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      memory: 128Mi

  env:
    HYRULE_HTTP_SERVER_ADDR: ":8000"
    HYRULE_REGION: us-east-2
    HYRULE_ENVIRONMENT: dev
    HYRULE_DATABASE_HOST: hyrule-database
    HYRULE_DATABASE_PORT: "5432"
    HYRULE_DATABASE_USER: hyrule_app
    HYRULE_DATABASE_PASSWORD: hyrule_app
    HYRULE_DATABASE_NAME: hyrule
    HYRULE_DATABASE_SSL_MODE: disable

platform: {}
EOF
	else
		cat >"cmd/${full_slug}/app/module.go" <<EOF
package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/internal/svc/${full_slug}"
)

const Name = "${full_slug}"

var Module = fx.Module(Name,
	fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	}),
	svc.Module,
)
EOF
		cat >"cmd/${full_slug}/main.go" <<EOF
package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/cmd/${full_slug}/app"
	"github.com/iamnande/hyrule/internal/lib/config"
	"github.com/iamnande/hyrule/internal/lib/runtime"
)

func main() {
	fx.New(
		runtime.NewModule(app.Name),
		config.BaseModule,
		app.Module,
	).Run()
}
EOF
		cat >"deploy/values/${full_slug}/values.yaml" <<EOF
app:
  fullnameOverride: ${full_slug}

  image:
    repository: ${ORG_NAME}/${full_slug}
    tag: latest

  service:
    port: 8000

  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      memory: 128Mi

  env:
    HYRULE_HTTP_SERVER_ADDR: ":8000"
    HYRULE_REGION: us-east-2
    HYRULE_ENVIRONMENT: dev

platform: {}
EOF
	fi

	cat >"api/${full_slug}/openapi.yaml" <<EOF
openapi: 3.0.3
info:
  title: ${NAME} API
  description: "${DESCRIPTION} (${role})"
  version: "$(date -u +%Y-%m-%d)"
paths: {}
EOF

	add_readme_row "${full_slug}" "cmd/${full_slug}" "${DESCRIPTION} (${role})"

	echo "scaffolded ${full_slug}"
}

for kind in "${KINDS[@]}"; do
	scaffold_service "$kind"
done

cat <<'EOF'

next, for each new service - follow internal/svc/pings as the reference:
  - internal/svc/<slug>/domain: first type + Service
  - api/<slug>/openapi.yaml: first path, then internal/svc/<slug>/api
    (codegen.yaml + oapi-codegen), wired into cmd/<slug>/app/module.go
  - cp only: migrations/, a new sql: entry in sqlc.yaml,
    internal/svc/<slug>/repository
EOF
