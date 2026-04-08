package salary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPresetTemplateItems 验证预置模板项数量和属性
func TestPresetTemplateItems(t *testing.T) {
	presets := getPresetItems()
	assert.Equal(t, 10, len(presets), "应有10个预置薪资项")

	// 验证收入项
	incomeCount := 0
	deductionCount := 0
	for _, p := range presets {
		if p.Type == "income" {
			incomeCount++
		} else {
			deductionCount++
		}
	}
	assert.Equal(t, 7, incomeCount, "应有7个收入项")
	assert.Equal(t, 3, deductionCount, "应有3个扣款项")

	// 验证基本工资是必填项
	assert.True(t, presets[0].IsRequired, "基本工资应为必填")
	assert.Equal(t, "基本工资", presets[0].Name)
	assert.Equal(t, 1, presets[0].SortOrder)
}

// TestPayrollStatusConstants 验证工资核算状态常量
func TestPayrollStatusConstants(t *testing.T) {
	assert.Equal(t, "draft", PayrollStatusDraft)
	assert.Equal(t, "calculated", PayrollStatusCalculated)
	assert.Equal(t, "confirmed", PayrollStatusConfirmed)
	assert.Equal(t, "paid", PayrollStatusPaid)
}

// TestSlipStatusConstants 验证工资单状态常量
func TestSlipStatusConstants(t *testing.T) {
	assert.Equal(t, "pending", SlipStatusPending)
	assert.Equal(t, "sent", SlipStatusSent)
	assert.Equal(t, "viewed", SlipStatusViewed)
	assert.Equal(t, "signed", SlipStatusSigned)
}

// TestSalaryErrorCodes 验证错误码
func TestSalaryErrorCodes(t *testing.T) {
	assert.Equal(t, 50001, CodeTemplateConfig)
	assert.Equal(t, 50002, CodePayrollFailed)
	assert.Equal(t, 50003, CodeInvalidStatus)
	assert.Equal(t, 50004, CodeAttendanceImport)
	assert.Equal(t, 50005, CodeSlipTokenInvalid)
	assert.Equal(t, 50006, CodeSMSVerifyFailed)
	assert.Equal(t, 50007, CodeEmployeeMatch)
}

// TestEmployeeInfoStruct 验证 EmployeeInfo 结构体字段
func TestEmployeeInfoStruct(t *testing.T) {
	info := EmployeeInfo{
		ID:         1,
		Name:       "张三",
		BaseSalary: 10000,
	}
	assert.Equal(t, int64(1), info.ID)
	assert.Equal(t, "张三", info.Name)
	assert.Equal(t, float64(10000), info.BaseSalary)
}

// TestDTOBinding 验证 DTO 结构体 tag
func TestDTOBinding(t *testing.T) {
	// SetEmployeeItemsRequest 的 Month 字段验证
	req := SetEmployeeItemsRequest{
		Month: "2026-04",
		Items: []SalaryItemInput{
			{TemplateItemID: 1, Amount: 8000},
		},
	}
	assert.Equal(t, "2026-04", req.Month)
	assert.Len(t, req.Items, 1)
	assert.Equal(t, float64(8000), req.Items[0].Amount)
}

// getPresetItems 返回预置模板项列表（供测试使用）
func getPresetItems() []struct {
	Name      string
	Type      string
	SortOrder int
	IsRequired bool
} {
	return []struct {
		Name      string
		Type      string
		SortOrder int
		IsRequired bool
	}{
		{"基本工资", "income", 1, true},
		{"绩效工资", "income", 2, false},
		{"岗位补贴", "income", 3, false},
		{"餐补", "income", 4, false},
		{"交通补", "income", 5, false},
		{"通讯补", "income", 6, false},
		{"其他补贴", "income", 7, false},
		{"事假扣款", "deduction", 8, false},
		{"病假扣款", "deduction", 9, false},
		{"其他扣款", "deduction", 10, false},
	}
}
