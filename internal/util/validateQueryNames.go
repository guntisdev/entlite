package util

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/naming"
	"github.com/guntisdev/entlite/internal/schema"
)

func ReservedNameSuffix(name string) string {
	for _, suffix := range naming.ReservedSuffixes() {
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
		query  schema.Query
	}

	seen := make(map[string]owner)
	for _, entity := range entities {
		for _, query := range queries(entity) {
			if query.Name == "" {
				continue
			}

			if previous, found := seen[query.Name]; found {
				return fmt.Errorf(
					"%s name %q is used twice, by %s and %s, %s",
					namespace, query.Name, previous.entity, entity.Name,
					renameHint(previous.entity == entity.Name && sameQuery(previous.query, query)),
				)
			}
			seen[query.Name] = owner{entity: entity.Name, query: query}
		}
	}

	return nil
}

func renameHint(duplicate bool) string {
	if duplicate {
		return "the same query is declared twice"
	}

	return "give one of them a different Name()"
}

// reports duplication
func sameQuery(a, b schema.Query) bool {
	if a.Type != b.Type || len(a.Fields) != len(b.Fields) || len(a.Filters) != len(b.Filters) {
		return false
	}

	for i := range a.Fields {
		if !strings.EqualFold(a.Fields[i], b.Fields[i]) {
			return false
		}
	}

	for i := range a.Filters {
		if a.Filters[i] != b.Filters[i] {
			return false
		}
	}

	return true
}
