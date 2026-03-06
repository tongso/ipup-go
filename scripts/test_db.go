//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	
	_ "modernc.org/sqlite"
)

func main() {
	// 打开数据库
	db, err := sql.Open("sqlite", "./ipup.db")
	if err != nil {
		fmt.Printf("❌ 打开数据库失败：%v\n", err)
		return
	}
	defer db.Close()
	
	// 测试连接
	if err := db.Ping(); err != nil {
		fmt.Printf("❌ 连接数据库失败：%v\n", err)
		return
	}
	
	fmt.Println("✅ 数据库连接成功")
	
	// 检查 domains 表结构
	rows, err := db.Query(`PRAGMA table_info(domains)`)
	if err != nil {
		fmt.Printf("❌ 查询表结构失败：%v\n", err)
		return
	}
	defer rows.Close()
	
	fmt.Println("\n📋 domains 表结构:")
	fmt.Println("CID | Name | Type | NotNull | Default | PK")
	fmt.Println("---|------|------|---------|---------|---")
	
	for rows.Next() {
		var cid, name, dataType, notNull, defaultValue, pk string
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			continue
		}
		fmt.Printf("%-3s | %-20s | %-10s | %-7s | %-20s | %s\n", 
			cid, name, dataType, notNull, defaultValue, pk)
	}
	
	// 查询域名数量
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM domains`).Scan(&count)
	if err != nil {
		fmt.Printf("\n❌ 查询域名数量失败：%v\n", err)
	} else {
		fmt.Printf("\n📊 当前域名数量：%d\n", count)
	}
	
	// 尝试插入测试数据
	fmt.Println("\n🧪 尝试插入测试数据...")
	testDomain := "test.example.com"
	insertSQL := `INSERT OR REPLACE INTO domains (domain, provider, token, access_key_id, access_key_secret, interval, enabled) VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(insertSQL, testDomain, "Aliyun", "test_token", "test_key_id", "test_secret", 300, 1)
	if err != nil {
		fmt.Printf("❌ 插入测试数据失败：%v\n", err)
	} else {
		fmt.Println("✅ 插入测试数据成功")
		
		id, _ := result.LastInsertId()
		fmt.Printf("   插入的 ID: %d\n", id)
	}
	
	// 查询所有域名
	fmt.Println("\n📝 域名列表:")
	querySQL := `SELECT id, domain, provider, token, access_key_id, access_key_secret, interval, enabled FROM domains ORDER BY id`
	rows2, err := db.Query(querySQL)
	if err != nil {
		fmt.Printf("❌ 查询域名失败：%v\n", err)
		return
	}
	defer rows2.Close()
	
	for rows2.Next() {
		var id, interval, enabled int
		var domain, provider, token, accessKeyID, accessKeySecret string
		if err := rows2.Scan(&id, &domain, &provider, &token, &accessKeyID, &accessKeySecret, &interval, &enabled); err != nil {
			continue
		}
		fmt.Printf("  ID: %d | Domain: %s | Provider: %s | Interval: %d | Enabled: %v\n", 
			id, domain, provider, interval, enabled)
	}
}
