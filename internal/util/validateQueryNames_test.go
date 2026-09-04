package util

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func bothContracts() []schema.Contract {
	return []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
}

// testEntity builds an entity where every query targets both contracts
func testEntity(name string, queries ...schema.Query) schema.Entity {
	for i := range queries {
		queries[i].Contracts = bothContracts()
	}

	return schema.Entity{Name: name, Contracts: bothContracts(), Queries: queries}
}

func TestValidateQueryNames(t *testing.T) {
	tests := []struct {
		name     string
		entities []schema.Entity
		wantErr  string
	}{
		{
			name: "generated names never collide",
			entities: []schema.Entity{
				testEntity("User", schema.Query{Type: schema.QueryGetBy, Fields: []string{"email"}}),
				testEntity("Post", schema.Query{Type: schema.QueryGetBy, Fields: []string{"email"}}),
			},
		},
		{
			name: "custom name on both entities",
			entities: []schema.Entity{
				testEntity("User", schema.Query{Type: schema.QueryListAll, Name: "ListActive"}),
				testEntity("Post", schema.Query{Type: schema.QueryListAll, Name: "ListActive"}),
			},
			wantErr: `"ListActive" is used twice, by User and Post, give one of them a different Name()`,
		},
		{
			name: "custom name collides with a generated one",
			entities: []schema.Entity{
				testEntity("User", schema.Query{Type: schema.QueryGetBy, Fields: []string{"email"}}),
				testEntity("Post", schema.Query{Type: schema.QueryListAll, Name: "GetUserByEmail"}),
			},
			wantErr: `"GetUserByEmail" is used twice, by User and Post, give one of them a different Name()`,
		},
		{
			name: "same query declared twice on one entity",
			entities: []schema.Entity{
				testEntity("User",
					schema.Query{Type: schema.QueryGetBy, Fields: []string{"email"}},
					schema.Query{Type: schema.QueryGetBy, Fields: []string{"email"}},
				),
			},
			wantErr: `"GetUserByEmail" is used twice, by User and User, the same query is declared twice`,
		},
		{
			name: "one name per contract is fine",
			entities: []schema.Entity{
				{
					Name:      "User",
					Contracts: bothContracts(),
					Queries: []schema.Query{{
						Type:      schema.QueryListAll,
						Name:      "ListActive",
						Contracts: []schema.Contract{{Type: schema.ContractSQLC}},
					}},
				},
				{
					Name:      "Post",
					Contracts: bothContracts(),
					Queries: []schema.Query{{
						Type:      schema.QueryListAll,
						Name:      "ListActive",
						Contracts: []schema.Contract{{Type: schema.ContractPROTO}},
					}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateQueryNames(test.entities)

			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got none", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("expected error containing %q, got %q", test.wantErr, err.Error())
			}
		})
	}
}

func TestReservedNameSuffix(t *testing.T) {
	tests := map[string]string{
		"ListActive":         "",
		"ListActiveRequest":  "Request",
		"ListActiveResponse": "Response",
		"CreateBulkUserRow":  "Row",
		"ListActiveParams":   "Params",
	}

	for name, want := range tests {
		if got := ReservedNameSuffix(name); got != want {
			t.Errorf("ReservedNameSuffix(%q) = %q, want %q", name, got, want)
		}
	}
}
