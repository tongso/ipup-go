package monitor

import (
	"fmt"
	"time"
	
	"ipup-go/internal/domain"
	"ipup-go/internal/log"
	"ipup-go/internal/provider"
	"ipup-go/pkg/types"
	"ipup-go/pkg/utils"
)

// Checker IP 检查器
type Checker struct {
	apiEndpoint string
}

// NewChecker 创建 IP 检查器
func NewChecker(apiEndpoint string) *Checker {
	return &Checker{apiEndpoint: apiEndpoint}
}

// GetPublicIP 获取公网 IP
func (c *Checker) GetPublicIP() (string, error) {
	ip, err := utils.GetPublicIP(c.apiEndpoint)
	if err != nil {
		return "", fmt.Errorf("获取公网 IP 失败：%w", err)
	}
	return ip, nil
}

// CheckIPChanged 检查 IP 是否变化
func (c *Checker) CheckIPChanged(currentIP string) (bool, error) {
	newIP, err := c.GetPublicIP()
	if err != nil {
		return false, err
	}
	
	return newIP != currentIP, nil
}

// MonitorService 监控服务
type MonitorService struct {
	domainRepo   *domain.Repository
	logger       *log.Logger
	checker      *Checker
	isRunning    bool
	stopChan     chan struct{}
}

// NewMonitorService 创建监控服务
func NewMonitorService(domainRepo *domain.Repository, logger *log.Logger, checker *Checker) *MonitorService {
	return &MonitorService{
		domainRepo: domainRepo,
		logger:     logger,
		checker:    checker,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动监控服务
func (m *MonitorService) Start(interval int) {
	if m.isRunning {
		fmt.Println("监控服务已在运行中")
		return
	}
	
	m.isRunning = true
	fmt.Printf("启动 IP 监控服务，检查间隔：%d 秒\n", interval)
	
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.checkAllDomains()
			case <-m.stopChan:
				fmt.Println("停止 IP 监控服务")
				m.isRunning = false
				return
			}
		}
	}()
}

// Stop 停止监控服务
func (m *MonitorService) Stop() {
	if !m.isRunning {
		return
	}
	close(m.stopChan)
}

// checkAllDomains 检查所有域名
func (m *MonitorService) checkAllDomains() {
	domains, err := m.domainRepo.ListEnabled()
	if err != nil {
		msg := fmt.Sprintf("获取启用域名失败：%v", err)
		fmt.Println(msg)
		if m.logger != nil {
			m.logger.Add("error", "", msg)
		}
		return
	}
	
	for _, d := range domains {
		m.checkDomain(d)
	}
}

// checkDomain 检查单个域名
func (m *MonitorService) checkDomain(d types.Domain) {
	logPrefix := fmt.Sprintf("[%s]", d.Domain)
	
	// 1. 获取当前公网 IP
	currentIP, err := m.checker.GetPublicIP()
	if err != nil {
		msg := fmt.Sprintf("获取公网 IP 失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		if m.logger != nil {
			m.logger.Add("error", d.Domain, msg)
		}
		return
	}
	
	// 2. 如果 IP 没有变化，跳过更新
	if d.CurrentIP == currentIP {
		fmt.Printf("%s IP 未变化：%s，跳过更新\n", logPrefix, currentIP)
		return
	}
	
	// 3. IP 发生变化，先更新数据库
	fmt.Printf("%s IP 已变化：%s -> %s，开始更新 DNS 解析\n", logPrefix, d.CurrentIP, currentIP)
	if err := m.domainRepo.UpdateIP(d.ID, currentIP); err != nil {
		msg := fmt.Sprintf("更新数据库 IP 失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		if m.logger != nil {
			m.logger.Add("error", d.Domain, msg)
		}
		return
	}
	
	// 4. 调用 DNS 提供商 API 更新解析记录
	m.updateDNSProvider(d, currentIP)
}

// updateDNSProvider 调用 DNS 提供商 API 更新解析
func (m *MonitorService) updateDNSProvider(d types.Domain, ip string) {
	logPrefix := fmt.Sprintf("[%s]", d.Domain)
	
	// 根据提供商类型创建对应的 Provider 实例
	p, err := provider.GetProvider(d.Provider, d.Domain, d.Token, d.AccessKeyID, d.AccessKeySecret)
	if err != nil {
		msg := fmt.Sprintf("获取 DNS 提供商失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		if m.logger != nil {
			m.logger.Add("error", d.Domain, msg)
		}
		return
	}
	
	// 调用 API 更新 DNS 记录
	err = p.UpdateRecord(d.Domain, ip)
	if err != nil {
		msg := fmt.Sprintf("调用%s API 更新 DNS 记录失败：%v", d.Provider, err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		if m.logger != nil {
			m.logger.Add("error", d.Domain, msg)
		}
		return
	}
	
	// 更新成功
	msg := fmt.Sprintf("成功调用%s API 更新 DNS 记录：%s -> %s", d.Provider, d.Domain, ip)
	fmt.Printf("%s %s\n", logPrefix, msg)
	if m.logger != nil {
		m.logger.Add("success", d.Domain, msg)
	}
}
