.PHONY: install-tools
install-tools:
	mkdir -p tools
	GOBIN=$(PWD)/tools go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	GOBIN=$(PWD)/tools go install golang.org/x/tools/cmd/goimports@latest

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	@./tools/golangci-lint run ./...

.PHONY: check-imports
check-imports:
	@files=$$(./tools/goimports -l .); if [ -n "$$files" ]; then echo "$$files"; exit 1; fi

.PHONY: fix-imports
fix-imports:
	@./tools/goimports -w .

.PHONY: check-run
check-run:
	go run  cmd/main.go --debug --config ./test/config_debug.yml