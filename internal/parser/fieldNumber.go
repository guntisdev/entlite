package parser

import (
	"fmt"
	"slices"
	"sort"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

func addFieldNumbers(fields []schema.Field) []schema.Field {
	var usedNumbers []int
	hasIdField := false

	// checks if there is id field with protoField number
	for _, field := range fields {
		if field.IsID() {
			field.Name = "ID" // capital letter for sqlc compatibility
			hasIdField = true
		}

		if field.ProtoField != 0 {
			usedNumbers = append(usedNumbers, field.ProtoField)
		}
	}

	if !hasIdField {
		idNumber := getNextAvailable(usedNumbers)
		usedNumbers = append(usedNumbers, idNumber)

		idField := schema.Field{
			Name:        "ID",
			Type:        schema.FieldTypeInt,
			ProtoField:  idNumber,
			Unique:      true,
			Permissions: permissions.Default,
		}

		fields = slices.Insert(fields, 0, idField)
	}

	// add proto field numbers if they are missing
	for i := range fields {
		if fields[i].ProtoField == 0 {
			num := getNextAvailable(usedNumbers)
			usedNumbers = append(usedNumbers, num)
			fields[i].ProtoField = num
		}
	}

	return fields
}

// if {1, 2, 4, 6] - it will find 3 as smallest available number
func getNextAvailable(usedNumbers []int) int {
	sort.Ints(usedNumbers)

	candidate := 1

	for _, num := range usedNumbers {
		if num == candidate {
			candidate++
		} else if num > candidate {
			break
		}
	}

	return candidate
}
func checkProtoFieldCollision(fields []schema.Field) error {
	existingNumbers := make(map[int]struct{})

	for _, field := range fields {
		if field.ProtoField == 0 {
			continue
		}

		num := field.ProtoField
		if _, exists := existingNumbers[num]; exists {
			return fmt.Errorf("ProtoField collision detected: field number %d used more than once", num)
		}

		existingNumbers[num] = struct{}{}
	}

	return nil
}
