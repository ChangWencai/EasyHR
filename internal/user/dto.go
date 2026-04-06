package user

type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
}

type LoginRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CompleteOnboardingRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=100"`
	CreditCode   string `json:"credit_code" binding:"required,len=18"`
	City         string `json:"city" binding:"required"`
	ContactName  string `json:"contact_name" binding:"required"`
	ContactPhone string `json:"contact_phone" binding:"required,len=11"`
}

type CreateSubAccountRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Name  string `json:"name" binding:"required,min=2,max=50"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

type LoginResponse struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	OnboardingRequired bool   `json:"onboarding_required"`
}

type UserInfoResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Role    string `json:"role"`
	OrgID   int64  `json:"org_id"`
	OrgName string `json:"org_name"`
}
