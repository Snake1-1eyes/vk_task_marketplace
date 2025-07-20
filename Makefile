update:
	go mod tidy

linter:
	golangci-lint run ./...