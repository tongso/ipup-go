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
		fmt.Println("请设置环境变量:")
		fmt.Println("  ALIYUN_ACCESS_KEY_ID=your_access_key_id")
		fmt.Println("  ALIYUN_ACCESS_KEY_SECRET=your_access_key_secret")
		fmt.Println("  TEST_DOMAIN=your_domain.com")
		os.Exit(1)
	}
	
	// 创建阿里云 DNS 提供商
	aliyunProvider := provider.NewAliyunProvider(accessKeyID, accessKeySecret)
	
	fmt.Printf("🧪 测试阿里云 DNS 更新\n")
	fmt.Printf("域名：%s\n", domain)
	
	// 获取当前 IP（模拟）
	currentIP := "123.123.123.123"
	
	// 测试更新 DNS 记录
	fmt.Println("\n📝 开始更新 DNS 记录...")
	err := aliyunProvider.UpdateRecord(domain, currentIP)
	if err != nil {
		fmt.Printf("❌ 更新失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ DNS 记录更新成功！")
	
	// 测试获取 DNS 记录
	fmt.Println("\n📖 查询当前 DNS 记录...")
	recordedIP, err := aliyunProvider.GetRecord(domain)
	if err != nil {
		fmt.Printf("❌ 查询失败：%v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ 当前解析 IP: %s\n", recordedIP)
	
	if recordedIP == currentIP {
		fmt.Println("\n🎉 测试通过！DNS 记录已正确更新。")
	} else {
		fmt.Printf("\n⚠️ 警告：记录的 IP (%s) 与预期值 (%s) 不一致\n", recordedIP, currentIP)
	}
}
