all: tidy lint build test

tidy:
	go mod tidy
	go mod verify

build:
	go build ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./examples/...

test:
	go test -v ./...

# runs through each example and checks if nothing is broken
integration:
	go test -v -count=1 -timeout=30m -tags=integration ./examples/...

bin:
	go build -o entlite ./cmd/entlite/main.go
