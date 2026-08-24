FROM golang:1.27-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=arm64

ARG ORG_NAME
ARG SERVICE_NAME
ARG PROJECT_COMMIT
ARG PROJECT_VERSION
ARG PROJECT_REPO_URL
ARG GO_VERSION_PACKAGE

WORKDIR /${ORG_NAME}-${SERVICE_NAME}-${PROJECT_VERSION}-${PROJECT_COMMIT}

COPY go.mod go.mod
COPY go.sum go.sum

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -X ${GO_VERSION_PACKAGE}.ServiceCommit=${PROJECT_COMMIT} -X ${GO_VERSION_PACKAGE}.ServiceVersion=${PROJECT_VERSION} -X ${GO_VERSION_PACKAGE}.RepositoryURL=https://${PROJECT_REPO_URL}" \
    -o /bin/${SERVICE_NAME} ./cmd/${SERVICE_NAME}

FROM gcr.io/distroless/base-debian12@sha256:94d5278016c07e4e328da0a983823b92a2341df326d59de3408b8c6fd50bb4b3

USER nonroot
COPY --from=build --chown=nonroot:nonroot /bin/${SERVICE_NAME} /service

ARG SERVICE_NAME
ARG PROJECT_COMMIT
ARG PROJECT_VERSION
ARG PROJECT_REPO_URL
ARG BUILD_DATETIME

ENV PROJECT_COMMIT=${PROJECT_COMMIT}
ENV PROJECT_VERSION=${PROJECT_VERSION}
ENV BUILD_DATETIME=${BUILD_DATETIME}

LABEL orgs.opencontainers.image.created="${BUILD_DATETIME}" \
      orgs.opencontainers.image.authors="Nick Anderson <nick@morethq.com>" \
      orgs.opencontainers.image.source="https://${PROJECT_REPO_URL}" \
      orgs.opencontainers.image.version="${PROJECT_VERSION}" \
      orgs.opencontainers.image.revision="${PROJECT_COMMIT}" \
      orgs.opencontainers.image.title="${SERVICE_NAME} platform" \
      orgs.opencontainers.image.description="${SERVICE_NAME} platform"

ENTRYPOINT [ "/service" ]
