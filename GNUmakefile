GO_VER        ?= go
SWEEP         ?= all
SWEEP_TIMEOUT ?= 360m

.DEFAULT_GOAL := help

##@ General

.PHONY: help
help: ## Show this help
	@awk 'BEGIN { FS = ":.*## " } \
	  /^##@ / { printf "\n\033[1m%s\033[0m\n", substr($$0, 5); next } \
	  /^[a-zA-Z0-9_.\/-]+:.*## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Development

.PHONY: deps
deps: ## Download Go module dependencies
	$(GO_VER) mod download

.PHONY: generate
generate: ## Generate code and documentation
	$(GO_VER) generate ./internal/apiclient
	$(GO_VER) generate ./

##@ Testing

.PHONY: testacc
testacc: ## Run acceptance tests
	TF_ACC=1 $(GO_VER) test ./... -v $(TESTARGS) -timeout 120m

.PHONY: sweep
sweep: ## Run sweepers
	# make sweep SWEEPARGS=-sweep-run=porkbun_url_forward
	# set SWEEPARGS=-sweep-allow-failures to continue after first failure
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO_VER) test ./... -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout $(SWEEP_TIMEOUT)

.PHONY: sweeper
sweeper: ## Run sweepers with failures allowed
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO_VER) test ./... -v -tags=sweep -sweep=$(SWEEP) -sweep-allow-failures -timeout $(SWEEP_TIMEOUT)
