package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// MigrateDatabase 执行数据库迁移
func MigrateDatabase(dbPath string) error {
	fmt.Printf("开始迁移数据库：%s\n", dbPath)
	
	// 如果数据库文件不存在，直接返回
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("数据库文件不存在，无需迁移")
		return nil
	}
	
	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败：%w", err)
	}
	defer db.Close()
	
	// 检查是否需要添加新字段
	fmt.Println("检查数据库表结构...")
	
	// 添加 access_key_id 和 access_key_secret 字段
	columns := getTableColumns(db, "domains")
	
	if !contains(columns, "access_key_id") {
		fmt.Println("添加 access_key_id 字段...")
		_, err := db.Exec(`ALTER TABLE domains ADD COLUMN access_key_id TEXT DEFAULT ''`)
		if err != nil {
			return fmt.Errorf("添加 access_key_id 字段失败：%w", err)
		}
		fmt.Println("✓ access_key_id 字段添加成功")
	}
	
	if !contains(columns, "access_key_secret") {
		fmt.Println("添加 access_key_secret 字段...")
		_, err := db.Exec(`ALTER TABLE domains ADD COLUMN access_key_secret TEXT DEFAULT ''`)
		if err != nil {
			return fmt.Errorf("添加 access_key_secret 字段失败：%w", err)
		}
		fmt.Println("✓ access_key_secret 字段添加成功")
	}
	
	fmt.Println("数据库迁移完成")
	return nil
}

// getTableColumns 获取表的所有列名
func getTableColumns(db *sql.DB, tableName string) []string {
	rows, err := db.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	
	var columns []string
	for rows.Next() {
		var cid, name, dataType, notNull, defaultValue, pk string
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			continue
		}
		columns = append(columns, name)
	}
	
	return columns
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
