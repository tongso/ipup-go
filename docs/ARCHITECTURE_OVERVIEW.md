# 项目架构说明

## 🏗️ 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (Vue 3 + TS)                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ StatusPanel  │  │ DomainManage │  │ Settings     │ │
│  │ 监控面板     │  │ 域名管理     │  │ 系统设置     │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
                          ↕ Wails Runtime
┌─────────────────────────────────────────────────────────┐
│                  Backend (Go 1.23)                      │
│  ┌──────────────────────────────────────────────────┐  │
│  │ app/ (API 暴露层)                                 │  │
│  │ - handlers_*.go (前端接口)                       │  │
│  │ - app.go (应用主结构)                            │  │
│  └──────────────────────────────────────────────────┘  │
│           ↕              ↕              ↕              │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐          │
│  │ monitor/ │   │ domain/  │   │ provider/│          │
│  │ 监控服务 │   │ 域名管理 │   │ DNS 服务商│          │
│  └──────────┘   └──────────┘   └──────────┘          │
│           ↕              ↕              ↕              │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐          │
│  │ log/     │   │ notify/  │   │ config/  │          │
│  │ 日志系统 │   │ 通知服务 │   │ 配置管理 │          │
│  └──────────┘   └──────────┘   └──────────┘          │
│                          ↕                             │
│  ┌──────────────────────────────────────────────────┐  │
│  │ database/ (SQLite)                                │  │
│  │ - domains, logs, settings, ip_history            │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 📁 核心模块说明

### 1. `internal/app/` - API 暴露层

**职责：** 提供前端可调用的 Go 方法

**关键文件：**
- `app.go` - 应用主结构，初始化各模块
- `handlers_domain.go` - 域名管理相关接口
- `handlers_config.go` - 配置管理接口
- `handlers_monitor.go` - 监控数据接口

**示例：**
```go
// app.go
type App struct {
    db       *database.Database
    logger   *log.Logger
    monitor  *monitor.MonitorService
    // ...
}

// GetDomains 获取所有域名列表
func (a *App) GetDomains() ([]types.Domain, error) {
    return a.db.GetDomains()
}
```

---

### 2. `internal/provider/` - DNS 服务提供商

**职责：** 封装各家 DNS 服务商的 API 调用细节

**支持的服务商：**
- ✅ **AliyunProvider** - 阿里云 DNS
- 🔲 TencentProvider - 腾讯云 DNSPod（待实现）
- 🔲 CloudflareProvider - Cloudflare（待实现）

**关键方法：**
```go
type DNSProvider interface {
    UpdateRecord(domain string, ip string) error
}

// AliyunProvider 实现
func (p *AliyunProvider) UpdateRecord(domain, ip string) error {
    recordID, err := p.describeDomainRecord(domainName, rr)
    if err != nil {
        return err
    }
    
    if recordID != "" {
        return p.updateDomainRecord(recordID, rr, ip)
    } else {
        return p.addDomainRecord(domainName, rr, ip)
    }
}
```

**代码位置：** `internal/provider/aliyun.go`

---

### 3. `internal/monitor/` - 监控服务

**职责：** 定时检查域名 IP 变化并触发更新

**核心功能：**
- 定时获取公网 IP（IPv4/IPv6）
- 比对数据库中的 IP 记录
- 发现变化时调用 DNS Provider 更新
- 生成详细日志

**关键方法：**
```go
type MonitorService struct {
    db      *database.Database
    logger  *log.Logger
    domains []DomainMonitor
}

func (m *MonitorService) Start() {
    for _, domain := range m.domains {
        go func(d DomainMonitor) {
            ticker := time.NewTicker(d.interval)
            for range ticker.C {
                m.checkDomain(d)
            }
        }(domain)
    }
}

func (m *MonitorService) checkDomain(d DomainMonitor) {
    // 1. 获取公网 IP
    currentIP, _ := m.getPublicIP()
    
    // 2. 查询数据库 IP
    storedIP, _ := m.db.GetLatestIP(d.domain.ID)
    
    // 3. 比对并更新
    if currentIP != storedIP {
        provider.UpdateRecord(d.domain.Name, currentIP)
    }
}
```

