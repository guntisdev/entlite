// naming rules that turn a schema into sqlc, proto, sqlc go, proto go, typescript
package naming

import (
	"fmt"
	"strings"
)

// sql table name, MyUser becomes my_user.
func TableName(entity string) string {
	var b strings.Builder

	for i := 0; i < len(entity); i++ {
		c := entity[i]
		if isUpper(c) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteByte(c + ('a' - 'A'))
			continue
		}
		b.WriteByte(c)
	}

	return b.String()
}

// sqlc lowercases the whole identifier, splits it on underscores and titles each part
// part equal to id becomes ID, so user_id is UserID.
func SqlcGoName(sqlName string) string {
	var b strings.Builder

	for _, part := range strings.Split(strings.ToLower(sqlName), "_") {
		if part == "" {
			continue
		}
		if part == "id" {
			b.WriteString("ID")
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}

	return b.String()
}

func ValidateEntityName(name string) error {
	if name == "" {
		return fmt.Errorf("entity name is empty")
	}
	if !isUpper(name[0]) {
		return fmt.Errorf("entity %q must start with a capital, it becomes an exported Go type, write %s", name, PascalSuggestion(name))
	}
	for i := 0; i < len(name); i++ {
		if !isLetter(name[i]) && !isDigit(name[i]) {
			return fmt.Errorf("entity %q must be PascalCase, letters and digits only, e.g. MyUser", name)
		}
		// two capitals in a row break the table name: MyDB -> my_d_b
		if i > 0 && isUpper(name[i]) && isUpper(name[i-1]) {
			return fmt.Errorf("entity %q has two capitals in a row, the table name needs one capital per word, write %s", name, PascalSuggestion(name))
		}
	}

	return nil
}

func ValidateFieldName(name string) error {
	if name == "" {
		return fmt.Errorf("field name is empty")
	}
	if !isLower(name[0]) {
		return fmt.Errorf("field %q must be lower snake_case, it is the sql column and the proto field, write %s", name, TableName(PascalSuggestion(name)))
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case isLower(c) || isDigit(c):
		case c == '_':
			// _ at the end, or two in a row, gives an empty word
			if i == len(name)-1 || name[i+1] == '_' {
				return fmt.Errorf("field %q has an empty word, write the underscores between words, e.g. is_active", name)
			}
		case isUpper(c):
			return fmt.Errorf("field %q must be lower snake_case, a capital is lost on the way through sqlc, write %s", name, TableName(PascalSuggestion(name)))
		default:
			return fmt.Errorf("field %q must be lower snake_case, letters, digits and underscores only, e.g. is_active", name)
		}
	}

	return nil
}

// rewrites letter queue of capitals in one word: HTTPServer -> HttpServer
func PascalSuggestion(name string) string {
	var words []string
	var word []byte

	for i := 0; i < len(name); i++ {
		c := name[i]
		if !isLetter(c) && !isDigit(c) {
			if len(word) > 0 {
				words = append(words, string(word))
				word = nil
			}
			continue
		}
		if isUpper(c) && len(word) > 0 {
			prev := word[len(word)-1]
			nextIsLower := i+1 < len(name) && isLower(name[i+1])
			// a capital after a small letter starts a word. so does the last capital of a run
			if isLower(prev) || nextIsLower {
				words = append(words, string(word))
				word = nil
			}
		}
		word = append(word, c)
	}
	if len(word) > 0 {
		words = append(words, string(word))
	}

	var b strings.Builder
	for _, w := range words {
		b.WriteString(strings.ToUpper(w[:1]) + strings.ToLower(w[1:]))
	}

	return b.String()
}

func isUpper(c byte) bool  { return c >= 'A' && c <= 'Z' }
func isLower(c byte) bool  { return c >= 'a' && c <= 'z' }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }
func isLetter(c byte) bool { return isUpper(c) || isLower(c) }
