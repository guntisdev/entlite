TODO

* change folder structure so sqlc generates in gen/db/internal
* create tools.go for sqlc, buf, connect // or from 1.24 tool in go.mod https://go.dev/doc/go1.24#tools
* create generate.go to launch entity generation and then sqlc and buf
* provide wrapper around sqlc export, to support custom functions on .DefaultFunc() and .Validate() 
