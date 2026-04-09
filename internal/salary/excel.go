package salary

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// AttendanceRow 考勤导入行
type AttendanceRow struct {
	Name               string
	SickLeaveDays      float64
	PersonalLeaveDays  float64
	Remark             string
}

// parseAttendanceExcel 解析考勤 Excel 文件
// 模板格式：员工姓名 | 事假(天) | 病假(天) | 备注
func parseAttendanceExcel(file []byte) ([]AttendanceRow, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(file)))
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer f.Close()

	// 读取第一个 sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel 文件没有工作表")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 行失败: %w", err)
	}

	var result []AttendanceRow
	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}
		if len(row) < 1 || strings.TrimSpace(row[0]) == "" {
			continue // 空行跳过
		}

		ar := AttendanceRow{
			Name: strings.TrimSpace(row[0]),
		}

		if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
			var sickDays float64
			fmt.Sscanf(strings.TrimSpace(row[1]), "%f", &sickDays)
			ar.SickLeaveDays = sickDays
		}
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			var personalDays float64
			fmt.Sscanf(strings.TrimSpace(row[2]), "%f", &personalDays)
			ar.PersonalLeaveDays = personalDays
		}
		if len(row) > 3 {
			ar.Remark = strings.TrimSpace(row[3])
		}

		result = append(result, ar)
	}

	return result, nil
}
