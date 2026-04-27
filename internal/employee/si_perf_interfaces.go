package employee

// SICreator 社保创建接口（解耦 employee 与 socialinsurance 包）
type SICreator interface {
	// CreateEmployeeSI 创建员工社保参保记录
	// cityCode: 参保城市编码；baseAmount: 社保基数（为空用员工薪资）；startMonth: 参保起始月 YYYY-MM；hfBase: 公积金基数
	CreateEmployeeSI(orgID, userID, empID int64, empName string, cityCode int64, baseAmount float64, startMonth string, hfBase float64) error
}

// PerfCreator 绩效系数初始化接口（解耦 employee 与 salary 包）
type PerfCreator interface {
	// InitEmployeePerf 初始化员工绩效系数
	InitEmployeePerf(orgID, userID, empID int64, year, month int, coefficient float64) error
}
