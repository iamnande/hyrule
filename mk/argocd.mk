ARGOCD_NAMESPACE := argocd
ARGOCD_CHART_REF := 10.8.1 # argo-cd, appVersion v3.5.2

.PHONY: argocd-up
argocd-up: ## argocd: install argocd + bootstrap the registry-driven project/service generators
	@echo $(PROJECT_LOG_FMT) "installing argocd"
	@helm repo add argo https://argoproj.github.io/argo-helm --force-update
	@helm upgrade --install argocd argo/argo-cd \
		--version $(ARGOCD_CHART_REF) \
		--namespace $(ARGOCD_NAMESPACE) --create-namespace
	@kubectl apply -f local/argocd/cluster-local.yaml
	@kubectl apply -f local/argocd/application-projects.yaml
	@kubectl apply -f local/argocd/applicationset-services.yaml

.PHONY: argocd-down
argocd-down: ## argocd: tear down argocd wholesale (registered service namespaces untouched)
	@echo $(PROJECT_LOG_FMT) "tearing down argocd"
	@helm uninstall argocd --namespace $(ARGOCD_NAMESPACE) --ignore-not-found
	@kubectl delete namespace $(ARGOCD_NAMESPACE) --ignore-not-found

.PHONY: argocd-ui
argocd-ui: ## argocd: port-forward the argocd UI/API to localhost:8080
	@kubectl port-forward svc/argocd-server -n $(ARGOCD_NAMESPACE) 8080:443
