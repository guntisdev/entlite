PORT ?= 8080

.PHONY: run gen clean

run: gen
	go run main.go -port=$(PORT)

gen:
	cd ent && go generate

clean:
	rm -f server/db.db
