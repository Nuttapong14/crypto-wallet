package repositories

// ListOptions captures common pagination and sorting parameters for repository queries.
type ListOptions struct {
	Limit     int
	Offset    int
	SortBy    string
	SortOrder SortOrder
}

// SortOrder defines the ordering direction for list queries.
type SortOrder string

const (
	SortAscending  SortOrder = "ASC"
	SortDescending SortOrder = "DESC"
)

// WithDefaults applies default pagination when unset.
func (o ListOptions) WithDefaults() ListOptions {
	result := o
	if result.Limit <= 0 {
		result.Limit = 50
	}
	if result.SortBy == "" {
		result.SortBy = "created_at"
	}
	if result.SortOrder == "" {
		result.SortOrder = SortDescending
	}
	if result.Offset < 0 {
		result.Offset = 0
	}
	return result
}
