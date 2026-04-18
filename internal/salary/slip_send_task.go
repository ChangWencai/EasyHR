package salary

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// asynq task type
const TypeSlipSend = "salary:slip:send"

// SlipSendPayload asynq 任务载荷
type SlipSendPayload struct {
	OrgID       int64   `json:"org_id"`
	UserID      int64   `json:"user_id"`
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	EmployeeIDs []int64 `json:"employee_ids"` // empty = all employees
	Channel     string  `json:"channel"`       // miniapp/sms/h5
}

// NewSlipSendTask 创建工资条发送任务
func NewSlipSendTask(payload *SlipSendPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal slip send payload: %w", err)
	}
	return asynq.NewTask(TypeSlipSend, data), nil
}
