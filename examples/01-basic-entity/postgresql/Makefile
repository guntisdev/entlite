# Prefer the compose plugin, fall back to the standalone binary
COMPOSE := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

PORT ?= 8080

.PHONY: run db gen web down clean

run: db gen web
	go run main.go -port=$(PORT)

db:
	$(COMPOSE) up -d --wait

gen:
	cd ent && go generate

web:
	cd web && npm install && npm run build

down:
	$(COMPOSE) down

clean: down
	rm -rf web/node_modules web/dist/bundle.js web/dist/bundle.js.map
