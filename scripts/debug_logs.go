//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	
	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("=== 检查日志数据和查询 ===\n")
	
	// 打开数据库
	db, err := sql.Open("sqlite", "./ipup.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// 测试 1: 直接查询所有日志
	fmt.Println("1. 查询所有日志（无过滤）:")
	rows, err := db.Query(`SELECT id, timestamp, level, domain, message FROM logs ORDER BY timestamp DESC LIMIT 10`)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		return
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var id int
		var timestamp, level, domain, message string
		if err := rows.Scan(&id, &timestamp, &level, &domain, &message); err != nil {
			fmt.Printf("扫描失败：%v\n", err)
			continue
		}
		fmt.Printf("   [%s] [%s] %s: %s\n", timestamp, level, domain, message)
		count++
	}
	fmt.Printf("共 %d 条\n", count)
	
	// 测试 2: 使用修复后的 GetLogs 查询逻辑
	fmt.Println("\n2. 模拟修复后的 GetLogs 查询（level='all', keyword=''）:")
	level = "all"
	keyword = ""
	
	// 修复后的逻辑
	levelParam := ""
	if level != "all" && level != "" {
		levelParam = level
	}
	
	querySQL = `
	SELECT id, timestamp, level, domain, message
	FROM logs
	WHERE (? = '' OR level = ?) AND (? = '' OR domain LIKE ? OR message LIKE ?)
	ORDER BY timestamp DESC
	LIMIT ?
	`
	
	rows, err = db.Query(querySQL, levelParam, levelParam, keyword, "%"+keyword+"%", "%"+keyword+"%", 100)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		return
	}
	defer rows.Close()
	
	count = 0
	for rows.Next() {
		var id int
		var timestamp, level, domain, message string
		if err := rows.Scan(&id, &timestamp, &level, &domain, &message); err != nil {
			fmt.Printf("扫描失败：%v\n", err)
			continue
		}
		fmt.Printf("   [%s] [%s] %s: %s\n", timestamp, level, domain, message)
		count++
	}
	fmt.Printf("共 %d 条\n", count)
	
	// 测试 3: 检查 level 字段的唯一值
	fmt.Println("\n3. 检查日志级别分布:")
	rows, err = db.Query(`SELECT level, COUNT(*) as count FROM logs GROUP BY level ORDER BY count DESC`)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			continue
		}
		fmt.Printf("   - %s: %d 条\n", level, count)
	}
	
	fmt.Println("\n=== 检查完成 ===")
}
