package naming

import "testing"

// names sqlc and protoc really give, read from their output
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
		// sqlc lowercases first, so the capital is lost. protoc keeps it
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
		{"PopularProduct", "popular_product"},
		{"User2Factor", "user2_factor"},
	}

	for _, tt := range tests {
		if got := TableName(tt.entity); got != tt.table {
			t.Errorf("TableName(%q) = %q, expected %q", tt.entity, got, tt.table)
		}
	}
}

func TestTableNameRoundTrips(t *testing.T) {
	names := []string{"User", "MyUser", "PopularProduct", "Build", "User2Factor", "TableConfig"}

	for _, name := range names {
		table := TableName(name)
		if got := SqlcGoName(table); got != name {
			t.Errorf("entity %q becomes table %q, which sqlc reads back as %q", name, table, got)
		}
	}
}

func TestValidateEntityName(t *testing.T) {
	valid := []string{"User", "MyUser", "PopularProduct", "User2Factor"}
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
		{"popular_product", "PopularProduct"},
		{"popularProduct", "PopularProduct"},
		{"Popular_Product", "PopularProduct"},
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

func TestSuffixNames(t *testing.T) {
	if got := ServiceName("MyUser"); got != "MyUserService" {
		t.Errorf("ServiceName = %q", got)
	}
	if got := RequestName("CreateMyUser"); got != "CreateMyUserRequest" {
		t.Errorf("RequestName = %q", got)
	}
	if got := ResponseName("ListMyUsers"); got != "ListMyUsersResponse" {
		t.Errorf("ResponseName = %q", got)
	}
	if got := RowName("CreateBulkMyUser"); got != "CreateBulkMyUserRow" {
		t.Errorf("RowName = %q", got)
	}
	if got := ParamsName("CreateMyUser"); got != "CreateMyUserParams" {
		t.Errorf("ParamsName = %q", got)
	}
	// every suffix we add must be one a custom Name() cannot use
	reserved := map[string]bool{}
	for _, s := range ReservedSuffixes() {
		reserved[s] = true
	}
	for _, s := range []string{SuffixRequest, SuffixResponse, SuffixRow, SuffixParams} {
		if !reserved[s] {
			t.Errorf("suffix %q is appended but not reserved", s)
		}
	}
}

func TestQueryNames(t *testing.T) {
	const e = "MyUser"

	tests := []struct{ got, want string }{
		{CreateQueryName(e), "CreateMyUser"},
		{CreateBulkQueryName(e), "CreateBulkMyUser"},
		{UpdateQueryName(e), "UpdateMyUser"},
		{DeleteQueryName(e), "DeleteMyUser"},
		{DeleteAllQueryName(e), "DeleteAllMyUser"},
		{GetByQueryName(e, []string{"id"}), "GetMyUserById"},
		{GetByQueryName(e, []string{"org_id", "email"}), "GetMyUserByOrgIdEmail"},
		{ListAllQueryName(e, nil), "ListAllMyUser"},
		{ListAllQueryName(e, []string{"branch"}), "ListAllMyUserDistinctBranch"},
		{ListByQueryName(e, nil, []string{"is_active"}, nil), "ListMyUserByIsActive"},
		{ListByQueryName(e, nil, nil, []string{"env", "status"}), "ListMyUserFilterByEnvStatus"},
		{ListByQueryName(e, []string{"branch"}, nil, nil), "ListMyUserDistinctBranch"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("got %q, expected %q", tt.got, tt.want)
		}
	}
}

// a query name joins a digit word, protoc splits it
// so the two rules cannot be one function
func TestQueryWordsIsNotProtocGoName(t *testing.T) {
	if got := QueryWords([]string{"field_2"}); got != "Field2" {
		t.Errorf("QueryWords = %q, expected Field2", got)
	}
	if got := ProtocGoName("field_2"); got != "Field_2" {
		t.Errorf("ProtocGoName = %q, expected Field_2", got)
	}
}
