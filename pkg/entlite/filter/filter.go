package filter

type Filter interface {
	Filter()
	GetField() string
	IsOptional() bool
}

type RangeFilter struct {
	field    string
	optional bool
}

func (rf RangeFilter) Filter()          {}
func (rf RangeFilter) GetField() string { return rf.field }
func (rf RangeFilter) IsOptional() bool { return rf.optional }

func (rf RangeFilter) Optional() RangeFilter {
	rf.optional = true
	return rf
}

func Range(field string) RangeFilter {
	return RangeFilter{field: field, optional: false}
}

type SearchFilter struct {
	field    string
	optional bool
}

func (sf SearchFilter) Filter()          {}
func (sf SearchFilter) GetField() string { return sf.field }
func (sf SearchFilter) IsOptional() bool { return sf.optional }

func (sf SearchFilter) Optional() SearchFilter {
	sf.optional = true
	return sf
}

func Search(field string) SearchFilter {
	return SearchFilter{field: field, optional: false}
}

type EqFilter struct {
	field    string
	optional bool
}

func (ef EqFilter) Filter()          {}
func (ef EqFilter) GetField() string { return ef.field }
func (ef EqFilter) IsOptional() bool { return ef.optional }

func (ef EqFilter) Optional() EqFilter {
	ef.optional = true
	return ef
}

func Eq(field string) EqFilter {
	return EqFilter{field: field, optional: false}
}
