package main

import (
	"fmt"
	
	"ipup-go/internal/provider"
)

func main() {
	fmt.Println("=== 阿里云 DNS Provider 测试 ===")
	
	// 注意：这里使用测试密钥，实际使用时应该从配置文件或环境变量读取
	accessKeyID := "your_access_key_id"
	accessKeySecret := "your_access_key_secret"
	testDomain := "test.example.com"
	testIP := "1.2.3.4"
	
	// 创建 Provider 实例
	p := provider.NewAliyunProvider(accessKeyID, accessKeySecret)
	
	fmt.Printf("测试域名：%s\n", testDomain)
	fmt.Printf("测试 IP: %s\n", testIP)
	
	// 测试 1: 查询现有记录
	fmt.Println("\n【测试 1】查询现有记录...")
	recordID, err := p.describeDomainRecord("example.com", "test")
	if err != nil {
		fmt.Printf("查询失败：%v\n", err)
	} else if recordID == "" {
		fmt.Println("未找到现有记录")
	} else {
		fmt.Printf("找到记录 ID: %s\n", recordID)
	}
	
	// 测试 2: 创建新记录（如果记录不存在）
	fmt.Println("\n【测试 2】创建新记录...")
	err = p.addDomainRecord("example.com", "test", testIP)
	if err != nil {
		fmt.Printf("创建失败：%v\n", err)
	} else {
		fmt.Println("创建成功")
	}
	
	// 测试 3: 更新现有记录（如果记录存在）
	fmt.Println("\n【测试 3】更新现有记录...")
	if recordID != "" {
		err = p.updateDomainRecord(recordID, "test", testIP)
		if err != nil {
			fmt.Printf("更新失败：%v\n", err)
		} else {
			fmt.Println("更新成功")
		}
	} else {
		fmt.Println("跳过更新测试（无现有记录）")
	}
	
	// 测试 4: 完整流程（自动判断创建或更新）
	fmt.Println("\n【测试 4】完整流程测试...")
	err = p.UpdateRecord(testDomain, testIP)
	if err != nil {
		fmt.Printf("UpdateRecord 失败：%v\n", err)
	} else {
		fmt.Println("UpdateRecord 执行成功")
	}
	
	fmt.Println("\n=== 测试完成 ===")
}