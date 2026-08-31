HELM_UNITTEST_REF := 33c48cac798e465deda9a66c8e6c07c0973cf53d # v1.1.2

.PHONY: bootstrap
bootstrap: ## setup: install mise (if missing) and provision pinned tool versions
	@command -v mise >/dev/null 2>&1 || curl -fsSL https://mise.run | sh
	@mise install
	@helm plugin list 2>/dev/null | grep -q '^unittest\b' \
		|| helm plugin install https://github.com/helm-unittest/helm-unittest \
			--version $(HELM_UNITTEST_REF) --verify=false
	@echo $(PROJECT_LOG_FMT) "toolchain ready - if this is a new shell, run: eval \"\$$(mise activate <your-shell>)\""

.PHONY: doctor
doctor: ## setup: verify the active toolchain matches what's pinned in mise.toml
	@mise install
	@mise ls --current
