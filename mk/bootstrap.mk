.PHONY: bootstrap
bootstrap: ## setup: install mise (if missing) and provision pinned tool versions
	@command -v mise >/dev/null 2>&1 || curl -fsSL https://mise.run | sh
	@mise install
	@echo $(PROJECT_LOG_FMT) "toolchain ready - if this is a new shell, run: eval \"\$$(mise activate <your-shell>)\""

.PHONY: doctor
doctor: ## setup: verify the active toolchain matches what's pinned in mise.toml
	@mise install
	@mise ls --current
