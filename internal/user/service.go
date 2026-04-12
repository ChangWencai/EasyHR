package user

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/pkg/jwt"
	"github.com/wencai/easyhr/pkg/sms"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	isTest := s.sms != nil && s.sms.IsTestMode()

	if !isTest {
		exists, err := s.rdb.Exists(ctx, limitKey).Result()
		if err != nil {
			return fmt.Errorf("check limit: %w", err)
		}
		if exists > 0 {
			return fmt.Errorf("请5分钟后再次发送")
		}
	}

	code := "123456"
	if !isTest {
		code = fmt.Sprintf("%06d", rand.Intn(1000000))
	}
	codeKey := "sms:code:" + phone
	if err := s.rdb.Set(ctx, codeKey, code, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("store code: %w", err)
	}

	if !isTest {
		if err := s.rdb.Set(ctx, limitKey, "1", 60*time.Second).Err(); err != nil {
			return fmt.Errorf("set limit: %w", err)
		}
		if err := s.sms.SendCode(ctx, phone, code); err != nil {
			s.rdb.Del(ctx, codeKey, limitKey)
			return fmt.Errorf("send sms: %w", err)
		}
	}
	return nil
}

// Register 仅创建用户（不创建企业），企业信息在 onboarding 阶段录入
func (s *Service) Register(ctx context.Context, phone, code string) (*LoginResponse, error) {
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
	existing, err := s.repo.FindByPhoneHash(phoneHash)
	if err != nil && !errors.Is(err, redis.Nil) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if err == nil && existing != nil {
		return nil, fmt.Errorf("该手机号已注册，请直接登录")
	}

	aesKey := []byte(s.crypto.AESKey)
	encryptedPhone, _ := crypto.Encrypt(phone, aesKey)
	newUser := &model.User{
		BaseModel: model.BaseModel{OrgID: 0}, // onboarding 阶段再关联真实 org_id
		Phone:     encryptedPhone,
		PhoneHash: phoneHash,
		Role:      "owner",
		Status:    "active",
	}
	if err := s.repo.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("创建账号失败: %w", err)
	}
	accessToken, _ := jwt.GenerateAccessToken(newUser.ID, newUser.OrgID, newUser.Role, s.jwtCfg.Secret, s.jwtCfg.AccessTTL)
	refreshToken, _ := jwt.GenerateRefreshToken(newUser.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshTTL)
	return &LoginResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		OnboardingRequired: true,
	}, nil
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
	if err == redis.Nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("该手机号未注册，请先注册")
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	// MEMBER 角色拒绝登录（per D-06, D-20）
	if user.Role == "member" {
		return nil, fmt.Errorf("MEMBER_ROLE_FORBIDDEN")
	}

	accessToken, _ := jwt.GenerateAccessToken(user.ID, user.OrgID, user.Role, s.jwtCfg.Secret, s.jwtCfg.AccessTTL)
	refreshToken, _ := jwt.GenerateRefreshToken(user.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshTTL)

	onboardingRequired := false
	if user.OrgID > 0 {
		org, _ := s.repo.FindByOrgID(user.OrgID)
		if org != nil && org.Status == "inactive" {
			onboardingRequired = true
		}
	} else {
		onboardingRequired = true
	}

	return &LoginResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		OnboardingRequired: onboardingRequired,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	claims, err := jwt.ParseRefreshToken(refreshToken, s.jwtCfg.Secret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 查询用户最新 orgID 和 role（refresh token 本身只含 userID，无法携带 org 信息）
	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	blacklistFunc := func(jti string, ttl time.Duration) error {
		return s.rdb.Set(ctx, "token:blacklist:"+jti, "1", ttl).Err()
	}
	return jwt.RefreshTokens(refreshToken, s.jwtCfg.Secret, s.jwtCfg.AccessTTL, s.jwtCfg.RefreshTTL, blacklistFunc, user.ID, user.OrgID, user.Role)
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

// CompleteOnboarding 创建企业并关联到用户，返回含新 token 的响应
func (s *Service) CompleteOnboarding(ctx context.Context, userID int64, req *CompleteOnboardingRequest) (*LoginResponse, error) {
	encryptedPhone, _ := crypto.Encrypt(req.ContactPhone, []byte(s.crypto.AESKey))
	org := &model.Organization{
		Name:         req.Name,
		CreditCode:   req.CreditCode,
		City:         req.City,
		ContactName:  req.ContactName,
		ContactPhone: encryptedPhone,
		Status:       "active",
	}
	if err := s.repo.CreateOrg(org); err != nil {
		return nil, fmt.Errorf("创建企业失败: %w", err)
	}
	if err := s.repo.UpdateUserOrgID(userID, org.ID); err != nil {
		return nil, fmt.Errorf("关联企业失败: %w", err)
	}

	// 生成含正确 org_id 的新 token（避免用户继续使用旧 token）
	accessToken, _ := jwt.GenerateAccessToken(userID, org.ID, "owner", s.jwtCfg.Secret, s.jwtCfg.AccessTTL)
	refreshToken, _ := jwt.GenerateRefreshToken(userID, s.jwtCfg.Secret, s.jwtCfg.RefreshTTL)

	return &LoginResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		OnboardingRequired: false,
	}, nil
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

// RegisterPassword 设置用户密码（bcrypt 哈希）
func (s *Service) RegisterPassword(ctx context.Context, phone, password string) error {
	phoneHash := crypto.HashSHA256(phone)
	user, err := s.repo.FindByPhoneHash(phoneHash)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	return s.repo.UpdateUserPassword(user.ID, string(hashed))
}

// LoginPassword 手机号+密码登录
func (s *Service) LoginPassword(ctx context.Context, phone, password string) (*LoginResponse, error) {
	phoneHash := crypto.HashSHA256(phone)
	user, err := s.repo.FindByPhoneHash(phoneHash)
	if err != nil {
		return nil, fmt.Errorf("手机号或密码错误")
	}

	// MEMBER 角色拒绝登录（per D-06, D-20）
	if user.Role == "member" {
		return nil, fmt.Errorf("MEMBER_ROLE_FORBIDDEN")
	}

	// 密码校验
	if user.PasswordHash == "" {
		return nil, fmt.Errorf("该账号未设置密码，请使用手机验证码登录")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("手机号或密码错误")
	}

	// 生成 token
	accessToken, _ := jwt.GenerateAccessToken(user.ID, user.OrgID, user.Role, s.jwtCfg.Secret, s.jwtCfg.AccessTTL)
	refreshToken, _ := jwt.GenerateRefreshToken(user.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshTTL)

	// 判断 onboarding 状态
	onboardingRequired := false
	if user.OrgID > 0 {
		org, _ := s.repo.FindByOrgID(user.OrgID)
		if org != nil && org.Status == "inactive" {
			onboardingRequired = true
		}
	} else {
		onboardingRequired = true
	}

	return &LoginResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		OnboardingRequired: onboardingRequired,
	}, nil
}

