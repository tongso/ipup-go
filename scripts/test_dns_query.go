//go:build ignore

package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("🔍 测试域名 DNS 实时查询功能...")
	fmt.Println("============================================================")
	
	// 测试域名列表（从数据库读取启用的域名）
	testDomains := []string{
		"test155555555.xx.com", // 示例域名
		"www.baidu.com",        // 测试用真实域名
		"www.qq.com",           // 测试用真实域名
	}
	
	fmt.Println("\n📊 实时查询域名 DNS 解析:")
	fmt.Println("-----------------------------------------------------------")
	
	for _, domain := range testDomains {
		fmt.Printf("\n🌐 域名：%s\n", domain)
		
		currentIP, queryTime := queryDomainDNS(domain)
		
		if currentIP == "" {
			fmt.Printf("   ❌ DNS 解析失败\n")
			fmt.Printf("   状态：等待首次更新\n")
		} else {
			fmt.Printf("   ✅ 当前 IP: %s\n", currentIP)
			fmt.Printf("   🕒 查询时间：%s\n", queryTime)
			fmt.Printf("   状态：解析正常\n")
		}
	}
	
	fmt.Println("\n============================================================")
	fmt.Println("💡 说明:")
	fmt.Println("   - 域名 IP 是实时从 DNS 服务商获取的")
	fmt.Println("   - 不依赖数据库存储")
	fmt.Println("   - 每次刷新都会重新查询")
	fmt.Println("   - 公网 IP 变化后调用 DNS API 也会更新")
	fmt.Println("\n✅ 测试完成！")
}

// queryDomainDNS 查询域名的 DNS 解析记录
func queryDomainDNS(domainName string) (string, string) {
	// 设置超时
	timeout := time.Duration(5 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// 使用系统 DNS 解析
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, domainName)
	queryTime := time.Now().Format("2006-01-02 15:04:05")
	
	if err != nil || len(ips) == 0 {
		// DNS 解析失败或无结果
		return "", ""
	}
	
	// 优先返回 IPv4
	for _, ip := range ips {
		if ipv4 := ip.IP.To4(); ipv4 != nil {
			return ipv4.String(), queryTime
		}
	}
	
	// 如果没有 IPv4，返回第一个 IPv6
	if len(ips) > 0 {
		return ips[0].IP.String(), queryTime
	}
	
	return "", ""
}
