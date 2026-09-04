package util

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

// So custom Name() wouldn't produce ListActiveRequestRequest.
var reservedNameSuffixes = []string{"Request", "Response", "Row", "Params"}

func ReservedNameSuffix(name string) string {
	for _, suffix := range reservedNameSuffixes {
		if strings.HasSuffix(name, suffix) {
			return suffix
		}
	}

	return ""
}

// check for name collision in sqlc and proto
func ValidateQueryNames(entities []schema.Entity) error {
	if err := validateNamespace(entities, schema.Entity.SQLCQueries, "sqlc query"); err != nil {
		return err
	}

	return validateNamespace(entities, schema.Entity.ProtoQueries, "proto message")
}

func validateNamespace(entities []schema.Entity, queries func(schema.Entity) []schema.Query, namespace string) error {
	type owner struct {
		entity string
		custom bool
	}

	seen := make(map[string]owner)
	for _, entity := range entities {
		for _, query := range queries(entity) {
			name := GenQueryName(query, entity.Name)
			if name == "" {
				continue
			}

			if previous, found := seen[name]; found {
				return fmt.Errorf(
					"%s name %q is used twice, by %s and %s, %s",
					namespace, name, previous.entity, entity.Name, renameHint(previous.custom, query.Name != ""),
				)
			}
			seen[name] = owner{entity: entity.Name, custom: query.Name != ""}
		}
	}

	return nil
}

func renameHint(previousCustom, currentCustom bool) string {
	if previousCustom || currentCustom {
		return "give one of them a different Name()"
	}

	return "the same query is declared twice"
}
