BINARY := devops-server
PKG    := ./cmd/server

.PHONY: build run run-http tidy fmt vet test clean

build: ## Compile the server binary into ./bin
	go build -o bin/$(BINARY) $(PKG)

run: ## Run over stdio (default, for locally spawned agents)
	go run $(PKG)

run-http: ## Run over streamable HTTP on MCP_ADDR
	MCP_TRANSPORT=http go run $(PKG)

tidy: ## Sync go.mod / go.sum
	go mod tidy

fmt: ## Format all Go sources
	gofmt -w .

vet: ## Run go vet
	go vet ./...

test: ## Run tests
	go test ./...

clean: ## Remove build output
	rm -rf bin
