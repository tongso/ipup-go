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
	timers       map[int]*time.Ticker  // 每个域名的独立定时器
	timerStopChans map[int]chan struct{}  // 每个定时器的停止通道
}

// NewMonitorService 创建监控服务
func NewMonitorService(domainRepo *domain.Repository, logger *log.Logger, checker *Checker) *MonitorService {
	return &MonitorService{
		domainRepo: domainRepo,
		logger:     logger,
		checker:    checker,
		stopChan:   make(chan struct{}),
		timers:     make(map[int]*time.Ticker),
		timerStopChans: make(map[int]chan struct{}),
	}
}

// Start 启动监控服务（为每个启用的域名创建独立定时器）
func (m *MonitorService) Start() {
	if m.isRunning {
		fmt.Println("监控服务已在运行中")
		return
	}
	
	m.isRunning = true
	fmt.Println("启动 IP 监控服务，为每个域名创建独立定时器...")
	
	// 加载所有启用的域名并为每个创建定时器
	m.refreshAllTimers()
}

// Stop 停止监控服务
func (m *MonitorService) Stop() {
	if !m.isRunning {
		return
	}
	
	// 停止所有定时器
	for domainID, ticker := range m.timers {
		if stopChan, exists := m.timerStopChans[domainID]; exists {
			close(stopChan)
			ticker.Stop()
			delete(m.timers, domainID)
			delete(m.timerStopChans, domainID)
		}
	}
	
	m.isRunning = false
	fmt.Println("已停止所有域名的监控服务")
}

// RefreshTimers 刷新所有定时器（当域名配置变更时调用）
func (m *MonitorService) RefreshTimers() {
	if !m.isRunning {
		return
	}
	m.refreshAllTimers()
}

// refreshAllTimers 重新加载所有域名并更新定时器
func (m *MonitorService) refreshAllTimers() {
	// 先停止所有现有定时器
	for domainID, ticker := range m.timers {
		if stopChan, exists := m.timerStopChans[domainID]; exists {
			close(stopChan)
			ticker.Stop()
			delete(m.timers, domainID)
			delete(m.timerStopChans, domainID)
		}
	}
	
	// 获取所有启用的域名
	domains, err := m.domainRepo.ListEnabled()
	if err != nil {
		msg := fmt.Sprintf("获取启用域名失败：%v", err)
		fmt.Println(msg)
		if m.logger != nil {
			m.logger.Add("error", "", msg)
		}
		return
	}
	
	// 为每个域名创建独立定时器
	for _, d := range domains {
		interval := d.Interval
		if interval <= 0 {
			interval = 300 // 默认 5 分钟
		}
		
		fmt.Printf("[%s] 创建定时器，间隔：%d秒\n", d.Domain, interval)
		m.createTimer(d, interval)
	}
	
	fmt.Printf("已为 %d 个域名创建独立定时器\n", len(m.timers))
}

// createTimer 为单个域名创建定时器
func (m *MonitorService) createTimer(d types.Domain, intervalSeconds int) {
	stopChan := make(chan struct{})
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	
	m.timers[d.ID] = ticker
	m.timerStopChans[d.ID] = stopChan
	
	// 立即执行一次
	go func(domain types.Domain) {
		m.checkDomain(domain)
	}(d)
	
	// 定时执行
	go func(domain types.Domain) {
		for {
			select {
			case <-ticker.C:
				m.checkDomain(domain)
			case <-stopChan:
				fmt.Printf("[%s] 停止定时器\n", domain.Domain)
				return
			}
		}
	}(d)
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
	
	// 1. 获取当前公网IP
	currentIP, err := m.checker.GetPublicIP()
	if err != nil {
		msg := fmt.Sprintf("获取公网IP 失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		if m.logger != nil {
			m.logger.Add("error", d.Domain, msg)
		}
		return
	}
	
	// 2. 验证 DNS 解析是否存在（通过查询 DNS 提供商 API）
	dnsExists, dnsIP := m.checkDNSRecordExists(d)
	
	if !dnsExists {
		// DNS 解析不存在，直接创建
		fmt.Printf("%s DNS 解析不存在，开始创建新解析\n", logPrefix)
		
		// 更新数据库 IP
		if err := m.domainRepo.UpdateIP(d.ID, currentIP); err != nil {
			msg := fmt.Sprintf("更新数据库 IP 失败：%v", err)
			fmt.Printf("%s %s\n", logPrefix, msg)
			if m.logger != nil {
				m.logger.Add("error", d.Domain, msg)
			}
			return
		}
		
		// 调用 DNS 提供商 API 创建解析记录
		m.createDNSProvider(d, currentIP)
		return
	}
	
	// 3. DNS 解析存在，检查 IP 是否变化
	if dnsIP == currentIP {
		fmt.Printf("%s IP 未变化：%s，跳过更新\n", logPrefix, currentIP)
		return
	}
	
	// 4. IP 发生变化，先更新数据库
	fmt.Printf("%s IP 已变化：%s -> %s，开始更新 DNS 解析\n", logPrefix, dnsIP, currentIP)
	if err := m.domainRepo.UpdateIP(d.ID, currentIP); err != nil {
		msg := fmt.Sprintf("更新数据库 IP 失败：%v", err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		if m.logger != nil {
			m.logger.Add("error", d.Domain, msg)
		}
		return
	}
	
	// 5. 调用 DNS 提供商 API 更新解析记录
	m.updateDNSProvider(d, currentIP)
}

// checkDNSRecordExists 检查 DNS 解析记录是否存在
func (m *MonitorService) checkDNSRecordExists(d types.Domain) (bool, string) {
	// 根据提供商类型创建对应的 Provider 实例
	p, err := provider.GetProvider(d.Provider, d.Domain, d.Token, d.AccessKeyID, d.AccessKeySecret)
	if err != nil {
		fmt.Printf("[%s] 获取 DNS 提供商失败：%v\n", d.Domain, err)
		return false, ""
	}
	
	// 调用 Provider 的 GetRecord 方法查询 DNS 解析
	ip, err := p.GetRecord(d.Domain)
	if err != nil {
		// 查询失败，认为解析不存在
		fmt.Printf("[%s] 查询 DNS 解析失败：%v\n", d.Domain, err)
		return false, ""
	}
	
	if ip == "" {
		// 返回空字符串，认为解析不存在
		fmt.Printf("[%s] DNS 解析记录不存在\n", d.Domain)
		return false, ""
	}
	
	// 解析存在，返回 true 和当前解析的 IP
	fmt.Printf("[%s] DNS 解析记录存在，当前 IP: %s\n", d.Domain, ip)
	return true, ip
}

// createDNSProvider 调用 DNS 提供商 API 创建解析
func (m *MonitorService) createDNSProvider(d types.Domain, ip string) {
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
	
	// 调用 API 创建 DNS 记录
	err = p.UpdateRecord(d.Domain, ip)
	if err != nil {
		msg := fmt.Sprintf("调用%s API 创建 DNS 记录失败：%v", d.Provider, err)
		fmt.Printf("%s %s\n", logPrefix, msg)
		if m.logger != nil {
			m.logger.Add("error", d.Domain, msg)
		}
		return
	}
	
	// 创建成功
	msg := fmt.Sprintf("成功调用%s API 创建 DNS 记录：%s -> %s", d.Provider, d.Domain, ip)
	fmt.Printf("%s %s\n", logPrefix, msg)
	if m.logger != nil {
		m.logger.Add("success", d.Domain, msg)
	}
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
