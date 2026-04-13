package user

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/pkg/jwt"
	"github.com/wencai/easyhr/pkg/sms"
)

type Service struct {
	repo   *Repository
	rdb    *redis.Client
	sms    *sms.Client
	jwtCfg config.JWTConfig
	crypto config.CryptoConfig
}

func NewService(repo *Repository, rdb *redis.Client, smsClient *sms.Client, jwtCfg config.JWTConfig, cryptoCfg config.CryptoConfig) *Service {
	return &Service{
		repo:   repo,
		rdb:    rdb,
		sms:    smsClient,
		jwtCfg: jwtCfg,
		crypto: cryptoCfg,
	}
}

func (s *Service) SendCode(ctx context.Context, phone string) error {
	limitKey := "sms:limit:" + phone
	exists, err := s.rdb.Exists(ctx, limitKey).Result()
	if err != nil {
		return fmt.Errorf("check limit: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("请60秒后再次发送")
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	// 开发模式：使用固定验证码方便调试
	code = "123456"
	codeKey := "sms:code:" + phone
	if err := s.rdb.Set(ctx, codeKey, code, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("store code: %w", err)
	}
	if err := s.rdb.Set(ctx, limitKey, "1", 60*time.Second).Err(); err != nil {
		return fmt.Errorf("set limit: %w", err)
	}
	if err := s.sms.SendCode(ctx, phone, code); err != nil {
		s.rdb.Del(ctx, codeKey, limitKey)
		return fmt.Errorf("send sms: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, phone, code string) (*LoginResponse, error) {
	codeKey := "sms:code:" + phone
	storedCode, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("验证码已过期")
	}
	if err != nil {
		return nil, fmt.Errorf("get code: %w", err)
	}

	errorKey := "sms:errors:" + phone
	errorCount, _ := s.rdb.Get(ctx, errorKey).Int()
	if errorCount >= 5 {
		s.rdb.Del(ctx, codeKey, errorKey)
		return nil, fmt.Errorf("验证码错误次数过多，请重新获取")
	}

	if storedCode != code {
		s.rdb.Incr(ctx, errorKey)
		s.rdb.Expire(ctx, errorKey, 5*time.Minute)
		return nil, fmt.Errorf("验证码错误")
	}

	s.rdb.Del(ctx, codeKey, errorKey)

	phoneHash := crypto.HashSHA256(phone)
	user, err := s.repo.FindByPhoneHash(phoneHash)
	if err == redis.Nil {
		org := &model.Organization{
			Name:   "",
			Status: "inactive",
		}
		aesKey := []byte(s.crypto.AESKey)
		encryptedPhone, _ := crypto.Encrypt(phone, aesKey)
		newUser := &model.User{}
		newUser.OrgID = org.ID
		newUser.Phone = encryptedPhone
		newUser.PhoneHash = phoneHash
		newUser.Role = "owner"
		newUser.Status = "active"
		if err := s.repo.CreateOrgAndOwner(org, newUser); err != nil {
			return nil, fmt.Errorf("auto register: %w", err)
		}
		accessToken, _ := jwt.GenerateAccessToken(newUser.ID, newUser.OrgID, newUser.Role, s.jwtCfg.Secret, s.jwtCfg.AccessTTL)
		refreshToken, _ := jwt.GenerateRefreshToken(newUser.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshTTL)
		return &LoginResponse{
			AccessToken:        accessToken,
			RefreshToken:       refreshToken,
			OnboardingRequired: true,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	accessToken, _ := jwt.GenerateAccessToken(user.ID, user.OrgID, user.Role, s.jwtCfg.Secret, s.jwtCfg.AccessTTL)
	refreshToken, _ := jwt.GenerateRefreshToken(user.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshTTL)

	onboardingRequired := false
	org, _ := s.repo.FindByOrgID(user.OrgID)
	if org != nil && org.Status == "inactive" {
		onboardingRequired = true
	}

	return &LoginResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		OnboardingRequired: onboardingRequired,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	blacklistFunc := func(jti string, ttl time.Duration) error {
		return s.rdb.Set(ctx, "token:blacklist:"+jti, "1", ttl).Err()
	}
	return jwt.RefreshTokens(refreshToken, s.jwtCfg.Secret, s.jwtCfg.AccessTTL, s.jwtCfg.RefreshTTL, blacklistFunc)
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	claims, err := jwt.ParseToken(accessToken, s.jwtCfg.Secret)
	if err == nil {
		s.rdb.Set(ctx, "token:blacklist:"+claims.ID, "1", s.jwtCfg.AccessTTL)
	}
	if refreshToken != "" {
		rClaims, err := jwt.ParseRefreshToken(refreshToken, s.jwtCfg.Secret)
		if err == nil {
			s.rdb.Set(ctx, "token:blacklist:"+rClaims.ID, "1", s.jwtCfg.RefreshTTL)
		}
	}
	return nil
}

func (s *Service) CompleteOnboarding(ctx context.Context, orgID int64, req *CompleteOnboardingRequest) error {
	aesKey := []byte(s.crypto.AESKey)
	encryptedPhone, _ := crypto.Encrypt(req.ContactPhone, aesKey)
	updates := map[string]interface{}{
		"name":          req.Name,
		"credit_code":   req.CreditCode,
		"city":          req.City,
		"contact_name":  req.ContactName,
		"contact_phone": encryptedPhone,
		"status":        "active",
	}
	return s.repo.UpdateOrg(orgID, updates)
}

func (s *Service) CreateSubAccount(ctx context.Context, orgID int64, req *CreateSubAccountRequest) error {
	if req.Role == "owner" {
		return fmt.Errorf("不可创建owner角色")
	}
	aesKey := []byte(s.crypto.AESKey)
	phoneHash := crypto.HashSHA256(req.Phone)
	existing, _ := s.repo.FindByPhoneHash(phoneHash)
	if existing != nil && existing.OrgID == orgID {
		return fmt.Errorf("该手机号已存在于企业中")
	}
	encryptedPhone, _ := crypto.Encrypt(req.Phone, aesKey)
	user := &model.User{
		Phone:     encryptedPhone,
		PhoneHash: phoneHash,
		Name:      req.Name,
		Role:      req.Role,
		Status:    "active",
	}
	user.OrgID = orgID
	return s.repo.CreateUser(user)
}

func (s *Service) ListSubAccounts(ctx context.Context, orgID int64, page, pageSize int) ([]UserInfoResponse, int64, error) {
	users, total, err := s.repo.ListUsers(orgID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	aesKey := []byte(s.crypto.AESKey)
	org, _ := s.repo.FindByOrgID(orgID)
	orgName := ""
	if org != nil {
		orgName = org.Name
	}
	var result []UserInfoResponse
	for _, u := range users {
		decryptedPhone, _ := crypto.Decrypt(u.Phone, aesKey)
		result = append(result, UserInfoResponse{
			ID:      u.ID,
			Name:    u.Name,
			Phone:   crypto.MaskPhone(decryptedPhone),
			Role:    u.Role,
			OrgID:   u.OrgID,
			OrgName: orgName,
		})
	}
	return result, total, nil
}

func (s *Service) UpdateSubAccountRole(ctx context.Context, orgID, targetUserID int64, role string) error {
	return s.repo.UpdateUserRole(orgID, targetUserID, role)
}

func (s *Service) DeleteSubAccount(ctx context.Context, orgID, targetUserID int64) error {
	return s.repo.DeleteUser(orgID, targetUserID)
}
