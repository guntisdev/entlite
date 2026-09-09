package naming

// suffixes sqlc and proto add to a name
// kept here because 4 generators used to repeat them
const (
	SuffixService  = "Service"
	SuffixRequest  = "Request"
	SuffixResponse = "Response"
	SuffixRow      = "Row"
	SuffixParams   = "Params"
)

// service of an entity: MyUser -> MyUserService
func ServiceName(entity string) string {
	return entity + SuffixService
}

// request message of a query
func RequestName(query string) string {
	return query + SuffixRequest
}

// response message of a query
// a query that returns the entity has no response message
func ResponseName(query string) string {
	return query + SuffixResponse
}

// row message of a bulk insert or a multi column distinct
func RowName(query string) string {
	return query + SuffixRow
}

// argument struct sqlc makes
// sqlc makes it only when the query takes 2 or more arguments
func ParamsName(query string) string {
	return query + SuffixParams
}

// endings a custom Name() cannot use
// name ListActive must not give the message ListActiveRequestRequest
func ReservedSuffixes() []string {
	return []string{SuffixRequest, SuffixResponse, SuffixRow, SuffixParams}
}
