package department

import (
	"github.com/wencai/easyhr/internal/common/model"
)

// Department 部门模型（邻接表模型，parent_id 自引用）
type Department struct {
	model.BaseModel
	Name      string `gorm:"column:name;type:varchar(100);not null;index;comment:部门名称" json:"name"`
	ParentID  *int64 `gorm:"column:parent_id;index;comment:父部门ID（顶级为空）" json:"parent_id"`
	SortOrder int    `gorm:"column:sort_order;not null;default:0;comment:排序" json:"sort_order"`
}

// TableName 指定表名
func (Department) TableName() string {
	return "departments"
}
