package sqlcwrap

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func notBlankValidate() any { return "logic.NotBlank" }

func TestValidationChecksGuardOptionalPointer(t *testing.T) {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
	entity := schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "subtitle", Type: schema.FieldTypeString, Optional: true, Contracts: contracts, Validate: notBlankValidate},
		},
	}

	got := addValidationChecks(entity, "create", errorOnlyReturn, "arg", "\t")

	want := "if arg.Subtitle != nil && !logic.NotBlank(*arg.Subtitle) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the generated checks:\n%s", want, got)
	}

	if strings.Contains(got, "!logic.NotBlank(arg.Subtitle)") {
		t.Errorf("validate call should not pass a pointer directly:\n%s", got)
	}
}

// a mandatory field's validate call stays a plain, dereference-free call.
func TestValidationChecksSkipGuardForMandatoryField(t *testing.T) {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
	entity := schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts, Validate: notBlankValidate},
		},
	}

	got := addValidationChecks(entity, "create", errorOnlyReturn, "arg", "\t")

	want := "if !logic.NotBlank(arg.Title) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the generated checks:\n%s", want, got)
	}
}

// an update skips immutable fields entirely, since they are not in the params struct.
func TestValidationChecksSkipImmutableFieldOnUpdate(t *testing.T) {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
	entity := schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "slug", Type: schema.FieldTypeString, Immutable: true, Contracts: contracts, Validate: notBlankValidate},
		},
	}

	got := addValidationChecks(entity, "update", "nil", "arg", "\t")

	if strings.Contains(got, "Slug") {
		t.Errorf("expected no reference to the immutable field on update:\n%s", got)
	}
}
