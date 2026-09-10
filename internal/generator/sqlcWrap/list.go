package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/naming"
	"github.com/guntisdev/entlite/internal/schema"
)

const totalSizeField = "TotalSize"

func (ctx *generationContext) generateListQuery(funcDecl *ast.FuncDecl, entity schema.Entity) string {
	var sb strings.Builder
	inputPkg := ctx.inputPackageName

	withCount := false
	if target, ok := ctx.dslQueries[funcDecl.Name.Name]; ok {
		withCount = target.query.Count
		if target.query.HasDistinct() {
			return ctx.generateDistinctListQuery(funcDecl, target)
		}
		if target.query.HasAggregates() {
			return ctx.generateAggregateQuery(funcDecl, target)
		}
	}

	params, args, prelude := ctx.wrapFilterParams(funcDecl, entity)

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context%s) ", receiverType, funcDecl.Name.Name, params))

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		if withCount {
			sb.WriteString(fmt.Sprintf("([]*%s, int64, error)", entity.Name))
		} else {
			sb.WriteString(fmt.Sprintf("([]*%s, error)", entity.Name))
		}
	}

	sb.WriteString(" {\n")
	sb.WriteString(prelude)

	errReturn := "\t\treturn nil, err\n"
	if withCount {
		errReturn = "\t\treturn nil, 0, err\n"
	}

	sb.WriteString(fmt.Sprintf("\tdbResults, err := (*%s.Queries)(q).%s(ctx%s)\n", inputPkg, funcDecl.Name.Name, args))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(errReturn)
	sb.WriteString("\t}\n")

	sb.WriteString(fmt.Sprintf("\tresult := make([]*%s, len(dbResults))\n", entity.Name))
	sb.WriteString("\tfor i := range dbResults {\n")
	if withCount {
		sb.WriteString(ctx.rowToEntity(entity))
	} else {
		sb.WriteString(fmt.Sprintf("\t\tresult[i] = %sFromSQL(&dbResults[i])\n", entity.Name))
	}
	sb.WriteString("\t}\n")

	if !withCount {
		sb.WriteString("\treturn result, nil\n")
		sb.WriteString("}\n\n")
		return sb.String()
	}

	sb.WriteString("\tvar totalSize int64\n")
	sb.WriteString("\tif len(dbResults) > 0 {\n")
	sb.WriteString(fmt.Sprintf("\t\ttotalSize = dbResults[0].%s\n", totalSizeField))
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn result, totalSize, nil\n")
	sb.WriteString("}\n\n")

	return sb.String()
}

func (ctx *generationContext) generateDistinctListQuery(funcDecl *ast.FuncDecl, target dslQuery) string {
	var sb strings.Builder
	entity := target.entity
	fields := distinctFields(entity, target.query)
	params, args, prelude := ctx.wrapFilterParams(funcDecl, entity)

	elemType := distinctRowName(target.query.Name)
	if len(fields) == 1 {
		elemType = fieldToGoType(fields[0])
	}

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context%s) ([]%s, error) {\n", receiverType, funcDecl.Name.Name, params, elemType))
	sb.WriteString(prelude)

	sb.WriteString(fmt.Sprintf("\tdbResults, err := (*%s.Queries)(q).%s(ctx%s)\n", ctx.inputPackageName, funcDecl.Name.Name, args))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, err\n")
	sb.WriteString("\t}\n")

	sb.WriteString(fmt.Sprintf("\tresult := make([]%s, len(dbResults))\n", elemType))
	sb.WriteString("\tfor i := range dbResults {\n")
	if len(fields) == 1 {
		sb.WriteString(fmt.Sprintf("\t\tresult[i] = %s\n", goFromSQL(fields[0], "dbResults[i]", ctx.sqlDialect)))
	} else {
		sb.WriteString(fmt.Sprintf("\t\tresult[i] = %s{\n", elemType))
		for _, field := range fields {
			fieldName := toDBFieldName(field)
			value := goFromSQL(field, fmt.Sprintf("dbResults[i].%s", fieldName), ctx.sqlDialect)
			sb.WriteString(fmt.Sprintf("\t\t\t%s: %s,\n", fieldName, value))
		}
		sb.WriteString("\t\t}\n")
	}
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn result, nil\n")
	sb.WriteString("}\n\n")

	return sb.String()
}

