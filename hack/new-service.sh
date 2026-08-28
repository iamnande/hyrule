#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

ORG_NAME="iamnande"

read -rp "slug (lowercase, hyphens ok, e.g. widgets): " SLUG
read -rp "name (human-readable, e.g. Widgets): " NAME
read -rp "description (one line): " DESCRIPTION
read -rp "owns a database? [y/n]: " HAS_DB_ANSWER
HAS_DB_ANSWER="$(echo "$HAS_DB_ANSWER" | tr '[:upper:]' '[:lower:]')"

if [[ ! "$SLUG" =~ ^[a-z][a-z0-9-]*$ ]]; then
	echo "slug must start with a letter and contain only lowercase letters, digits, hyphens" >&2
	exit 1
fi

case "$HAS_DB_ANSWER" in
y | yes) HAS_DB=1 ;;
n | no) HAS_DB=0 ;;
*)
	echo "owns a database? must be y or n" >&2
	exit 1
	;;
esac

if [[ -d "cmd/${SLUG}" || -d "internal/svc/${SLUG}" ]]; then
	echo "already exists: ${SLUG}" >&2
	exit 1
fi

PKG="$(echo "$SLUG" | tr -d '-')"

mkdir -p "cmd/${SLUG}/app"
mkdir -p "internal/svc/${SLUG}"
mkdir -p "api/${SLUG}"
mkdir -p "deploy/values/${SLUG}"

cat >"internal/svc/${SLUG}/module.go" <<EOF
package ${PKG}

import "go.uber.org/fx"

var Module = fx.Module("${SLUG}")
EOF

if [[ "$HAS_DB" -eq 1 ]]; then
	cat >"cmd/${SLUG}/app/module.go" <<EOF
package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/database"
	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/internal/svc/${SLUG}"
)

const Name = "${SLUG}"

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
	cat >"cmd/${SLUG}/main.go" <<EOF
package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/cmd/${SLUG}/app"
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
	cat >"deploy/values/${SLUG}/values.yaml" <<EOF
app:
  fullnameOverride: ${SLUG}

  image:
    repository: ${ORG_NAME}/${SLUG}
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
	cat >"cmd/${SLUG}/app/module.go" <<EOF
package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/internal/svc/${SLUG}"
)

const Name = "${SLUG}"

var Module = fx.Module(Name,
	fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	}),
	svc.Module,
)
EOF
	cat >"cmd/${SLUG}/main.go" <<EOF
package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/cmd/${SLUG}/app"
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
	cat >"deploy/values/${SLUG}/values.yaml" <<EOF
app:
  fullnameOverride: ${SLUG}

  image:
    repository: ${ORG_NAME}/${SLUG}
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

cat >"api/${SLUG}/openapi.yaml" <<EOF
openapi: 3.0.3
info:
  title: ${NAME} API
  description: "${DESCRIPTION}"
  version: "$(date -u +%Y-%m-%d)"
paths: {}
EOF

ROW="| \`${SLUG}\` | \`cmd/${SLUG}\` | ${DESCRIPTION} |"
awk -v newrow="$ROW" '
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

echo "scaffolded ${SLUG}"
cat <<EOF

next - follow internal/svc/pings as the reference:
  - internal/svc/${SLUG}/domain: first type + Service
  - api/${SLUG}/openapi.yaml: first path, then internal/svc/${SLUG}/api
    (codegen.yaml + oapi-codegen), wired into cmd/${SLUG}/app/module.go
EOF
if [[ "$HAS_DB" -eq 1 ]]; then
	cat <<EOF
  - migrations/, a new sql: entry in sqlc.yaml, internal/svc/${SLUG}/repository
EOF
fi
