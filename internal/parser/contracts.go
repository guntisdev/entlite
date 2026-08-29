package parser

import (
	"fmt"
	"go/ast"

	"github.com/guntisdev/entlite/internal/schema"
)

func parseContractsMethod(funcDecl *ast.FuncDecl) ([]schema.Contract, error) {
	var contracts []schema.Contract

	if funcDecl.Body == nil {
		return contracts, nil
	}

	for _, stmt := range funcDecl.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		for _, result := range retStmt.Results {
			if compLit, ok := result.(*ast.CompositeLit); ok {
				for _, elt := range compLit.Elts {
					if callExpr, ok := elt.(*ast.CallExpr); ok {
						contract, err := parseContractCall(callExpr)
						if err != nil {
							return nil, err
						}
						if contract.Type != "" {
							contracts = append(contracts, contract)
						}
					}
				}
			}
		}
	}

	return contracts, nil
}

func parseContractCall(callExpr *ast.CallExpr) (schema.Contract, error) {
	var contract schema.Contract

	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return contract, nil
	}

	if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "entlite" {
		switch selExpr.Sel.Name {
		case "SQLC":
			contract.Type = schema.ContractSQLC
		case "PROTO":
			contract.Type = schema.ContractPROTO
		}

		return contract, nil
	}

	innerCall, ok := selExpr.X.(*ast.CallExpr)
	if !ok {
		return contract, nil
	}

	contract, err := parseContractCall(innerCall)
	if err != nil || contract.Type == "" {
		return contract, err
	}

	if len(callExpr.Args) != 0 {
		return contract, fmt.Errorf("%s does not accept arguments", selExpr.Sel.Name)
	}

	switch selExpr.Sel.Name {
	case "ReadOnly":
		contract.Access = schema.AccessRead
	case "WriteOnly":
		contract.Access = schema.AccessWrite
	default:
		return contract, fmt.Errorf("unsupported contract operation %q, expected ReadOnly or WriteOnly", selExpr.Sel.Name)
	}

	return contract, nil
}
