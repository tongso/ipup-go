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
	
	fmt.Println("🔍 测试域名状态显示逻辑...")
	fmt.Println("============================================================")
	
	// 查询所有启用的域名
	querySQL := `
	SELECT domain, COALESCE(current_ip, ''), COALESCE(last_update, '')
	FROM domains
	WHERE enabled = 1
	ORDER BY domain
	`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		return
	}
	defer rows.Close()
	
	fmt.Println("\n📊 启用域名的状态分析:")
	fmt.Println("-----------------------------------------------------------")
	
	count := 0
	pendingCount := 0
	successCount := 0
	
	for rows.Next() {
		var domain, currentIP, lastUpdate string
		if err := rows.Scan(&domain, &currentIP, &lastUpdate); err != nil {
			continue
		}
		
		count++
		
		// 模拟后端的状态判断逻辑
		status := ""
		message := ""
		
		if currentIP == "" {
			status = "pending"
			message = "等待首次更新"
			pendingCount++
		} else if lastUpdate == "" {
			status = "pending"
			message = "已获取 IP，等待更新 DNS"
			pendingCount++
		} else {
			status = "success"
			message = "解析正常"
			successCount++
		}
		
		// 显示详细信息
		fmt.Printf("\n🌐 域名：%s\n", domain)
		fmt.Printf("   当前 IP: %s\n", formatIP(currentIP))
		fmt.Printf("   最后更新：%s\n", formatTime(lastUpdate))
		fmt.Printf("   状态：%s - %s\n", getStatusIcon(status), message)
	}
	
	fmt.Println("\n============================================================")
	fmt.Println("📈 统计信息:")
	fmt.Printf("   总域名数：%d\n", count)
	fmt.Printf("   ✅ 解析正常：%d\n", successCount)
	fmt.Printf("   ⏳ 等待更新：%d\n", pendingCount)
	
	if count == 0 {
		fmt.Println("\n⚠️  暂无启用的域名")
		fmt.Println("💡 提示：请前往「域名管理」添加域名配置")
	} else if pendingCount > 0 {
		fmt.Println("\n💡 提示：部分域名等待 DDNS 更新")
		fmt.Println("   - 新添加的域名会显示「等待首次更新」")
		fmt.Println("   - 系统会在后台自动执行 DDNS 更新")
		fmt.Println("   - 更新后这里会显示实际的 IP 地址和更新时间")
	} else {
		fmt.Println("\n✅ 所有域名解析正常！")
	}
	
	fmt.Println("\n============================================================")
	fmt.Println("✅ 测试完成！")
}

// formatIP 格式化 IP 显示
func formatIP(ip string) string {
	if ip == "" {
		return "⏳ (空，等待更新)"
	}
	return "✅ " + ip
}

// formatTime 格式化时间显示
func formatTime(t string) string {
	if t == "" {
		return "⏳ (空，未更新)"
	}
	return "✅ " + t
}

// getStatusIcon 返回状态图标
func getStatusIcon(status string) string {
	switch status {
	case "success":
		return "✅"
	case "pending":
		return "⏳"
	case "error":
		return "❌"
	default:
		return "❓"
	}
}