**代码位置：** `internal/monitor/checker.go`

---

### 4. `internal/log/` - 日志系统

**职责：** 记录所有操作的详细日志

**关键特性：**
- 支持多级别（INFO、SUCCESS、WARNING、ERROR）
- 时区转换（UTC → 用户配置的时区）
- 按域名、时间筛选
- 多种时间格式兼容解析

**日志结构：**
```go
type LogEntry struct {
    ID        int64     `json:"id"`
    Timestamp string    `json:"timestamp"`  // UTC 存储
    Level     string    `json:"level"`
    Domain    string    `json:"domain"`
    Action    string    `json:"action"`
    Message   string    `json:"message"`
    Details   string    `json:"details,omitempty"`
}
```

**时区转换逻辑：**
```go
func (l *Logger) convertTimezone(utcTimestamp string) string {
    // 尝试多种格式解析
    formats := []string{
        "2006-01-02 15:04:05",
        "2006-01-02T15:04:05Z",
        time.RFC3339,
    }
    
    // 解析成功后转换为配置时区
    loc, _ := time.LoadLocation(l.timezone)
    localTime := utcTime.In(loc)
    
    return localTime.Format("2006-01-02 15:04:05")
}
```

**代码位置：** `internal/log/logger.go`

---

### 5. `internal/database/` - 数据库操作

**职责：** SQLite 数据库 CRUD 操作

**数据表：**
- `domains` - 域名配置表
- `logs` - 日志表
- `settings` - 系统设置表
- `ip_history` - IP 历史记录表

**关键方法：**
```go
type Database struct {
    *gorm.DB
}

func (db *Database) GetDomains() ([]types.Domain, error) {
    var domains []types.Domain
    err := db.Find(&domains).Error
    return domains, err
}

func (db *Database) GetLatestByDomain(domainID int) (*Log, error) {
    var log Log
    err := db.Where("domain_id = ?", domainID).
             Order("timestamp DESC").
             First(&log).Error
    return &log, err
}
```

**代码位置：** `internal/database/database.go`

---

### 6. `pkg/types/` - 类型定义

**职责：** 定义全项目共享的数据结构

**关键类型：**
```go
// Domain 域名配置
type Domain struct {
    ID          int    `json:"id" gorm:"primaryKey"`
    Name        string `json:"name"`
    Provider    string `json:"provider"`  // aliyun, tencent, ...
    AccessKeyID string `json:"access_key_id"`
    Secret      string `json:"secret"`
    Enabled     bool   `json:"enabled"`
    Interval    int    `json:"interval"`  // 检查间隔（秒）
}

// Log 日志记录
type Log struct {
    ID        int64     `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`
    Domain    string    `json:"domain"`
    Action    string    `json:"action"`
    Message   string    `json:"message"`
}

// Settings 系统设置
type Settings struct {
    Timezone string `json:"timezone"`  // 时区设置
    Notify   bool   `json:"notify"`    // 是否启用通知
}
```

**代码位置：** `pkg/types/domain.go`, `pkg/types/log.go`, etc.

---

## 🔄 典型调用流程

### 手动更新 DNS 解析

```
用户点击"更新 DNS"按钮
    ↓
Frontend: updateDNS(domainId)
    ↓
Wails Runtime: window.go.app.App.UpdateDNS(domainId)
    ↓
Backend: internal/app/handlers_domain.go::UpdateDNS()
    ↓
internal/monitor/checker.go::checkDomain()
    ↓
internal/provider/aliyun.go::UpdateRecord()
    ↓
┌─────────────────────────────────┐
│ 1. describeDomainRecord()       │
│    - Action: DescribeDomainRecords
│    - 返回：RecordID             │
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ 2. modifyDomainRecord()         │
│    - Action: ModifyDomainRecord │
│    - 参数：RecordId, RR, Type, Value
│    - 返回：成功/失败            │
└─────────────────────────────────┘
    ↓
internal/log/logger.go::Log()
    ↓
database: 插入日志记录
    ↓
Frontend: 显示通知和状态更新
```

---

