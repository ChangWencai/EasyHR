package finance

// ClosingValidationResponse is the result of validating a period before closing.
type ClosingValidationResponse struct {
	CanClose bool     `json:"can_close"`
	Errors   []string `json:"errors,omitempty"`
}

// ClosePeriodRequest is the request body for closing a period.
type ClosePeriodRequest struct {
	PeriodID int64 `json:"period_id" binding:"required"`
}

// RevertPeriodRequest is the request body for reverting a closed period.
type RevertPeriodRequest struct {
	PeriodID int64 `json:"period_id" binding:"required"`
	Confirm  bool  `json:"confirm"` // Must be true to confirm the revert action
}

// PeriodListResponse is the response for listing periods.
type PeriodListResponse struct {
	Periods []PeriodItem `json:"periods"`
}

// PeriodItem is a single period in the list.
type PeriodItem struct {
	ID     int64        `json:"id"`
	Year   int          `json:"year"`
	Month  int          `json:"month"`
	Status PeriodStatus `json:"status"`
}
