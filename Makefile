build:
	@go build -o bin/peervault ./cmd/peervault

run: build
	@./bin/peervault

test:
	@go test ./...