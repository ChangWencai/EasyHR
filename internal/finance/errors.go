package finance

import "fmt"

// Finance module error codes in the 60xxx range (D-34).
const (
	CodeVoucherUnbalanced   = 60201
	CodeVoucherAudited       = 60202
	CodePeriodClosed         = 60203
	CodeClosingValidation    = 60204
	CodeVoucherNotFound      = 60205
	CodePeriodNotFound       = 60206
	CodeAccountNotFound      = 60207
	CodeInvalidStatus        = 60208
	CodeInvalidDC            = 60209
)

// FinanceError wraps an error with a numeric code for structured error handling.
type FinanceError struct {
	Code int
	Err  error
}

func (e *FinanceError) Error() string {
	return e.Err.Error()
}

func (e *FinanceError) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with a finance error code.
func WrapError(code int, err error) *FinanceError {
	return &FinanceError{Code: code, Err: err}
}

// Pre-defined errors.

var (
	// ErrVoucherUnbalanced (60201): 借贷不平衡，借方合计不等于贷方合计
	ErrVoucherUnbalanced = &FinanceError{
		Code: CodeVoucherUnbalanced,
		Err:  fmt.Errorf("借贷不平衡：借方合计不等于贷方合计"),
	}

	// ErrVoucherAudited (60202): 凭证已审核，禁止修改或删除
	ErrVoucherAudited = &FinanceError{
		Code: CodeVoucherAudited,
		Err:  fmt.Errorf("凭证已审核，禁止修改"),
	}

	// ErrPeriodClosed (60203): 会计期间已结账，禁止操作
	ErrPeriodClosed = &FinanceError{
		Code: CodePeriodClosed,
		Err:  fmt.Errorf("会计期间已结账，禁止操作"),
	}

	// ErrClosingValidation (60204): 结账校验失败
	ErrClosingValidation = &FinanceError{
		Code: CodeClosingValidation,
		Err:  fmt.Errorf("结账校验失败"),
	}
)
