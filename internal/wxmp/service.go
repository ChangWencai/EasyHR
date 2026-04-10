package wxmp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/common/crypto"
)

// WXMPService 微信小程序业务逻辑层
type WXMPService struct {
	repo       WXMPRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	rdb        *redis.Client
	cryptoKey  []byte
}

// NewWXMPService 创建 WXMPService
func NewWXMPService(repo WXMPRepository, jwtSecret string, accessTTL, refreshTTL time.Duration, rdb *redis.Client, cryptoKey string) *WXMPService {
	return &WXMPService{
		repo:       repo,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		rdb:        rdb,
		cryptoKey:  []byte(cryptoKey),
	}
}

// LoginMember 员工手机号+验证码登录
func (s *WXMPService) LoginMember(ctx context.Context, phone, code string) (*LoginResponse, error) {
	// 1. 验证短信验证码
	codeKey := "sms:code:" + phone
	storedCode, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return nil, errors.New("验证码已过期")
	}
	if err != nil {
		return nil, fmt.Errorf("验证短信: %w", err)
	}
	if storedCode != code {
		return nil, errors.New("验证码错误")
	}

	// 验证通过，清除验证码
	s.rdb.Del(ctx, codeKey)

	// 2. 查找会员
	phoneHash := crypto.HashSHA256(phone)
	member, err := s.repo.GetMemberByPhone(ctx, phoneHash)
	if err != nil {
		return nil, errors.New("该手机号未关联员工账号")
	}

	// 3. 生成 JWT
	accessToken, err := s.generateAccessToken(member.UserID, member.EmployeeID, member.OrgID, member.Role)
	if err != nil {
		return nil, fmt.Errorf("生成token: %w", err)
	}
	refreshToken, err := s.generateRefreshToken(member.UserID)
	if err != nil {
		return nil, fmt.Errorf("生成refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTTL.Seconds()),
		Member:       *member,
	}, nil
}

// generateAccessToken 生成小程序访问令牌
func (s *WXMPService) generateAccessToken(userID, employeeID, orgID uint, role string) (string, error) {
	claims := WXMPClaims{
		UserID:     userID,
		EmployeeID: employeeID,
		OrgID:      orgID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// generateRefreshToken 生成刷新令牌
func (s *WXMPService) generateRefreshToken(userID uint) (string, error) {
	claims := WXMPRefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// WXMPClaims 小程序 JWT Claims
type WXMPClaims struct {
	UserID     uint   `json:"user_id"`
	EmployeeID uint   `json:"employee_id"`
	OrgID      uint   `json:"org_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// WXMPRefreshClaims 小程序刷新令牌 Claims
type WXMPRefreshClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// BindWechat 绑定微信 openid 到用户
func (s *WXMPService) BindWechat(ctx context.Context, userID uint, openID string) error {
	return s.repo.BindWechatOpenID(ctx, userID, openID)
}

// GetPayslips 获取工资条列表
func (s *WXMPService) GetPayslips(ctx context.Context, orgID, employeeID uint) ([]PayslipSummary, error) {
	return s.repo.ListPayslips(ctx, orgID, employeeID)
}

// GetPayslipDetail 获取工资条明细（需要短信验证）
func (s *WXMPService) GetPayslipDetail(ctx context.Context, orgID, employeeID, payslipID uint, verifyToken string) (*PayslipDetail, error) {
	// 验证 verifyToken
	if verifyToken == "" {
		return nil, errors.New("请先验证身份")
	}

	payslipKey := fmt.Sprintf("wxmp:verify:%d:%d", orgID, payslipID)
	storedToken, err := s.rdb.Get(ctx, payslipKey).Result()
	if err == redis.Nil || storedToken != verifyToken {
		return nil, errors.New("验证已过期，请重新验证")
	}
	if err != nil {
		return nil, fmt.Errorf("验证token检查: %w", err)
	}

	return s.repo.GetPayslipByID(ctx, orgID, employeeID, payslipID)
}

// VerifyPayslipAccess 短信验证工资条访问权限
func (s *WXMPService) VerifyPayslipAccess(ctx context.Context, orgID, employeeID, payslipID uint, code, phone string) (*VerifyPayslipResponse, error) {
	// 验证短信验证码
	codeKey := "sms:code:" + phone
	storedCode, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return nil, errors.New("验证码已过期")
	}
	if err != nil {
		return nil, fmt.Errorf("验证短信: %w", err)
	}
	if storedCode != code {
		return nil, errors.New("验证码错误")
	}
	s.rdb.Del(ctx, codeKey)

	// 生成短效 verify_token
	verifyToken := uuid.New().String()
	payslipKey := fmt.Sprintf("wxmp:verify:%d:%d", orgID, payslipID)
	// 15分钟有效期
	if err := s.rdb.Set(ctx, payslipKey, verifyToken, 15*time.Minute).Err(); err != nil {
		return nil, fmt.Errorf("存储verify token: %w", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute).Format("2006-01-02 15:04:05")
	return &VerifyPayslipResponse{
		VerifyToken: verifyToken,
		ExpiresAt:   expiresAt,
	}, nil
}

// SignPayslip 签收工资条
func (s *WXMPService) SignPayslip(ctx context.Context, orgID, employeeID, payslipID uint) error {
	return s.repo.SignPayslip(ctx, orgID, employeeID, payslipID)
}

// GetContracts 获取合同列表
func (s *WXMPService) GetContracts(ctx context.Context, orgID, employeeID uint) ([]ContractDTO, error) {
	return s.repo.ListContracts(ctx, orgID, employeeID)
}

// GetContractPDF 获取合同 PDF URL
func (s *WXMPService) GetContractPDF(ctx context.Context, orgID, employeeID, contractID uint) (string, error) {
	contract, err := s.repo.GetContractByID(ctx, orgID, employeeID, contractID)
	if err != nil {
		return "", err
	}
	return contract.PDFURL, nil
}

// GetSocialInsurance 获取社保记录
func (s *WXMPService) GetSocialInsurance(ctx context.Context, orgID, employeeID uint) ([]SocialInsuranceDTO, error) {
	return s.repo.ListSocialInsurance(ctx, orgID, employeeID)
}

// CreateExpense 创建报销单
func (s *WXMPService) CreateExpense(ctx context.Context, orgID, employeeID uint, req *ExpenseRequest) (*ExpenseDTO, error) {
	return s.repo.CreateExpense(ctx, orgID, employeeID, req)
}

// GetExpenses 获取报销单列表
func (s *WXMPService) GetExpenses(ctx context.Context, orgID, employeeID uint) ([]ExpenseDTO, error) {
	return s.repo.ListExpenses(ctx, orgID, employeeID)
}

// GetExpenseDetail 获取报销单详情
func (s *WXMPService) GetExpenseDetail(ctx context.Context, orgID, employeeID, expenseID uint) (*ExpenseDTO, error) {
	return s.repo.GetExpenseByID(ctx, orgID, employeeID, expenseID)
}

// WXMPRepository 接口（由 repository.go 实现）
// 已在 model.go 中定义
