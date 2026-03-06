//go:build ignore

package main

import (
	"context"
	"fmt"
	
	"ipup-go/internal/app"
)

func main() {
	fmt.Println("🧪 测试 App API 调用...")
	
	// 创建 App 实例
	appInstance := app.NewApp()
	
	// 初始化
	ctx := context.Background()
	err := appInstance.Startup(ctx)
	if err != nil {
		fmt.Printf("❌ 初始化失败：%v\n", err)
		return
	}
	defer appInstance.Shutdown(ctx)
	
	fmt.Println("✅ App 初始化成功")
	
	// 测试 ListDomains
	fmt.Println("\n📝 测试 ListDomains...")
	domains, err := appInstance.ListDomains()
	if err != nil {
		fmt.Printf("❌ ListDomains 失败：%v\n", err)
	} else {
		fmt.Printf("✅ ListDomains 成功，获取到 %d 个域名\n", len(domains))
		for i, d := range domains {
			fmt.Printf("  [%d] ID: %d | Domain: %s | Provider: %s | Enabled: %v\n", 
				i+1, d.ID, d.Domain, d.Provider, d.Enabled)
		}
	}
	
	// 测试 GetDomainStatus
	fmt.Println("\n📊 测试 GetDomainStatus...")
	statuses, err := appInstance.GetDomainStatus()
	if err != nil {
		fmt.Printf("❌ GetDomainStatus 失败：%v\n", err)
	} else {
		fmt.Printf("✅ GetDomainStatus 成功，获取到 %d 个域名状态\n", len(statuses))
		for i, s := range statuses {
			fmt.Printf("  [%d] Domain: %s | Status: %s | Message: %s\n", 
				i+1, s.Domain, s.Status, s.Message)
		}
	}
	
	// 测试 AddDomain
	fmt.Println("\n➕ 测试 AddDomain...")
	fmt.Println("⚠️ 跳过 AddDomain 测试（已在 test_db.go 中验证）")
	
	fmt.Println("\n✅ 所有测试完成！")
}
