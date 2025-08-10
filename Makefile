build:
	@go build -o bin/distrigo ./cmd/distrigo

run: build
	@./bin/distrigo

test:
	@go test ./...