// GetMe 获取当前用户信息（含企业信息）
func (s *Service) GetMe(ctx context.Context, userID, orgID int64) (*MeResponse, error) {
	user, err := s.repo.FindUserByID(orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	resp := &MeResponse{
		ID:     user.ID,
		Name:   user.Name,
		Role:   user.Role,
		Avatar: user.Avatar,
	}

	if orgID > 0 {
		org, err := s.repo.FindByOrgID(orgID)
		if err == nil && org != nil {
			resp.Org = &OrgInfo{
				ID:           org.ID,
				Name:         org.Name,
				CreditCode:   org.CreditCode,
				City:         org.City,
				ContactName:  org.ContactName,
				ContactPhone: org.ContactPhone,
			}
			if org.Status == "inactive" {
				resp.OnboardingRequired = true
			}
		}
	} else {
		resp.OnboardingRequired = true
	}

	return resp, nil
}

// ChangePassword 修改用户密码
func (s *Service) ChangePassword(ctx context.Context, userID int64, req *ChangePasswordRequest) error {
	// 先根据 userID 查找用户（跨租户查询）
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 校验旧密码
	if user.PasswordHash == "" {
		return fmt.Errorf("该账号未设置密码，请使用手机验证码登录")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("旧密码错误")
	}

	// 生成新密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	return s.repo.UpdateUserPassword(user.ID, string(hashed))
}

// UpdateOrg 更新企业信息
func (s *Service) UpdateOrg(ctx context.Context, orgID int64, req *UpdateOrgRequest) error {
	encryptedPhone, _ := crypto.Encrypt(req.ContactPhone, []byte(s.crypto.AESKey))
	return s.repo.UpdateOrg(orgID, map[string]interface{}{
		"name":          req.Name,
		"credit_code":  req.CreditCode,
		"city":          req.City,
		"contact_name":  req.ContactName,
		"contact_phone": encryptedPhone,
	})
}

// UpdateAvatar 更新用户头像
func (s *Service) UpdateAvatar(ctx context.Context, userID int64, avatar string) error {
	return s.repo.UpdateUserAvatar(userID, avatar)
}

// UpdateName 更新用户姓名
func (s *Service) UpdateName(ctx context.Context, userID int64, name string) error {
	return s.repo.UpdateUserName(userID, name)
}