## 📊 数据流图

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   用户操作   │────▶│  Frontend   │────▶│   Wails     │
│ (点击按钮)  │     │  (Vue 3)    │     │  Runtime    │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   数据库    │◀────│   Logger    │◀────│   Handler   │
│  (SQLite)   │     │  (日志记录) │     │  (API 层)   │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ 阿里云 API   │◀────│  Provider   │◀────│  Monitor    │
│  (DNS 服务)  │     │  (API 封装) │     │  (调度器)   │
└─────────────┘     └─────────────┘     └─────────────┘
```

---

## 🎯 设计模式应用

### 1. **策略模式** - DNS Provider
```go
type DNSProvider interface {
    UpdateRecord(domain string, ip string) error
}

// 不同服务商的实现
type AliyunProvider struct{ ... }
type TencentProvider struct{ ... }
type CloudflareProvider struct{ ... }
```

### 2. **观察者模式** - 事件驱动
```go
// 前端事件
window.dispatchEvent(new CustomEvent('domain-updated', { detail: {...} }))

// 监听组件
window.addEventListener('domain-updated', (e) => {
    this.refreshData()
})
```

### 3. **工厂模式** - Provider 创建
```go
func NewProvider(domain types.Domain) (DNSProvider, error) {
    switch domain.Provider {
    case "aliyun":
        return NewAliyunProvider(domain.AccessKeyID, domain.Secret)
    case "tencent":
        return NewTencentProvider(...)
    default:
        return nil, errors.New("unsupported provider")
    }
}
```

---

## 🔒 安全机制

### 1. 敏感信息保护
```go
// 日志中隐藏 AccessKey
debugParams[k] = "***HIDDEN***"

// 数据库加密存储（可选）
encryptedSecret := encrypt(secret)
```

### 2. 签名验证
- HMAC-SHA1 签名算法
- 唯一 SignatureNonce 防重放
- Timestamp 时效性检查

### 3. 输入校验
```go
// 域名格式校验
if !isValidDomain(domain.Name) {
    return errors.New("invalid domain format")
}

// IP 格式校验
if net.ParseIP(ip) == nil {
    return errors.New("invalid IP address")
}
```

---

## 📈 性能优化

### 1. 并发获取 IP
```go
// IPv4 和 IPv6 并发请求
var ipv4, ipv6 string
var wg sync.WaitGroup

wg.Add(2)
go func() {
    defer wg.Done()
    ipv4, _ = getIPv4()
}()
go func() {
    defer wg.Done()
    ipv6, _ = getIPv6()
}()
wg.Wait()
```

### 2. 智能降级
```go
// 多个 IP 服务降级
services := []string{
    "https://api.ipify.org",
    "https://ident.me",
    "https://ifconfig.co",
}

for _, service := range services {
    if ip, err := getIP(service); err == nil {
        return ip, nil
    }
}
```

### 3. 连接池复用
```go
// HTTP Client 复用
client := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}
```

---

## 🧪 测试策略

### 1. 单元测试
```go
func TestGenerateSignature(t *testing.T) {
    provider := NewAliyunProvider("test-key", "test-secret")
    
    params := map[string]string{
        "Action": "DescribeDomainRecords",
        "Timestamp": "2026-03-06T01:51:27Z",
    }
    
    signature := provider.generateSignature(params)
    
    expected := "xxx"
    if signature != expected {
        t.Errorf("expected %s, got %s", expected, signature)
    }
}
```

### 2. 集成测试
```bash
# 测试脚本
go run scripts/test_aliyun_provider.go
```

### 3. E2E 测试
```typescript
// 前端测试
describe('Domain Management', () => {
    it('should update DNS successfully', async () => {
        await page.click('[data-testid="update-dns-btn"]')
        const notification = await page.waitForSelector('.notification-success')
        expect(notification).toBeTruthy()
    })
})
```

---

## 📚 相关文档

- [项目开发规范](./PROJECT_SPECIFICATIONS.md)
- [API 集成指南](./API_INTEGRATION_GUIDE.md)
- [踩坑记录](./LESSONS_LEARNED.md)
- [快速开始](../QUICKSTART.md)

---

*最后更新：2026-03-06*
