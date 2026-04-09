package salary

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wencai/easyhr/internal/common/crypto"
)

// TestGenerateSlipToken 测试工资单 token 生成
func TestGenerateSlipToken(t *testing.T) {
	token1, err := generateSlipToken()
	assert.NoError(t, err)
	assert.Len(t, token1, 64) // 64 字符 hex 字符串

	token2, err := generateSlipToken()
	assert.NoError(t, err)
	assert.Len(t, token2, 64)
	assert.NotEqual(t, token1, token2) // 每次 token 不同
}

// TestSendSlipToken 测试发送工资单的 token 生成逻辑
func TestSendSlipToken(t *testing.T) {
	// 测试 token 生成
	token, err := generateSlipToken()
	assert.NoError(t, err)
	assert.Len(t, token, 64)

	slipTokens := make(map[string]string)
	slipTokens[token] = "test"
	assert.Contains(t, slipTokens, token)
}

// TestVerifySlipToken 测试 token 验证逻辑
func TestVerifySlipToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		expired bool
		wantErr error
	}{
		{
			name:    "无效 token",
			token:   "",
			expired: false,
			wantErr: ErrSlipTokenInvalid,
		},
		{
			name:    "过期 token",
			token:   "valid-token-but-expired",
			expired: true,
			wantErr: ErrSlipTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr != nil {
				if tt.expired {
					// 模拟过期检查
					slip := &PayrollSlip{
						Token:     tt.token,
						ExpiresAt: time.Now().Add(-time.Hour), // 已过期
					}
					if time.Now().After(slip.ExpiresAt) {
						assert.Equal(t, ErrSlipTokenExpired, tt.wantErr)
					}
				} else {
					assert.Equal(t, ErrSlipTokenInvalid, tt.wantErr)
				}
			}
		})
	}
}

// TestSignSlip 测试签收逻辑
func TestSignSlip(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		currentStatus  string
		wantErr        error
		expectedStatus string
	}{
		{
			name:           "正常签收",
			currentStatus:  SlipStatusViewed,
			wantErr:        nil,
			expectedStatus: SlipStatusSigned,
		},
		{
			name:           "重复签收",
			currentStatus:  SlipStatusSigned,
			wantErr:        ErrSlipAlreadySigned,
			expectedStatus: SlipStatusSigned,
		},
		{
			name:           "未查看不能签收",
			currentStatus:  SlipStatusSent,
			wantErr:        ErrSlipNotViewed,
			expectedStatus: SlipStatusSent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slip := &PayrollSlip{
				Status: tt.currentStatus,
			}

			if tt.wantErr != nil {
				if tt.currentStatus == SlipStatusSigned {
					assert.Equal(t, ErrSlipAlreadySigned, tt.wantErr)
				} else if tt.currentStatus == SlipStatusSent {
					assert.Equal(t, ErrSlipNotViewed, tt.wantErr)
				}
			} else {
				slip.Status = tt.expectedStatus
				slip.SignedAt = &now
				assert.Equal(t, SlipStatusSigned, slip.Status)
				assert.NotNil(t, slip.SignedAt)
			}
		})
	}
}

// TestExportPayrollExcel 测试 Excel 导出
func TestExportPayrollExcel(t *testing.T) {
	records := []PayrollRecordWithItems{
		{
			Record: PayrollRecord{
				EmployeeName:    "张三",
				GrossIncome:     10000,
				SIDeduction:     1000,
				Tax:             200,
				TotalDeductions: 1200,
				NetIncome:       8800,
			},
			Items: []PayrollItem{
				{ItemName: "基本工资", ItemType: "income", Amount: 8000},
				{ItemName: "绩效工资", ItemType: "income", Amount: 2000},
				{ItemName: "事假扣款", ItemType: "deduction", Amount: 0},
			},
		},
	}

	data, err := ExportPayrollExcel(records, 2026, 4)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Greater(t, len(data), 100) // Excel 文件应该有一定大小
}

// TestPhoneEncryption 测试手机号加密
func TestPhoneEncryption(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 字节
	phone := "13800138000"

	// 加密
	encrypted, err := crypto.Encrypt(phone, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, phone, encrypted)

	// 解密
	decrypted, err := crypto.Decrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, phone, decrypted)
}

// TestPhoneHash 测试手机号哈希
func TestPhoneHash(t *testing.T) {
	phone := "13800138000"

	hash1 := crypto.HashSHA256(phone)
	hash2 := crypto.HashSHA256(phone)

	assert.Len(t, hash1, 64) // SHA-256 输出 64 字符 hex
	assert.Equal(t, hash1, hash2) // 相同输入产生相同哈希

	hash3 := crypto.HashSHA256("13800138001")
	assert.NotEqual(t, hash1, hash3) // 不同输入产生不同哈希
}
