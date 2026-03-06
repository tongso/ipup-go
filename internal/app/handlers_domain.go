package app

import (
	"context"
	"fmt"
	"net"
	"time"
	
	"ipup-go/internal/provider"
	"ipup-go/pkg/types"
	"ipup-go/pkg/utils"
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
			ID:         domain.ID,
			Domain:     domain.Domain,
			Provider:   domain.Provider,
			CurrentIP:  domain.CurrentIP,
			LastUpdate: domain.LastUpdate,
		}
		
		// 注意：这里需要将 domain.ID 转换为 int 类型赋给 status
		// 由于 DomainStatus 结构中没有 ID 字段，我们需要通过其他方式传递
		// 暂时在 message 中携带 ID 信息，或者修改 DomainStatus 结构
		
		// 获取最近的 API 调用日志
		apiLogs, err := a.logger.Get("", domain.Domain, 1)
		if err == nil && len(apiLogs) > 0 {
			latestLog := apiLogs[0]
			status.LastAPICall = latestLog.Timestamp
			status.APIMessage = latestLog.Message
			
			// 根据日志级别设置状态
			switch latestLog.Level {
			case "success":
				status.APIStatus = "success"
			case "error":
				status.APIStatus = "error"
			default:
				status.APIStatus = "pending"
			}
		} else {
			status.APIStatus = "pending"
			status.APIMessage = "暂无 API 调用记录"
		}
		
		// 如果没有 CurrentIP，设置为等待首次更新
		if domain.CurrentIP == "" {
			status.Status = "pending"
			status.Message = "等待首次更新"
		} else {
			// 实时查询域名的 DNS 解析记录进行验证
			currentIP, queryTime := a.queryDomainDNS(domain.Domain)
			if currentIP == "" {
				status.Status = "warning"
				status.Message = "DNS 解析失败或未配置"
			} else if currentIP == domain.CurrentIP {
				status.Status = "success"
				status.Message = "解析正常"
				status.LastUpdate = queryTime
			} else {
				status.Status = "warning"
				status.Message = fmt.Sprintf("数据库 IP(%s) 与 DNS 解析不一致 (%s)", domain.CurrentIP, currentIP)
			}
		}
		
		statuses = append(statuses, status)
	}
	
	return statuses, nil
}

// UpdateDomainDNS 手动触发更新域名的 DNS 解析
func (a *App) UpdateDomainDNS(domainID int) (string, error) {
	// 获取域名配置
	domain, err := a.domainRepo.GetByID(domainID)
	if err != nil {
		return "", fmt.Errorf("获取域名配置失败：%w", err)
	}
	
	logPrefix := fmt.Sprintf("[%s]", domain.Domain)
	fmt.Printf("%s 开始手动更新 DNS 解析\n", logPrefix)
	
	// 1. 获取当前公网 IP
	currentIP, err := utils.GetPublicIP("")
	if err != nil {
		msg := fmt.Sprintf("获取公网 IP 失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		a.logger.Add("error", domain.Domain, msg)
		return "", fmt.Errorf(msg)
	}
	
	fmt.Printf("%s 获取到公网 IP: %s\n", logPrefix, currentIP)
	
	// 2. 更新数据库中的 IP
	if err := a.domainRepo.UpdateIP(domainID, currentIP); err != nil {
		msg := fmt.Sprintf("更新数据库 IP 失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		a.logger.Add("error", domain.Domain, msg)
		return "", fmt.Errorf(msg)
	}
	
	fmt.Printf("%s 数据库 IP 已更新：%s\n", logPrefix, currentIP)
	
	// 3. 调用 DNS 提供商 API 更新解析
	p, err := provider.GetProvider(domain.Provider, domain.Domain, domain.Token, domain.AccessKeyID, domain.AccessKeySecret)
	if err != nil {
		msg := fmt.Sprintf("获取 DNS 提供商失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		a.logger.Add("error", domain.Domain, msg)
		return "", fmt.Errorf(msg)
	}
	
	// 4. 执行 API 调用
	err = p.UpdateRecord(domain.Domain, currentIP)
	if err != nil {
		msg := fmt.Sprintf("调用%s API 更新 DNS 记录失败：%v", domain.Provider, err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		a.logger.Add("error", domain.Domain, msg)
		return "", fmt.Errorf(msg)
	}
	
	// 5. 成功
	successMsg := fmt.Sprintf("✅ 成功调用%s API 更新 DNS 记录：%s -> %s", domain.Provider, domain.Domain, currentIP)
	fmt.Printf("%s %s\n", logPrefix, successMsg)
	a.logger.Add("success", domain.Domain, successMsg)
	
	return successMsg, nil
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
