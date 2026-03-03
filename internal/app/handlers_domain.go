package app

import (
	"context"
	"fmt"
	"net"
	"time"
	
	"ipup-go/pkg/types"
)

// ==================== 域名管理 API ====================

// AddDomain 添加域名配置
func (a *App) AddDomain(d types.Domain) error {
	if d.Domain == "" {
		return fmt.Errorf("域名不能为空")
	}
	
	id, err := a.domainRepo.Create(d)
	if err != nil {
		return fmt.Errorf("添加域名失败：%w", err)
	}
	
	a.addLog("info", d.Domain, fmt.Sprintf("添加域名配置：%s (ID: %d)", d.Domain, id))
	return nil
}

// UpdateDomain 更新域名配置
func (a *App) UpdateDomain(d types.Domain) error {
	if err := a.domainRepo.Update(d); err != nil {
		return fmt.Errorf("更新域名失败：%w", err)
	}
	
	a.addLog("info", d.Domain, fmt.Sprintf("更新域名配置：%s", d.Domain))
	return nil
}

// DeleteDomain 删除域名配置
func (a *App) DeleteDomain(id int) error {
	if err := a.domainRepo.Delete(id); err != nil {
		return fmt.Errorf("删除域名失败：%w", err)
	}
	
	a.addLog("info", "", fmt.Sprintf("删除域名配置 (ID: %d)", id))
	return nil
}

// ListDomains 获取所有域名列表
func (a *App) ListDomains() ([]types.Domain, error) {
	domains, err := a.domainRepo.List()
	if err != nil {
		return nil, fmt.Errorf("查询域名列表失败：%w", err)
	}
	return domains, nil
}

// ToggleDomain 切换域名启用状态
func (a *App) ToggleDomain(id int) (bool, error) {
	newStatus, err := a.domainRepo.Toggle(id)
	if err != nil {
		return false, fmt.Errorf("切换状态失败：%w", err)
	}
	
	statusText := "禁用"
	if newStatus {
		statusText = "启用"
	}
	a.addLog("info", "", fmt.Sprintf("%s域名 (ID: %d)", statusText, id))
	
	return newStatus, nil
}

// GetDomainStatus 获取所有域名的状态（实时查询 DNS 解析）
func (a *App) GetDomainStatus() ([]types.DomainStatus, error) {
	domains, err := a.domainRepo.List()
	if err != nil {
		return nil, fmt.Errorf("查询域名列表失败：%w", err)
	}
	
	var statuses []types.DomainStatus
	
	for _, domain := range domains {
		if !domain.Enabled {
			continue
		}
		
		status := types.DomainStatus{
			Domain: domain.Domain,
		}
		
		// 实时查询域名的 DNS 解析记录
		currentIP, queryTime := a.queryDomainDNS(domain.Domain)
		
		if currentIP == "" {
			status.CurrentIP = ""
			status.LastUpdate = ""
			status.Status = "pending"
			status.Message = "等待首次更新"
		} else {
			status.CurrentIP = currentIP
			status.LastUpdate = queryTime
			status.Status = "success"
			status.Message = "解析正常"
		}
		
		statuses = append(statuses, status)
	}
	
	return statuses, nil
}

// queryDomainDNS 查询域名的 DNS 解析记录
func (a *App) queryDomainDNS(domainName string) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
