//go:build ignore

package main

import (
	"fmt"
	"os"
	
	"ipup-go/internal/provider"
)

func main() {
	// 从环境变量获取阿里云 AccessKey 配置
	accessKeyID := os.Getenv("ALIYUN_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	domain := os.Getenv("TEST_DOMAIN")
	
	if accessKeyID == "" || accessKeySecret == "" || domain == "" {
		fmt.Println("❌ 请设置环境变量:")
		fmt.Println("  ALIYUN_ACCESS_KEY_ID=your_access_key_id")
		fmt.Println("  ALIYUN_ACCESS_KEY_SECRET=your_access_key_secret")
		fmt.Println("  TEST_DOMAIN=your_domain.com")
		os.Exit(1)
	}
	
	fmt.Println("========================================")
	fmt.Println("🧪 测试阿里云 UpdateDomainRecord API")
	fmt.Println("========================================")
	fmt.Printf("域名：%s\n", domain)
	fmt.Printf("AccessKey ID: %s\n", maskString(accessKeyID))
	fmt.Println()
	
	// 创建阿里云 DNS 提供商
	aliyunProvider := provider.NewAliyunProvider(accessKeyID, accessKeySecret)
	
	// 步骤 1: 查询当前 DNS 记录
	fmt.Println("📖 步骤 1: 查询当前 DNS 记录...")
	recordedIP, err := aliyunProvider.GetRecord(domain)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 当前解析 IP: %s\n", recordedIP)
	
	// 步骤 2: 更新 DNS 记录（使用一个测试 IP）
	testIP := "192.0.2.1" // 使用 RFC 5737 文档地址作为测试
	fmt.Printf("\n📝 步骤 2: 更新 DNS 记录到测试 IP: %s\n", testIP)
	err = aliyunProvider.UpdateRecord(domain, testIP)
	if err != nil {
		fmt.Printf("❌ 更新失败：%v\n", err)
		fmt.Println("\n💡 可能的原因:")
		fmt.Println("   1. AccessKey 权限不足（需要 AliyunDNSFullAccess）")
		fmt.Println("   2. 域名不在该 AccessKey 账号下")
		fmt.Println("   3. API 参数错误")
		fmt.Println("   4. 网络问题")
		os.Exit(1)
	}
	fmt.Println("✅ DNS 记录更新成功！")
	
	// 步骤 3: 验证更新结果
	fmt.Println("\n📖 步骤 3: 验证更新结果...")
	updatedIP, err := aliyunProvider.GetRecord(domain)
	if err != nil {
		fmt.Printf("❌ 验证失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 更新后解析 IP: %s\n", updatedIP)
	
	// 步骤 4: 判断是否一致
	fmt.Println("\n🔍 验证结果:")
	if updatedIP == testIP {
		fmt.Println("🎉 测试通过！DNS 记录已正确更新。")
		fmt.Println("✅ UpdateDomainRecord API 工作正常")
	} else {
		fmt.Printf("⚠️ 警告：实际更新的 IP (%s) 与预期值 (%s) 不一致\n", updatedIP, testIP)
		fmt.Println("❌ 可能存在其他问题，请检查日志")
		os.Exit(1)
	}
	
	// 可选：恢复原来的 IP
	if recordedIP != "" && recordedIP != testIP {
		fmt.Printf("\n🔄 恢复原始 DNS 记录到：%s\n", recordedIP)
		err = aliyunProvider.UpdateRecord(domain, recordedIP)
		if err != nil {
			fmt.Printf("⚠️ 恢复失败：%v\n", err)
			fmt.Println("⚠️ 请手动恢复原始 IP 地址")
		} else {
			fmt.Println("✅ 已恢复原始 DNS 记录")
		}
	}
}

// maskString 隐藏字符串中间部分，用于安全显示
func maskString(s string) string {
	if len(s) <= 8 {
		return "***HIDDEN***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}
