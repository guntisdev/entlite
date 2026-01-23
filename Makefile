all: tidy build test

tidy:
	go mod tidy
	go mod verify

build:
	go build ./...

test:
	go test -v ./...

bin:
	go build -o entlite ./cmd/entlite/main.go
