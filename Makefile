all: tidy lint build test

tidy:
	go mod tidy
	go mod verify

build:
	go build ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

test:
	go test -v ./...

# regenerates every example, needs sqlc, buf and a network
gen:
	@for dir in examples/*/*/ent; do \
		echo "==> $$dir"; \
		(cd $$dir && go generate .) || exit 1; \
	done

# regenerates docs/ from the source tree
docs:
	go run ./internal/docs/cmd -out docs

# builds the html site into dist/, it regenerates docs/ on the way
site:
	go run ./internal/docs/cmd -out docs -html dist

# runs through each example and checks if nothing is broken
integration:
	go test -v -count=1 -timeout=30m -tags=integration ./examples/...

bin:
	go build -o entlite ./cmd/entlite/main.go
