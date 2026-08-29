package protovalidate

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/parser"
	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
)

func Generate(entities []schema.Entity, imports map[string]parser.ImportInfo) (string, error) {
	var content strings.Builder

	var methods strings.Builder
	for _, entity := range entities {
		if !hasValidateField(entity) && !hasJSONField(entity) {
			continue
		}
		for _, queryType := range []schema.QueryType{schema.QueryCreate, schema.QueryUpdate} {
			if !hasQueryType(entity, queryType) {
				continue
			}
			methods.WriteString(generateValidateMethod(entity, queryType))
		}
	}
	body := methods.String()

	content.WriteString(util.GeneratedGo)
	content.WriteString("package pb\n\n")

	content.WriteString("import (\n")
	content.WriteString("\t\"context\"\n")
	if strings.Contains(body, "json.") {
		content.WriteString("\t\"encoding/json\"\n")
	}
	if strings.Contains(body, "fmt.") {
		content.WriteString("\t\"fmt\"\n")
	}
	content.WriteString("\t\"connectrpc.com/connect\"\n")

	for _, importInfo := range imports {
		content.WriteString(fmt.Sprintf("\t\"%s\"\n", importInfo.Path))
	}
	content.WriteString(")\n\n")

	content.WriteString(body)

	content.WriteString(interceptorSource)

	return content.String(), nil
}

const interceptorSource = `type validator interface {
	Validate() error
}

// ValidateInterceptor calls the generated Validate() method on any request message that implements it
type ValidateInterceptor struct{}

var _ connect.Interceptor = (*ValidateInterceptor)(nil)

func NewValidateInterceptor() *ValidateInterceptor {
	return &ValidateInterceptor{}
}

func validateMsg(msg any) error {
	v, ok := msg.(validator)
	if !ok {
		return nil
	}
	if err := v.Validate(); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return nil
}

// WrapUnary implements connect.Interceptor.
func (i *ValidateInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := validateMsg(req.Any()); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (i *ValidateInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements connect.Interceptor.
func (i *ValidateInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, &validatingHandlerConn{StreamingHandlerConn: conn})
	}
}

type validatingHandlerConn struct {
	connect.StreamingHandlerConn
}

func (c *validatingHandlerConn) Receive(msg any) error {
	if err := c.StreamingHandlerConn.Receive(msg); err != nil {
		return err
	}
	return validateMsg(msg)
}
`

// generateValidateMethod writes Validate() for the create/update queries
func generateValidateMethod(entity schema.Entity, queryType schema.QueryType) string {
	var content strings.Builder
	messageName := util.GenEntityQueryName(entity, queryType)
	content.WriteString(fmt.Sprintf("func (r *%sRequest) Validate() error {\n", messageName))

	// json text is checked before the request reaches the handler
	for _, field := range entity.Fields {
		if field.Type != schema.FieldTypeJSON || !inRequest(field, queryType) {
			continue
		}

		fieldName := toProtoFieldName(field)
		ref := fmt.Sprintf("r.%s", fieldName)
		cond := fmt.Sprintf("!json.Valid([]byte(%s))", ref)
		if isPointerField(field, queryType) {
			cond = fmt.Sprintf("%s != nil && !json.Valid([]byte(*%s))", ref, ref)
		}
		content.WriteString(fmt.Sprintf("\tif %s {\n", cond))
		content.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"Invalid json for field name: %s\")\n", fieldName))
		content.WriteString("\t}\n")
	}

	// TODO fix Optional() with Validate() - a pointer is passed to a value func and does not compile
	for _, field := range entity.Fields {
		if field.Validate == nil {
			continue
		}

		fieldName := toProtoFieldName(field)
		content.WriteString(fmt.Sprintf("\tif !%s(r.%s) {\n", field.Validate().(string), fieldName))
		content.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"Validation failed for field name: %s\")\n", fieldName))
		content.WriteString("\t}\n")
	}
	content.WriteString("\treturn nil\n")
	content.WriteString("}\n\n")
	return content.String()
}

// inRequest reports if the field exists in the create or update request message
func inRequest(field schema.Field, queryType schema.QueryType) bool {
	if !field.CanApiWrite() {
		return false
	}
	if queryType == schema.QueryCreate {
		return !field.IsID()
	}
	return field.IsID() || !field.Immutable
}

// isPointerField reports if protoc emits the field as a pointer
func isPointerField(field schema.Field, queryType schema.QueryType) bool {
	if field.Optional || field.DefaultValue != nil || field.DefaultFunc != nil {
		return true
	}
	// update makes write-only fields optional so they can be left alone
	return queryType == schema.QueryUpdate && !field.CanApiRead()
}

// only a declared query has a request message to hang Validate() on
func hasQueryType(entity schema.Entity, queryType schema.QueryType) bool {
	for _, query := range entity.ProtoQueries() {
		if query.Type == queryType {
			return true
		}
	}

	return false
}

func hasJSONField(entity schema.Entity) bool {
	for _, field := range entity.Fields {
		if field.Type == schema.FieldTypeJSON {
			return true
		}
	}
	return false
}

func hasValidateField(entity schema.Entity) bool {
	for _, field := range entity.Fields {
		if field.Validate != nil {
			return true
		}
	}
	return false
}

func toProtoFieldName(field schema.Field) string {
	if field.IsID() {
		return strings.ToUpper(field.Name[:1]) + field.Name[1:]
	}
	return snakeToCamelCase(field.Name)
}

func snakeToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}
