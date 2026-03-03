package monitor

import (
	"fmt"
	"time"
	
	"ipup-go/internal/domain"
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
	checker      *Checker
	isRunning    bool
	stopChan     chan struct{}
}

// NewMonitorService 创建监控服务
func NewMonitorService(domainRepo *domain.Repository, checker *Checker) *MonitorService {
	return &MonitorService{
		domainRepo: domainRepo,
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
		fmt.Printf("获取启用域名失败：%v\n", err)
		return
	}
	
	for _, d := range domains {
		m.checkDomain(d)
	}
}

// checkDomain 检查单个域名
func (m *MonitorService) checkDomain(d types.Domain) {
	currentIP, err := m.checker.GetPublicIP()
	if err != nil {
		fmt.Printf("[%s] 获取 IP 失败：%v\n", d.Domain, err)
		return
	}
	
	// 如果 IP 没有变化，跳过
	if d.CurrentIP == currentIP {
		fmt.Printf("[%s] IP 未变化：%s\n", d.Domain, currentIP)
		return
	}
	
	// IP 发生变化，更新数据库
	fmt.Printf("[%s] IP 已变化：%s -> %s\n", d.Domain, d.CurrentIP, currentIP)
	if err := m.domainRepo.UpdateIP(d.ID, currentIP); err != nil {
		fmt.Printf("[%s] 更新 IP 失败：%v\n", d.Domain, err)
		return
	}
	
	fmt.Printf("[%s] IP 更新成功\n", d.Domain)
}
