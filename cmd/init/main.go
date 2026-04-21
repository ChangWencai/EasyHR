package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/database"
	"github.com/wencai/easyhr/internal/common/logger"
	"go.uber.org/zap"
)

type AreaCode struct {
	Code     int64       `json:"code"`
	Name     string      `json:"name"`
	Level    int         `json:"level"`
	Pcode    int64       `json:"pcode"`
	Category *int        `json:"category,omitempty"`
	Children []AreaCode  `json:"children,omitempty"`
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.InitWithConfig(&cfg.Log)
	db := database.Init(&cfg.Database)

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get sql.DB: %v\n", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	jsonPath := "migrations/area_code_2024.json"
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "file not found: %s\n", jsonPath)
		os.Exit(1)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", jsonPath, err)
		os.Exit(1)
	}

	var areas []AreaCode
	if err := json.Unmarshal(data, &areas); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正在初始化区划编码数据...")
	fmt.Printf("  JSON 文件: %s\n", filepath.Base(jsonPath))

	// 创建表
	createSQL := `
		DROP TABLE IF EXISTS area_code;
		CREATE TABLE area_code (
			code     bigint        NOT NULL,
			name     varchar(128) NOT NULL DEFAULT '',
			level    smallint      NOT NULL,
			pcode    bigint,
			category integer       DEFAULT NULL,
			PRIMARY KEY (code)
		);
		CREATE INDEX idx_area_code_level ON area_code (level);
		CREATE INDEX idx_area_code_pcode ON area_code (pcode);
	`
	if err := db.Exec(createSQL).Error; err != nil {
		logger.Logger.Error("create table failed", zap.Error(err))
		fmt.Fprintf(os.Stderr, "创建表失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ 表已创建: area_code")

	// 扁平化 JSON 数据
	var records [][]interface{}
	flattenAreaCodes(areas, &records)

	fmt.Printf("  ✓ 解析完成，共 %d 条记录\n", len(records))

	// 批量插入（每 5000 条一批）
	batchSize := 5000
	inserted := 0
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		valueStrings := ""
		valueArgs := []interface{}{}
		for _, r := range batch {
			if valueStrings != "" {
				valueStrings += ","
			}
			valueStrings += "(?, ?, ?, ?, ?)"
			valueArgs = append(valueArgs, r[0], r[1], r[2], r[3], r[4])
		}

		sql := fmt.Sprintf("INSERT INTO area_code (code, name, level, pcode, category) VALUES %s", valueStrings)
		if err := db.Exec(sql, valueArgs...).Error; err != nil {
			logger.Logger.Error("insert failed", zap.Error(err))
			fmt.Fprintf(os.Stderr, "插入数据失败: %v\n", err)
			os.Exit(1)
		}

		inserted += len(batch)
		fmt.Printf("  ✓ 已插入 %d / %d 条\n", inserted, len(records))
	}

	fmt.Println("\n区划编码数据初始化完成。")
}

func flattenAreaCodes(areas []AreaCode, records *[][]interface{}) {
	for _, a := range areas {
		rec := []interface{}{a.Code, a.Name, a.Level, a.Pcode, 0}
		if a.Category != nil {
			rec[4] = *a.Category
		}
		*records = append(*records, rec)

		if len(a.Children) > 0 {
			flattenAreaCodes(a.Children, records)
		}
	}
}
