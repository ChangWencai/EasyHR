package todo

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// statusLabel 状态中文映射
func statusLabel(s string) string {
	switch s {
	case TodoStatusPending:
		return "待办"
	case TodoStatusCompleted:
		return "已完成"
	case TodoStatusTerminated:
		return "已终止"
	default:
		return s
	}
}

// urgencyLabel 紧迫状态中文映射
func urgencyLabel(s string) string {
	switch s {
	case UrgencyNormal:
		return "正常"
	case UrgencyOverdue:
		return "超时"
	case UrgencyExpired:
		return "失效"
	default:
		return ""
	}
}

// ExportTodosExcel 导出待办列表 Excel
// 列：序号 | 事项名称 | 类型 | 员工姓名 | 发起人 | 创建时间 | 截止日期 | 状态 | 紧迫状态 | 限时任务
func ExportTodosExcel(c *gin.Context, items []TodoItem) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "待办事项"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"序号", "事项名称", "类型", "员工姓名", "发起人",
		"创建时间", "截止日期", "状态", "紧迫状态", "限时任务",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4F6EF7"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}
	lastHeaderCol, _ := excelize.CoordinatesToCellName(len(headers), 1)
	f.SetCellStyle(sheetName, "A1", lastHeaderCol, headerStyle)

	cstZone := time.FixedZone("CST", 8*3600)

	for rowIdx, item := range items {
		row := rowIdx + 2

		cstCreatedAt := item.CreatedAt.In(cstZone)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), rowIdx+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Type)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.EmployeeName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.CreatorName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), cstCreatedAt.Format("2006-01-02 15:04"))
		if item.Deadline != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.Deadline.Format("2006-01-02"))
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), statusLabel(item.Status))
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), urgencyLabel(item.UrgencyStatus))
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), map[bool]string{true: "是", false: "否"}[item.IsTimeLimited])

		// 状态颜色：超时=红色，失效=灰色
		statusStyle := -1
		if item.UrgencyStatus == UrgencyOverdue {
			statusStyle, _ = f.NewStyle(&excelize.Style{
				Font: &excelize.Font{Color: "#FF5630"},
			})
		} else if item.UrgencyStatus == UrgencyExpired {
			statusStyle, _ = f.NewStyle(&excelize.Style{
				Font: &excelize.Font{Color: "#8C8C8C"},
			})
		}
		if statusStyle >= 0 {
			cell, _ := excelize.CoordinatesToCellName(8, row)
			f.SetCellStyle(sheetName, cell, cell, statusStyle)
		}
	}

	// 列宽
	colWidths := map[string]float64{"A": 6, "B": 40, "C": 16, "D": 12, "E": 12, "F": 20, "G": 14, "H": 10, "I": 10, "J": 10}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return fmt.Errorf("write excel buffer: %w", err)
	}

	timestamp := time.Now().In(time.FixedZone("CST", 8*3600)).Format("20060102_150405")
	filename := fmt.Sprintf("待办事项_%s.xlsx", timestamp)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())

	return nil
}
