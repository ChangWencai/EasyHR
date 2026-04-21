package city

import "gorm.io/gorm"

// Repository 区划数据仓库
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓库
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetByLevel 获取指定层级的所有记录
func (r *Repository) GetByLevel(level int) ([]AreaCode, error) {
	var areas []AreaCode
	if err := r.db.Where("level = ?", level).Order("code ASC").Find(&areas).Error; err != nil {
		return nil, err
	}
	return areas, nil
}

// GetByLevelAndPcode 获取指定层级和父级编码的记录
func (r *Repository) GetByLevelAndPcode(level int, pcode int64) ([]AreaCode, error) {
	var areas []AreaCode
	if err := r.db.Where("level = ? AND pcode = ?", level, pcode).Order("code ASC").Find(&areas).Error; err != nil {
		return nil, err
	}
	return areas, nil
}

// GetByPcode 获取所有子级（不限制层级）
func (r *Repository) GetByPcode(pcode int64) ([]AreaCode, error) {
	var areas []AreaCode
	if err := r.db.Where("pcode = ?", pcode).Order("code ASC").Find(&areas).Error; err != nil {
		return nil, err
	}
	return areas, nil
}

// GetByCode 按编码查询
func (r *Repository) GetByCode(code int64) (*AreaCode, error) {
	var area AreaCode
	if err := r.db.Where("code = ?", code).First(&area).Error; err != nil {
		return nil, err
	}
	return &area, nil
}

// GetNameByCode 根据行政区划编码获取城市名称
func (r *Repository) GetNameByCode(code int64) string {
	area, err := r.GetByCode(code)
	if err != nil {
		return "未知城市"
	}
	return area.Name
}