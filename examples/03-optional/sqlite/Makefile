.PHONY: run gen clean

run: gen
	go run main.go

gen:
	cd ent && go generate

clean:
	rm -rf server/db.db
