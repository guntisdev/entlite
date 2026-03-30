all: tidy lint build test

tidy:
	go mod tidy
	go mod verify

build:
	go build ./...

lint:
	golangci-lint run ./examples/01-sqlite-entity/server/...

test:
	go test -v ./...

integration:
	go test -v -tags=integration ./examples/...

bin:
	go build -o entlite ./cmd/entlite/main.go
