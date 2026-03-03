package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GetPublicIP 从第三方 API 获取公网 IP（同时获取 IPv4 和 IPv6）
func GetPublicIP(apiEndpoint string) (string, error) {
	// 获取 IPv4 和 IPv6
	ipv4 := getIPv4()
	ipv6 := getIPv6()
	
	// 返回优先使用的 IP
	if ipv4 != "" {
		return ipv4, nil
	}
	if ipv6 != "" {
		return ipv6, nil
	}
	
	return "", fmt.Errorf("无法获取 IPv4 或 IPv6 地址")
}

// GetDualStackIP 获取双栈 IP 信息
func GetDualStackIP() (ipv4, ipv6 string) {
	ipv4 = getIPv4()
	ipv6 = getIPv6()
	return
}

// getIPv4 获取 IPv4 地址
func getIPv4() string {
	// IPv4 API 端点列表
	endpoints := []string{
		"https://api.ipify.org",
		"https://v4.ident.me",
		"https://v4.icanhazip.com",
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	for _, endpoint := range endpoints {
		ip := fetchIP(endpoint, client)
		if ip != "" && isIPv4(ip) {
			return ip
		}
	}
	
	return ""
}

// getIPv6 获取 IPv6 地址
func getIPv6() string {
	// IPv6 API 端点列表
	endpoints := []string{
		"https://v6.ident.me",
		"https://v6.icanhazip.com",
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	for _, endpoint := range endpoints {
		ip := fetchIP(endpoint, client)
		if ip != "" && isIPv6(ip) {
			return ip
		}
	}
	
	return ""
}

// fetchIP 从指定端点获取 IP
func fetchIP(endpoint string, client *http.Client) string {
	resp, err := client.Get(endpoint)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
	ip := strings.TrimSpace(string(body))
	return ip
}

// isIPv4 判断是否为 IPv4 地址
func isIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

// isIPv6 判断是否为 IPv6 地址
func isIPv6(ip string) bool {
	return strings.Contains(ip, ":")
}

// trimSpace 移除字符串两端的空白字符（包括换行符）
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	// 移除前导空白
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	
	// 移除尾部空白
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}

// FormatTime 格式化时间为字符串
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// CurrentTime 获取当前时间的字符串表示
func CurrentTime() string {
	return FormatTime(time.Now())
}
