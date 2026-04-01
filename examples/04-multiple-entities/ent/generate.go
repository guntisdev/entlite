//go:generate go generate ./schema
//go:generate go tool sqlc generate
//go:generate go tool buf dep update
//go:generate go tool buf generate
//go:generate go run github.com/guntisdev/entlite/cmd/entlite sqlc-wrap
//go:generate go run github.com/guntisdev/entlite/cmd/entlite proto-validate ./gen/pb

package ent
