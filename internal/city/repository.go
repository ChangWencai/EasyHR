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
// 支持6位编码（如110000）和12位编码（如110000000000）的查询
func (r *Repository) GetNameByCode(code int64) string {
	// 先尝试精确匹配
	area, err := r.GetByCode(code)
	if err == nil {
		return area.Name
	}

	// 6位编码需要补0变成12位进行查询（如110000 -> 110000000000）
	// 12位编码取前2位补0得到省级编码
	var provinceCode int64

	if code >= 100000000000 {
		// 12位编码：取前2位（省）+10个0
		provinceCode = (code / 100000000) * 100000000
	} else if code >= 100000 {
		// 6位编码：补6个0变成12位
		provinceCode = code * 1000000
	} else {
		return "未知城市"
	}

	// 查询省级名称
	var province AreaCode
	if err := r.db.Where("code = ?", provinceCode).First(&province).Error; err != nil {
		return "未知城市"
	}
	return province.Name
}