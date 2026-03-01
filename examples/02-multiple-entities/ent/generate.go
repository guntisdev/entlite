//go:generate go generate ./schema
//go:generate go tool sqlc generate
//go:generate go tool buf generate
//go:generate go run github.com/guntisdev/entlite/cmd/entlite sqlc-wrap ./gen/db/internal ./gen/db
//go:generate go run github.com/guntisdev/entlite/cmd/entlite proto-validate ./gen/pb
//go:generate go run github.com/guntisdev/entlite/cmd/entlite convert ./gen/db ./gen/pb

package ent
