package protovalidate

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func notBlankValidate() any { return "logic.NotBlank" }

func TestValidateMethodGuardsOptionalPointer(t *testing.T) {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
	entity := schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "subtitle", Type: schema.FieldTypeString, Optional: true, Contracts: contracts, Validate: notBlankValidate},
		},
	}
	query := schema.Query{Type: schema.QueryCreate, Name: "CreateArticle"}

	got := generateValidateMethod(entity, query)

	want := "if r.Subtitle != nil && !logic.NotBlank(*r.Subtitle) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the generated Validate():\n%s", want, got)
	}

	if strings.Contains(got, "!logic.NotBlank(r.Subtitle)") {
		t.Errorf("validate call should not pass a pointer directly:\n%s", got)
	}
}

// a mandatory field's validate call stays a plain, dereference-free call.
func TestValidateMethodSkipsGuardForMandatoryField(t *testing.T) {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
	entity := schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts, Validate: notBlankValidate},
		},
	}
	query := schema.Query{Type: schema.QueryCreate, Name: "CreateArticle"}

	got := generateValidateMethod(entity, query)

	want := "if !logic.NotBlank(r.Title) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the generated Validate():\n%s", want, got)
	}
}
