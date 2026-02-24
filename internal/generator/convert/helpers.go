package convert

import "strings"

func generateHelperFunctions() string {
	var content strings.Builder

	content.WriteString("// Helper functions for type conversions\n\n")

	// TimeToProtoTimestamp
	content.WriteString("// TimeToProtoTimestamp converts a Go time.Time pointer to proto Timestamp\n")
	content.WriteString("func TimeToProtoTimestamp(t *time.Time) *timestamppb.Timestamp {\n")
	content.WriteString("\tif t == nil {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn timestamppb.New(*t)\n")
	content.WriteString("}\n\n")

	// ProtoTimestampToTime
	content.WriteString("// ProtoTimestampToTime converts a proto Timestamp to Go time.Time pointer\n")
	content.WriteString("func ProtoTimestampToTime(ts *timestamppb.Timestamp) *time.Time {\n")
	content.WriteString("\tif ts == nil {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n")
	content.WriteString("\tt := ts.AsTime()\n")
	content.WriteString("\treturn &t\n")
	content.WriteString("}\n\n")

	// ProtoTimestampToTimeValue
	content.WriteString("// ProtoTimestampToTimeValue converts a proto Timestamp to Go time.Time value\n")
	content.WriteString("func ProtoTimestampToTimeValue(ts *timestamppb.Timestamp) time.Time {\n")
	content.WriteString("\tif ts == nil {\n")
	content.WriteString("\t\treturn time.Time{}\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn ts.AsTime()\n")
	content.WriteString("}\n\n")

	// StringPtr
	content.WriteString("// StringPtr returns a pointer to the string value\n")
	content.WriteString("func StringPtr(s string) *string {\n")
	content.WriteString("\treturn &s\n")
	content.WriteString("}\n\n")

	// StringValue
	content.WriteString("// StringValue returns the string value from a pointer, or empty string if nil\n")
	content.WriteString("func StringValue(s *string) string {\n")
	content.WriteString("\tif s == nil {\n")
	content.WriteString("\t\treturn \"\"\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn *s\n")
	content.WriteString("}\n\n")

	// Int32Ptr
	content.WriteString("// Int32Ptr returns a pointer to the int32 value\n")
	content.WriteString("func Int32Ptr(i int32) *int32 {\n")
	content.WriteString("\treturn &i\n")
	content.WriteString("}\n\n")

	// Int32Value
	content.WriteString("// Int32Value returns the int32 value from a pointer, or 0 if nil\n")
	content.WriteString("func Int32Value(i *int32) int32 {\n")
	content.WriteString("\tif i == nil {\n")
	content.WriteString("\t\treturn 0\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn *i\n")
	content.WriteString("}\n\n")

	// NullInt32Value
	content.WriteString("// NullInt32ToPtr converts sql.NullInt32 to int32 pointer\n")
	content.WriteString("func NullInt32ToPtr(n sql.NullInt32) *int32 {\n")
	content.WriteString("\tif !n.Valid {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn &n.Int32\n")
	content.WriteString("}\n\n")

	content.WriteString("// PtrToNullInt32 converts int32 pointer to NullInt32\n")
	content.WriteString("func PtrToNullInt32(i *int32) sql.NullInt32 {\n")
	content.WriteString("\tif i == nil {\n")
	content.WriteString("\t\treturn sql.NullInt32{Valid: false}\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn sql.NullInt32{Int32: *i, Valid: true}\n")
	content.WriteString("}\n\n")

	return content.String()
}
