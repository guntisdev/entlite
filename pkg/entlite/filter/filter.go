// Package filter holds the filters used by a ListBy query.
package filter

// Filter is one condition on a field in a ListBy query.
type Filter interface {
	Filter()
	// GetField returns the field the filter works on.
	GetField() string
	// IsOptional reports if the filter can be skipped at call time.
	IsOptional() bool
}

// RangeFilter matches values between a lower and an upper bound.
type RangeFilter struct {
	field    string
	optional bool
}

// marker method for sealed interface
func (rf RangeFilter) Filter() {}

// GetField returns the field the filter works on.
func (rf RangeFilter) GetField() string { return rf.field }

// IsOptional reports if the filter can be skipped at call time.
func (rf RangeFilter) IsOptional() bool { return rf.optional }

// Optional makes the filter skippable at call time.
func (rf RangeFilter) Optional() RangeFilter {
	rf.optional = true
	return rf
}

// Range filters a field by a lower and an upper bound.
func Range(field string) RangeFilter {
	return RangeFilter{field: field, optional: false}
}

// SearchFilter matches a field by a text pattern.
type SearchFilter struct {
	field    string
	optional bool
}

// marker method for sealed interface
func (sf SearchFilter) Filter() {}

// GetField returns the field the filter works on.
func (sf SearchFilter) GetField() string { return sf.field }

// IsOptional reports if the filter can be skipped at call time.
func (sf SearchFilter) IsOptional() bool { return sf.optional }

// Optional makes the filter skippable at call time.
func (sf SearchFilter) Optional() SearchFilter {
	sf.optional = true
	return sf
}

// Search filters a field by a text pattern.
func Search(field string) SearchFilter {
	return SearchFilter{field: field, optional: false}
}

// EqFilter matches a field by an exact value.
type EqFilter struct {
	field    string
	optional bool
}

// marker method for sealed interface
func (ef EqFilter) Filter() {}

// GetField returns the field the filter works on.
func (ef EqFilter) GetField() string { return ef.field }

// IsOptional reports if the filter can be skipped at call time.
func (ef EqFilter) IsOptional() bool { return ef.optional }

// Optional makes the filter skippable at call time.
func (ef EqFilter) Optional() EqFilter {
	ef.optional = true
	return ef
}

// Eq filters a field by an exact value.
func Eq(field string) EqFilter {
	return EqFilter{field: field, optional: false}
}
