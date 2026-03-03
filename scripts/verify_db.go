//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "./ipup.db"
	
	// 检查数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ 数据库文件不存在：%s\n", dbPath)
		fmt.Println("请先运行应用，会自动创建数据库")
		return
	}
	
	// 打开数据库
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("❌ 打开数据库失败：%v\n", err)
		return
	}
	defer db.Close()
	
	fmt.Println("✅ 数据库文件存在:", dbPath)
	fmt.Println("============================================================")
	
	// 1. 检查 settings 表
	fmt.Println("\n📋 检查 settings 表:")
	settingsRows, err := db.Query("SELECT key, value, updated_at FROM settings ORDER BY key")
	if err != nil {
		fmt.Printf("❌ 查询 settings 表失败：%v\n", err)
	} else {
		defer settingsRows.Close()
		
		count := 0
		for settingsRows.Next() {
			var key, value, updatedAt string
			if err := settingsRows.Scan(&key, &value, &updatedAt); err != nil {
				continue
			}
			fmt.Printf("  %-20s | %-10s | %s\n", key, value, updatedAt)
			count++
		}
		
		if count == 0 {
			fmt.Println("  ⚠️  settings 表为空，没有保存任何设置")
		} else {
			fmt.Printf("  ✅ 共 %d 条设置记录\n", count)
		}
	}
	
	// 2. 检查 domains 表
	fmt.Println("\n🌐 检查 domains 表:")
	domainsRows, err := db.Query("SELECT id, domain, provider, enabled, current_ip, last_update FROM domains ORDER BY id")
	if err != nil {
		fmt.Printf("❌ 查询 domains 表失败：%v\n", err)
	} else {
		defer domainsRows.Close()
		
		count := 0
		enabledCount := 0
		for domainsRows.Next() {
			var id int
			var domain, provider string
			var enabled bool
			var currentIP, lastUpdate sql.NullString
			
			if err := domainsRows.Scan(&id, &domain, &provider, &enabled, &currentIP, &lastUpdate); err != nil {
				continue
			}
			
			ipStr := ""
			if currentIP.Valid {
				ipStr = currentIP.String
			}
			
			status := "❌"
			if enabled {
				status = "✅"
				enabledCount++
			}
			
			fmt.Printf("  [%s] %s | %s | IP: %s\n", status, domain, provider, ipStr)
			count++
		}
		
		if count == 0 {
			fmt.Println("  ⚠️  domains 表为空，没有域名配置")
		} else {
			fmt.Printf("  ✅ 共 %d 个域名 (%d 个启用)\n", count, enabledCount)
		}
	}
	
	// 3. 检查 logs 表
	fmt.Println("\n📝 检查 logs 表 (最近 10 条):")
	logsRows, err := db.Query(`
		SELECT timestamp, level, domain, message 
		FROM logs 
		ORDER BY timestamp DESC 
		LIMIT 10
	`)
	if err != nil {
		fmt.Printf("❌ 查询 logs 表失败：%v\n", err)
	} else {
		defer logsRows.Close()
		
		count := 0
		for logsRows.Next() {
			var timestamp, level, domain, message string
			if err := logsRows.Scan(&timestamp, &level, &domain, &message); err != nil {
				continue
			}
			
			icon := "ℹ️"
			switch level {
			case "error":
				icon = "❌"
			case "warning":
				icon = "⚠️"
			case "success":
				icon = "✅"
			}
			
			fmt.Printf("  %s [%s] %s: %s\n", timestamp, icon, domain, message)
			count++
		}
		
		if count == 0 {
			fmt.Println("  ⚠️  logs 表为空，没有日志记录")
		} else {
			fmt.Printf("  ✅ 最近 %d 条日志\n", count)
		}
	}
	
	// 4. 统计信息
	fmt.Println("\n============================================================")
	fmt.Println("📊 统计信息:")
	
	var settingCount, domainCount, logCount int
	db.QueryRow("SELECT COUNT(*) FROM settings").Scan(&settingCount)
	db.QueryRow("SELECT COUNT(*) FROM domains").Scan(&domainCount)
	db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&logCount)
	
	fmt.Printf("  Settings: %d 条\n", settingCount)
	fmt.Printf("  Domains:  %d 个\n", domainCount)
	fmt.Printf("  Logs:     %d 条\n", logCount)
	
	fmt.Println("\n✅ 数据库检查完成!")
}
