run:
	go run .

lint:
	go tool goimports -w .
	gofmt -w .
	go tool golangci-lint run -v

test:
	go test ./... -v

generate:
	go generate ./...
