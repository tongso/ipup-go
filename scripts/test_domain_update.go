//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

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
	
	fmt.Println("🔍 测试域名更新功能...")
	fmt.Println("============================================================")
	
	// 1. 先查看当前所有域名
	fmt.Println("\n📋 当前域名列表:")
	rows, err := db.Query("SELECT id, domain, provider, enabled FROM domains ORDER BY id")
	if err != nil {
		fmt.Printf("❌ 查询域名失败：%v\n", err)
		return
	}
	
	var domains []struct {
		ID       int
		Domain   string
		Provider string
		Enabled  bool
	}
	
	for rows.Next() {
		var d struct {
			ID       int
			Domain   string
			Provider string
			Enabled  bool
		}
		if err := rows.Scan(&d.ID, &d.Domain, &d.Provider, &d.Enabled); err != nil {
			continue
		}
		domains = append(domains, d)
		status := "❌"
		if d.Enabled {
			status = "✅"
		}
		fmt.Printf("  [%d] %s | %s | %s\n", d.ID, status, d.Domain, d.Provider)
	}
	rows.Close()
	
	if len(domains) == 0 {
		fmt.Println("  ⚠️  没有域名配置，无法测试")
		return
	}
	
	// 2. 选择第一个域名进行测试
	testDomain := domains[0]
	fmt.Printf("\n🧪 测试更新域名 ID: %d (%s)\n", testDomain.ID, testDomain.Domain)
	
	// 3. 执行更新操作（修改域名）
	newDomainName := fmt.Sprintf("updated-%d.%s", time.Now().Unix(), testDomain.Domain)
	updateSQL := `UPDATE domains SET domain = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := db.Exec(updateSQL, newDomainName, testDomain.ID)
	if err != nil {
		fmt.Printf("❌ 更新失败：%v\n", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ 更新成功，影响行数：%d\n", rowsAffected)
	
	// 4. 验证更新结果
	fmt.Println("\n📋 验证更新后的域名:")
	var updatedDomain string
	err = db.QueryRow("SELECT domain FROM domains WHERE id = ?", testDomain.ID).Scan(&updatedDomain)
	if err != nil {
		fmt.Printf("❌ 查询更新后的域名失败：%v\n", err)
		return
	}
	
	fmt.Printf("  域名 ID %d 现在是：%s\n", testDomain.ID, updatedDomain)
	
	if updatedDomain == newDomainName {
		fmt.Println("\n✅ 域名更新测试通过！")
	} else {
		fmt.Println("\n❌ 域名未正确更新！")
	}
	
	// 5. 恢复原始域名
	fmt.Println("\n🔄 恢复原始域名...")
	_, err = db.Exec(`UPDATE domains SET domain = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, testDomain.Domain, testDomain.ID)
	if err != nil {
		fmt.Printf("❌ 恢复失败：%v\n", err)
		return
	}
	
	var restoredDomain string
	err = db.QueryRow("SELECT domain FROM domains WHERE id = ?", testDomain.ID).Scan(&restoredDomain)
	if err != nil {
		fmt.Printf("❌ 查询恢复后的域名失败：%v\n", err)
		return
	}
	
	fmt.Printf("  域名已恢复为：%s\n", restoredDomain)
	
	if restoredDomain == testDomain.Domain {
		fmt.Println("\n✅ 域名已恢复原状")
	} else {
		fmt.Println("\n❌ 域名恢复失败")
	}
	
	fmt.Println("\n============================================================")
	fmt.Println("✅ 测试完成！")
}
