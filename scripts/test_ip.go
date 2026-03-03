//go:build ignore

package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 测试 IP 获取功能...")
	
	// 测试 1: 直接 HTTP 请求
	fmt.Println("\n📡 测试 1: 直接 HTTP 请求 api.ipify.org")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		fmt.Printf("❌ 请求失败：%v\n", err)
	} else {
		defer resp.Body.Close()
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		fmt.Printf("✅ 响应状态码：%d\n", resp.StatusCode)
		fmt.Printf("✅ 获取到 IP: %s\n", string(buf[:n]))
	}
	
	// 测试 2: 备用 API
	fmt.Println("\n📡 测试 2: 备用 API ifconfig.me")
	resp2, err := client.Get("https://ifconfig.me/ip")
	if err != nil {
		fmt.Printf("❌ 请求失败：%v\n", err)
	} else {
		defer resp2.Body.Close()
		buf := make([]byte, 1024)
		n, _ := resp2.Body.Read(buf)
		fmt.Printf("✅ 响应状态码：%d\n", resp2.StatusCode)
		fmt.Printf("✅ 获取到 IP: %s\n", string(buf[:n]))
	}
	
	fmt.Println("\n✅ 测试完成！")
}
