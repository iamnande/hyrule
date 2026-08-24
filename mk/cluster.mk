KIND_CLUSTER_NAME := hyrule
CLUSTER_BIN_DIR    := $(PROJECT_WORKDIR)/.cluster/bin

PODMAN_MACHINE := $(shell podman machine list --format '{{.Name}}' 2>/dev/null | head -1)
PODMAN_SOCKET  := $(shell [ -n "$(PODMAN_MACHINE)" ] && podman machine inspect $(PODMAN_MACHINE) --format '{{.ConnectionInfo.PodmanSocket.Path}}' 2>/dev/null)

ifeq ($(CONTAINER_ENGINE),podman)
	KIND_ENV := KIND_EXPERIMENTAL_PROVIDER=podman
	TILT_ENV := KIND_EXPERIMENTAL_PROVIDER=podman DOCKER_BUILDKIT=0 DOCKER_HOST=unix://$(PODMAN_SOCKET) PATH="$(CLUSTER_BIN_DIR):$$PATH"
else
	KIND_ENV :=
	TILT_ENV :=
endif

.PHONY: cluster-up
cluster-up: ## cluster: create the local kind cluster
	@echo $(PROJECT_LOG_FMT) "creating kind cluster '$(KIND_CLUSTER_NAME)'"
ifeq ($(CONTAINER_ENGINE),podman)
	@mkdir -p $(CLUSTER_BIN_DIR)
	@ln -sf "$$(command -v podman)" $(CLUSTER_BIN_DIR)/docker
endif
	@$(KIND_ENV) kind create cluster --name $(KIND_CLUSTER_NAME) --config deploy/kind/config.yaml

.PHONY: cluster-down
cluster-down: ## cluster: delete the local kind cluster
	@echo $(PROJECT_LOG_FMT) "deleting kind cluster '$(KIND_CLUSTER_NAME)'"
	@$(KIND_ENV) kind delete cluster --name $(KIND_CLUSTER_NAME)

.PHONY: cluster-status
cluster-status: ## cluster: show local kind cluster + workload status
	@kubectl --context kind-$(KIND_CLUSTER_NAME) get nodes,pods -A

.PHONY: dev
dev: ## cluster: run the tilt dev loop against the local kind cluster
	@$(TILT_ENV) tilt up
