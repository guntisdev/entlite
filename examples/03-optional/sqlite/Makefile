PORT ?= 8080

.PHONY: run gen web clean

run: gen web
	go run main.go -port=$(PORT)

gen:
	cd ent && go generate

web:
	cd web && npm install && npm run build

clean:
	rm -rf web/node_modules web/dist/bundle.js web/dist/bundle.js.map server/db.db
