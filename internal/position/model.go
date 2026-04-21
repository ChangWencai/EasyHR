package position

import (
	"github.com/wencai/easyhr/internal/common/model"
)

// Position 岗位模型（独立岗位管理，支持跨部门复用）
// department_id 为 NULL 时表示通用岗位（任何部门可用）
type Position struct {
	model.BaseModel
	Name         string `gorm:"column:name;type:varchar(100);not null;index;comment:岗位名称" json:"name"`
	DepartmentID *int64 `gorm:"column:department_id;index;comment:所属部门（NULL=通用岗位）" json:"department_id"`
	SortOrder    int    `gorm:"column:sort_order;not null;default:0;comment:排序" json:"sort_order"`
}

// TableName 指定表名
func (Position) TableName() string {
	return "positions"
}
