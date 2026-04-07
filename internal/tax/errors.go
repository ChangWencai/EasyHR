package tax

import "fmt"

// 错误定义 (错误码范围: 40xxx)

// ErrBracketNotFound 税率表未找到
var ErrBracketNotFound = fmt.Errorf("税率表未找到")

// ErrDeductionNotFound 专项附加扣除记录未找到
var ErrDeductionNotFound = fmt.Errorf("专项附加扣除记录未找到")

// ErrTaxRecordNotFound 个税计算记录未找到
var ErrTaxRecordNotFound = fmt.Errorf("个税计算记录未找到")

// ErrDeclarationNotFound 个税申报记录未找到
var ErrDeclarationNotFound = fmt.Errorf("个税申报记录未找到")

// ErrDuplicateDeduction 该员工已有同类型的专项附加扣除记录
var ErrDuplicateDeduction = fmt.Errorf("该员工已有同类型的专项附加扣除记录")

// ErrMutuallyExclusiveDeduction 住房贷款利息和住房租金不可同时享受
var ErrMutuallyExclusiveDeduction = fmt.Errorf("住房贷款利息和住房租金不可同时享受")

// ErrInvalidDeductionType 无效的扣除类型
var ErrInvalidDeductionType = fmt.Errorf("无效的专项附加扣除类型")

// ErrorCodeMap 错误码映射
var ErrorCodeMap = map[error]int{
	ErrBracketNotFound:             40001,
	ErrDeductionNotFound:           40002,
	ErrTaxRecordNotFound:           40003,
	ErrDeclarationNotFound:         40004,
	ErrDuplicateDeduction:          40005,
	ErrMutuallyExclusiveDeduction:  40006,
	ErrInvalidDeductionType:        40007,
}
