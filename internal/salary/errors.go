package salary

import "errors"

// 工资模块错误码 50xxx
var (
	ErrTemplateConfig   = errors.New("薪资模板配置错误")        // 50001
	ErrPayrollFailed    = errors.New("工资核算失败")            // 50002
	ErrInvalidStatus    = errors.New("工资表状态不允许操作")      // 50003
	ErrAttendanceImport = errors.New("考勤导入失败")            // 50004
	ErrSlipTokenInvalid = errors.New("工资单 token 无效或过期") // 50005
	ErrSMSVerifyFailed  = errors.New("短信验证失败")            // 50006
	ErrEmployeeMatch    = errors.New("员工匹配失败")            // 50007
	ErrPayrollNotFound  = errors.New("工资记录不存在")          // 50008
)

// 错误码映射
const (
	CodeTemplateConfig   = 50001
	CodePayrollFailed    = 50002
	CodeInvalidStatus    = 50003
	CodeAttendanceImport = 50004
	CodeSlipTokenInvalid = 50005
	CodeSMSVerifyFailed  = 50006
	CodeEmployeeMatch    = 50007
	CodePayrollNotFound  = 50008
)

// WrapError 包装错误并附加错误码
func WrapError(code int, err error) error {
	return &SalaryError{Code: code, Err: err}
}

// SalaryError 工资模块错误
type SalaryError struct {
	Code int
	Err  error
}

func (e *SalaryError) Error() string {
	return e.Err.Error()
}
