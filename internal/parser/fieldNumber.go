package parser

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

/*
   TODO add extra logic to put protoField values
   - check if id exists with 1 - if no add it
   - check if fields have duplicated protoField values - if so return error
   - automatically add protoField value if not applied to field, check to not duplicate values
*/

func addFieldNumbers(fields []schema.Field) []schema.Field {
	var usedNumbers []int
	hasIdField := false

	// checks if there is id field with protoField number
	for _, field := range fields {
		if strings.ToLower(field.Name) == "id" {
			hasIdField = true
		}

		if field.ProtoField != nil {
			usedNumbers = append(usedNumbers, *field.ProtoField)
		}
	}

	if !hasIdField {
		idNumber := getNextAvailable(usedNumbers)
		usedNumbers = append(usedNumbers, idNumber)

		idField := schema.Field{
			Name:       "id",
			Type:       schema.FieldTypeInt32,
			ProtoField: &idNumber,
			Unique:     true,
		}

		fields = slices.Insert(fields, 0, idField)
	}

	// add proto field numbers if they are missing
	for _, field := range fields {
		if field.ProtoField == nil {
			num := getNextAvailable(usedNumbers)
			usedNumbers = append(usedNumbers, num)
			field.ProtoField = &num
			fmt.Printf("num: %d\n", num)
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
		if field.ProtoField == nil {
			continue
		}

		num := *field.ProtoField
		if _, exists := existingNumbers[num]; exists {
			return fmt.Errorf("ProtoField collision detected: field number %d used more than once", num)
		}

		existingNumbers[num] = struct{}{}
	}

	return nil
}
