//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	
	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("=== 检查数据库日志记录 ===\n")
	
	// 打开数据库
	db, err := sql.Open("sqlite", "./ipup.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// 检查表结构
	fmt.Println("1. 检查日志表结构:")
	rows, err := db.Query(`PRAGMA table_info(logs)`)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull int
		var dflt_value, pk interface{}
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		fmt.Printf("   - %s (%s)\n", name, dtype)
	}
	
	// 统计日志数量
	fmt.Println("\n2. 统计日志数量:")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
	} else {
		fmt.Printf("   📝 共有 %d 条日志\n", count)
	}
	
	// 显示最近 10 条日志
	fmt.Println("\n3. 最近 10 条日志:")
	rows, err = db.Query(`
		SELECT id, timestamp, level, domain, message 
		FROM logs 
		ORDER BY timestamp DESC 
		LIMIT 10
	`)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		return
	}
	defer rows.Close()
	
	found := false
	for rows.Next() {
		found = true
		var id int
		var timestamp, level, domain, message string
		if err := rows.Scan(&id, &timestamp, &level, &domain, &message); err != nil {
			continue
		}
		fmt.Printf("   [%s] [%s] %s: %s\n", timestamp, level, domain, message)
	}
	
	if !found {
		fmt.Println("   (暂无日志记录)")
	}
	
	// 检查域名操作
	fmt.Println("\n4. 域名配置数量:")
	err = db.QueryRow("SELECT COUNT(*) FROM domains").Scan(&count)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
	} else {
		fmt.Printf("   🌐 共有 %d 个域名\n", count)
	}
	
	// 检查设置
	fmt.Println("\n5. 系统设置数量:")
	err = db.QueryRow("SELECT COUNT(*) FROM settings").Scan(&count)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
	} else {
		fmt.Printf("   ⚙️ 共有 %d 项设置\n", count)
	}
	
	fmt.Println("\n=== 检查完成 ===")
}
