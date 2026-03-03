package app

import (
	"fmt"
	
	"ipup-go/pkg/types"
	"ipup-go/pkg/utils"
)

// ==================== IP 和监控 API ====================

// GetPublicIP 获取公网 IP 信息（同时返回 IPv4 和 IPv6）
func (a *App) GetPublicIP() (types.IPInfo, error) {
	ipv4, ipv6 := utils.GetDualStackIP()
	
	publicIP := ipv4
	if publicIP == "" && ipv6 != "" {
		publicIP = ipv6 // 如果没有 IPv4，使用 IPv6
	}
	
	// TODO: 实现 IP 归属地查询
	return types.IPInfo{
		PublicIP: publicIP,
		IPv4:     ipv4,
		IPv6:     ipv6,
		Location: "未知",
		ISP:      "未知",
	}, nil
}

// RefreshStatus 刷新状态
func (a *App) RefreshStatus() error {
	// TODO: 实现状态刷新逻辑
	return nil
}

// StartDDNS 启动 DDNS 服务
func (a *App) StartDDNS(domainID int) error {
	// TODO: 实现 DDNS 服务启动逻辑
	a.addLog("info", "", fmt.Sprintf("启动 DDNS 服务 - 域名 ID: %d", domainID))
	return nil
}

// CheckAndUpdate 检查并更新 IP
func (a *App) CheckAndUpdate(domainID int) error {
	// TODO: 实现 IP 检查和更新逻辑
	a.addLog("info", "", fmt.Sprintf("检查并更新 IP - 域名 ID: %d", domainID))
	return nil
}