// grouped query returns one row per group, without a group the aggregates returns single value
func (ctx *generationContext) generateAggregateQuery(funcDecl *ast.FuncDecl, target dslQuery) string {
	var sb strings.Builder
	query := target.query
	fields := aggregateRowFields(target.entity, query)
	params, args, prelude := ctx.wrapFilterParams(funcDecl, target.entity)
	receiverType := formatType(funcDecl.Recv.List[0].Type)
	rowType := naming.RowName(query.Name)

	scalar := ""
	if !query.HasGroupBy() && len(fields) == 1 {
		scalar = fieldToGoType(fields[0])
	}

	results := fmt.Sprintf("([]%s, error)", rowType)
	zero := "nil"
	switch {
	case scalar != "":
		results = fmt.Sprintf("(%s, error)", scalar)
		zero = zeroValue(fields[0])
	case !query.HasGroupBy():
		results = fmt.Sprintf("(%s, error)", rowType)
		zero = rowType + "{}"
	}

	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context%s) %s {\n", receiverType, funcDecl.Name.Name, params, results))
	sb.WriteString(prelude)

	sb.WriteString(fmt.Sprintf("\tdbResult, err := (*%s.Queries)(q).%s(ctx%s)\n", ctx.inputPackageName, funcDecl.Name.Name, args))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\treturn %s, err\n", zero))
	sb.WriteString("\t}\n")

	if scalar != "" {
		sb.WriteString(fmt.Sprintf("\treturn %s, nil\n}\n\n", goFromSQL(fields[0], "dbResult", ctx.sqlDialect)))
		return sb.String()
	}

	if !query.HasGroupBy() {
		sb.WriteString(fmt.Sprintf("\treturn %s", rowType))
		sb.WriteString(ctx.aggregateRowLiteral(fields, "dbResult", "\t"))
		sb.WriteString(", nil\n}\n\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("\tresult := make([]%s, len(dbResult))\n", rowType))
	sb.WriteString("\tfor i := range dbResult {\n")
	sb.WriteString(fmt.Sprintf("\t\tresult[i] = %s", rowType))
	sb.WriteString(ctx.aggregateRowLiteral(fields, "dbResult[i]", "\t\t"))
	sb.WriteString("\n\t}\n")
	sb.WriteString("\treturn result, nil\n}\n\n")

	return sb.String()
}

func (ctx *generationContext) aggregateRowLiteral(fields []schema.Field, dbRef, indent string) string {
	var sb strings.Builder

	sb.WriteString("{\n")
	for _, field := range fields {
		fieldName := toDBFieldName(field)
		value := goFromSQL(field, fmt.Sprintf("%s.%s", dbRef, fieldName), ctx.sqlDialect)
		sb.WriteString(fmt.Sprintf("%s\t%s: %s,\n", indent, fieldName, value))
	}
	sb.WriteString(indent + "}")

	return sb.String()
}

func generateAggregateRowStruct(entity schema.Entity, query schema.Query) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("type %s struct {\n", naming.RowName(query.Name)))
	for _, field := range aggregateRowFields(entity, query) {
		sb.WriteString(fmt.Sprintf("\t%s %s\n", toDBFieldName(field), fieldToGoType(field)))
	}
	sb.WriteString("}\n\n")

	return sb.String()
}

// grouped columns keep their own type, aggregate takes the type it folds into
func aggregateRowFields(entity schema.Entity, query schema.Query) []schema.Field {
	fields := make([]schema.Field, 0, len(query.GroupBy)+len(query.Aggregates))

	for _, fieldName := range query.GroupBy {
		if field, ok := entity.GetFieldByName(fieldName); ok {
			fields = append(fields, field)
		}
	}

	for _, aggregate := range query.Aggregates {
		field, ok := entity.GetFieldByName(aggregate.Field)
		if !ok {
			continue
		}
		fields = append(fields, schema.Field{
			Name: naming.AggregateColumn(string(aggregate.Func), aggregate.Field),
			Type: schema.AggregateResultType(aggregate.Func, field.Type),
		})
	}

	return fields
}

func zeroValue(field schema.Field) string {
	if field.Type == schema.FieldTypeString {
		return `""`
	}

	return "0"
}

func generateDistinctRowStruct(entity schema.Entity, query schema.Query) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("type %s struct {\n", distinctRowName(query.Name)))
	for _, field := range distinctFields(entity, query) {
		sb.WriteString(fmt.Sprintf("\t%s %s\n", toDBFieldName(field), fieldToGoType(field)))
	}
	sb.WriteString("}\n\n")

	return sb.String()
}

func distinctRowName(queryName string) string {
	return naming.RowName(queryName)
}

func distinctFields(entity schema.Entity, query schema.Query) []schema.Field {
	fields := make([]schema.Field, 0, len(query.Distinct))
	for _, fieldName := range query.Distinct {
		if field, ok := entity.GetFieldByName(fieldName); ok {
			fields = append(fields, field)
		}
	}

	return fields
}

func (ctx *generationContext) rowToEntity(entity schema.Entity) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\t\tresult[i] = &%s{\n", entity.Name))
	for _, field := range entity.Fields {
		if field.IsVirtual() {
			continue
		}

		fieldName := toDBFieldName(field)
		convertedValue := goFromSQL(field, fmt.Sprintf("dbResults[i].%s", fieldName), ctx.sqlDialect)
		sb.WriteString(fmt.Sprintf("\t\t\t%s: %s,\n", fieldName, convertedValue))
	}
	sb.WriteString("\t\t}\n")

	return sb.String()
}
