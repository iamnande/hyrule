RD_BIN_DIR := $(HOME)/.rd/bin

.PHONY: cluster-up
cluster-up: ## cluster: start rancher desktop (kubernetes + containerd)
	@echo $(PROJECT_LOG_FMT) "starting rancher desktop"
	@$(RD_BIN_DIR)/rdctl start --container-engine.name=containerd --kubernetes.enabled=true

.PHONY: cluster-down
cluster-down: ## cluster: shut down rancher desktop
	@echo $(PROJECT_LOG_FMT) "shutting down rancher desktop"
	@$(RD_BIN_DIR)/rdctl shutdown

.PHONY: cluster-status
cluster-status: ## cluster: show local cluster + workload status
	@kubectl --context rancher-desktop get nodes,pods -A

.PHONY: dev
dev: ## cluster: run the tilt dev loop against the local cluster
	@PATH="$$PATH:$(RD_BIN_DIR)" tilt up

.PHONY: helm-vendor
helm-vendor: ## cluster: re-vendor app-platform's chart dependencies (run after editing deploy/helm/app or deploy/helm/platform)
	@echo $(PROJECT_LOG_FMT) "vendoring app-platform chart dependencies"
	@helm dependency build deploy/helm/app-platform
