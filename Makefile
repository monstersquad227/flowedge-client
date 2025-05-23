##@ The commands are:

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<command>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: generate
generate: ## Generated Linux
	@rm -rf ./flowedge_client
	GOOS=linux GOARCH=amd64 go build -v -o flowedge_client

.PHONY: generate-mac-intel
generate-mac-intel: ## Generate Mac Intel
	@rm -rf ./flowedge_client
	GOOS=darwin GOARCH=amd64 go build -v -o flowedge_client

.PHONY: generate-win
generate-win: ## Generate Win
	@rm -rf ./flowedge_client.exe
	GOOS=windows GOARCH=amd64 go build -v -o flowedge_client.exe