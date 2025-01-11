FROM golang:1.23-alpine AS build

ARG ORG_NAME
ARG APP_NAME
ARG APP_COMMIT
ARG APP_VERSION
ARG APP_REPO_URL
ARG GO_VERSION_PACKAGE

WORKDIR /${ORG_NAME}-${APP_NAME}-${APP_VERSION}-${APP_COMMIT}

COPY go.mod go.mod
COPY go.sum go.sum

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="-w -s -X ${GO_VERSION_PACKAGE}.ServiceCommit=${APP_COMMIT} -X ${GO_VERSION_PACKAGE}.ServiceVersion=${APP_VERSION}" \
    -o /bin/${APP_NAME} ./cmd/main.go

FROM gcr.io/distroless/base@sha256:726031c89427790433f6a5bb453a0434c5370c65068b0c66631573cbdc64e9a1

USER nonroot
COPY --from=build --chown=nonroot:nonroot /bin/${APP_NAME} /${APP_NAME}

ARG APP_NAME
ARG APP_COMMIT
ARG APP_VERSION
ARG APP_REPO_URL
ARG BUILD_DATETIME

ENV APP_COMMIT=${APP_COMMIT}
ENV APP_VERSION=${APP_VERSION}
ENV BUILD_DATETIME=${BUILD_DATETIME}

LABEL orgs.opencontainers.image.created="${BUILD_DATETIME}" \
      orgs.opencontainers.image.authors="Nick Anderson <engineering@morethq.com>" \
      orgs.opencontainers.image.source="https://${APP_REPO_URL}" \
      orgs.opencontainers.image.version="${APP_VERSION}" \
      orgs.opencontainers.image.revision="${APP_COMMIT}" \
      orgs.opencontainers.image.title="${APP_NAME} platform" \
      orgs.opencontainers.image.description="${APP_NAME} platform"

ENTRYPOINT [ "/${APP_NAME}" ]
