package naming

import "testing"

// the values sqlc and protoc actually produce, measured from their output
func TestGoNames(t *testing.T) {
	tests := []struct{ sqlName, sqlc, protoc string }{
		{"id", "ID", "Id"},
		{"user_id", "UserID", "UserId"},
		{"ids", "Ids", "Ids"},
		{"is_active", "IsActive", "IsActive"},
		{"api_key", "ApiKey", "ApiKey"},
		{"http_status", "HttpStatus", "HttpStatus"},
		{"display_name", "DisplayName", "DisplayName"},
		{"sample_rate_ms", "SampleRateMs", "SampleRateMs"},
		// sqlc lowercases first, so a capital in the column is lost, protoc keeps it
		{"initCount", "Initcount", "InitCount"},
	}

	for _, tt := range tests {
		if got := SqlcGoName(tt.sqlName); got != tt.sqlc {
			t.Errorf("SqlcGoName(%q) = %q, expected %q", tt.sqlName, got, tt.sqlc)
		}
		if got := ProtocGoName(tt.sqlName); got != tt.protoc {
			t.Errorf("ProtocGoName(%q) = %q, expected %q", tt.sqlName, got, tt.protoc)
		}
	}
}

func TestTableName(t *testing.T) {
	tests := []struct{ entity, table string }{
		{"User", "user"},
		{"MyUser", "my_user"},
		{"PopularCasino", "popular_casino"},
		{"User2Factor", "user2_factor"},
	}

	for _, tt := range tests {
		if got := TableName(tt.entity); got != tt.table {
			t.Errorf("TableName(%q) = %q, expected %q", tt.entity, got, tt.table)
		}
	}
}

func TestTableNameRoundTrips(t *testing.T) {
	names := []string{"User", "MyUser", "PopularCasino", "Build", "User2Factor", "TableConfig"}

	for _, name := range names {
		table := TableName(name)
		if got := SqlcGoName(table); got != name {
			t.Errorf("entity %q becomes table %q, which sqlc reads back as %q", name, table, got)
		}
	}
}

func TestValidateEntityName(t *testing.T) {
	valid := []string{"User", "MyUser", "PopularCasino", "User2Factor"}
	for _, name := range valid {
		if err := ValidateEntityName(name); err != nil {
			t.Errorf("expected %q to be valid, got: %v", name, err)
		}
	}

	invalid := []string{"", "user", "myUser", "my_user", "My_User", "HTTPServer", "UserID", "MyDB", "My User"}
	for _, name := range invalid {
		if err := ValidateEntityName(name); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
}

func TestValidateFieldName(t *testing.T) {
	valid := []string{"id", "env", "is_active", "api_key", "sample_rate_ms", "last_viewed_ms", "user2"}
	for _, name := range valid {
		if err := ValidateFieldName(name); err != nil {
			t.Errorf("expected %q to be valid, got: %v", name, err)
		}
	}

	invalid := []string{"", "ID", "Id", "isActive", "IsActive", "_id", "is__active", "is_active_", "is-active"}
	for _, name := range invalid {
		if err := ValidateFieldName(name); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
}

func TestPascalSuggestion(t *testing.T) {
	tests := []struct{ in, want string }{
		{"popular_casino", "PopularCasino"},
		{"popularCasino", "PopularCasino"},
		{"Popular_Casino", "PopularCasino"},
		{"HTTPServer", "HttpServer"},
		{"UserID", "UserId"},
		{"MyDB", "MyDb"},
		{"initCount", "InitCount"},
		{"ID", "Id"},
	}

	for _, tt := range tests {
		if got := PascalSuggestion(tt.in); got != tt.want {
			t.Errorf("PascalSuggestion(%q) = %q, expected %q", tt.in, got, tt.want)
		}
	}

	for _, tt := range tests {
		if err := ValidateEntityName(tt.want); err != nil {
			t.Errorf("suggested %q is itself invalid: %v", tt.want, err)
		}
	}
}
