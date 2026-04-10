package finance

// Common types used across finance DTOs.

// ListResponse is a generic paginated list response.
type ListResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// ErrorCode extracts the finance error code from an error, returning 0 if not a FinanceError.
func ErrorCode(err error) int {
	if fe, ok := err.(*FinanceError); ok {
		return fe.Code
	}
	return 0
}
