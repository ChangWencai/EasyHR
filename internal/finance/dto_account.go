package finance

// AccountTreeResponse is the API response format for the account tree.
type AccountTreeResponse struct {
	ID            int64                  `json:"id"`
	Code          string                 `json:"code"`
	Name          string                 `json:"name"`
	Category      AccountCategory        `json:"category"`
	ParentID      *int64                `json:"parent_id,omitempty"`
	Level         int                    `json:"level"`
	NormalBalance NormalBalance          `json:"normal_balance"`
	IsActive      bool                   `json:"is_active"`
	IsSystem      bool                   `json:"is_system"`
	Children      []*AccountTreeResponse `json:"children,omitempty"`
}

// CreateAccountRequest is the request body for creating a custom account.
// Custom accounts must have code starting with 8xxxx (D-08).
type CreateAccountRequest struct {
	Code     string          `json:"code" binding:"required,min=1,max=20"`
	Name     string          `json:"name" binding:"required,min=1,max=100"`
	Category AccountCategory `json:"category" binding:"required"`
	ParentID *int64         `json:"parent_id"`
}

// UpdateAccountRequest is the request body for updating an account.
// Only name and is_active can be updated; system accounts cannot be deactivated.
type UpdateAccountRequest struct {
	Name     string `json:"name" binding:"omitempty,min=1,max=100"`
	IsActive *bool  `json:"is_active"`
}
