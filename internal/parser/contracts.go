package parser

import (
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
						contract := parseContractCall(callExpr)
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

func parseContractCall(callExpr *ast.CallExpr) schema.Contract {
	var contract schema.Contract

	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "entlite" {
			switch selExpr.Sel.Name {
			case "SQLC":
				contract.Type = schema.ContractSQLC
			case "PROTO":
				contract.Type = schema.ContractPROTO
			}
		}
	}

	return contract
}
