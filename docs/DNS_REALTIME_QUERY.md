# 域名状态显示修复 - 实时 DNS 查询

## 🎯 修复内容

### 1️⃣ 域名当前 IP 显示逻辑重构

#### ❌ 修复前（错误设计）
```go
// 从数据库读取
querySQL := `
SELECT domain, COALESCE(current_ip, ''), COALESCE(last_update, '')
FROM domains
WHERE enabled = 1
`

// 问题：
// 1. 显示的是数据库中保存的 IP（可能已过时）
// 2. 更新时间是最后写入数据库的时间
// 3. 无法反映域名真实的 DNS 解析状态
```

#### ✅ 修复后（正确设计）
```go
// 实时查询 DNS 解析
func (a *App) GetDomainStatus() ([]types.DomainStatus, error) {
    domains, err := a.domainRepo.List()
    
    for _, domain := range domains {
        // 实时查询域名的 DNS 解析记录
        currentIP, queryTime := a.queryDomainDNS(domain.Domain)
        
        status.CurrentIP = currentIP  // 域名实际指向的 IP
        status.LastUpdate = queryTime // 刚刚查询的时间
    }
}

// queryDomainDNS 实现 DNS 实时查询
func (a *App) queryDomainDNS(domainName string) (string, string) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // 使用系统 DNS 解析
    ips, err := net.DefaultResolver.LookupIPAddr(ctx, domainName)
    queryTime := time.Now().Format("2006-01-02 15:04:05")
    
    if err != nil || len(ips) == 0 {
        return "", "" // DNS 解析失败
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
```

---

### 2️⃣ IPv4/IPv6 双栈显示

公网 IP 地址卡片现在同时显示 IPv4 和 IPv6：

```vue
<!-- 显示 IPv4 和 IPv6 -->
<div class="dual-stack-ips" v-if="ipInfo.ipv4 || ipInfo.ipv6">
  <div class="ip-item" v-if="ipInfo.ipv4" title="IPv4 地址">
    <span class="ip-label">IPv4:</span>
    <span class="ip-value">{{ ipInfo.ipv4 }}</span>
  </div>
  <div class="ip-item" v-if="ipInfo.ipv6" title="IPv6 地址">
    <span class="ip-label">IPv6:</span>
    <span class="ip-value">{{ ipInfo.ipv6 }}</span>
  </div>
</div>
```

**后端支持**:
```go
func (a *App) GetPublicIP() (types.IPInfo, error) {
    ipv4, ipv6 := utils.GetDualStackIP()
    
    return types.IPInfo{
        PublicIP: ipv4,     // 主显示 IP（优先 IPv4）
        IPv4:     ipv4,     // 单独显示
        IPv6:     ipv6,     // 单独显示
        Location: "未知",
        ISP:      "未知",
    }, nil
}
```

---

## 📊 数据流程对比

### 修复前（数据库驱动）

```
用户刷新页面
    ↓
GetDomainStatus()
    ↓
查询数据库 domains 表
    ↓
读取 current_ip 字段（可能是旧的）
    ↓
读取 last_update 字段（写入时间）
    ↓
前端显示数据库中的 IP
```

**问题**: 
- ❌ IP 可能已过时（与真实 DNS 解析不一致）
- ❌ 无法及时发现 DNS 解析问题
- ❌ 依赖定时任务更新数据库

---

### 修复后（实时查询）

```
用户刷新页面
    ↓
GetDomainStatus()
    ↓
遍历所有启用的域名
    ↓
对每个域名执行 DNS 查询
    ↓
net.DefaultResolver.LookupIPAddr()
    ↓
获取域名实际指向的 IP
    ↓
记录查询时间
    ↓
前端显示真实的 DNS 解析结果
```

**优势**:
- ✅ IP 始终是最新的（实时从 DNS 服务商获取）
- ✅ 立即发现 DNS 解析问题
- ✅ 不依赖数据库存储
- ✅ 公网 IP 变化后立即查询 DNS API 更新

---

## 🎯 更新时机

### 1. 定时查询（自动）

```go
// 未来实现：后台定时任务
ticker := time.NewTicker(5 * time.Minute) // 每 5 分钟查询一次
go func() {
    for range ticker.C {
        // 刷新所有域名的 DNS 解析状态
        refreshAllDomainStatus()
    }
}()
```

### 2. 公网 IP 变化后（触发式）

```go
// 当检测到公网 IP 变化并更新 DNS 后
func CheckAndUpdate(domainID int) {
    // 1. 获取新的公网 IP
    newIP := GetPublicIP()
    
    // 2. 调用 DNS 服务商 API 更新
    UpdateDNSRecord(domain, newIP)
    
    // 3. 立即查询 DNS 确认更新成功
    currentIP, _ := queryDomainDNS(domain)
    
    // 4. 更新前端显示
    updateUI(currentIP)
}
```

### 3. 用户手动刷新

```
用户点击「刷新」按钮
    ↓
调用 GetDomainStatus()
    ↓
重新查询所有域名的 DNS
    ↓
更新界面显示
```

---

## 🧪 测试验证

### 运行测试脚本

```bash
go run scripts/test_dns_query.go
```

### 预期输出

```
🔍 测试域名 DNS 实时查询功能...
============================================================

📊 实时查询域名 DNS 解析:
-----------------------------------------------------------

🌐 域名：test155555555.xx.com
   ❌ DNS 解析失败
   状态：等待首次更新

🌐 域名：www.baidu.com
   ✅ 当前 IP: 180.101.49.44
   🕒 查询时间：2026-03-03 17:32:47
   状态：解析正常

🌐 域名：www.qq.com
   ✅ 当前 IP: 101.91.42.232
   🕒 查询时间：2026-03-03 17:32:47
   状态：解析正常

💡 说明:
   - 域名 IP 是实时从 DNS 服务商获取的
   - 不依赖数据库存储
   - 每次刷新都会重新查询
   - 公网 IP 变化后调用 DNS API 也会更新
```

---

## 📋 字段说明

### DomainStatus 结构

```go
type DomainStatus struct {
    Domain     string `json:"domain"`     // 域名
    CurrentIP  string `json:"currentIP"`  // ✅ 实时查询到的 IP
    LastUpdate string `json:"lastUpdate"` // ✅ 查询时间
    Status     string `json:"status"`     // pending/success/error
    Message    string `json:"message"`    // 状态描述
}
```

### 字段含义

| 字段 | 含义 | 数据来源 | 示例 |
|------|------|----------|------|
| `currentIP` | **域名当前指向的 IP** | 实时 DNS 查询 | `180.101.49.44` |
| `lastUpdate` | **查询到该 IP 的时间** | 查询时的系统时间 | `2026-03-03 17:32:47` |
| `status` | 解析状态 | 根据查询结果判断 | `success` |
| `message` | 状态描述 | 根据查询结果判断 | `解析正常` |

---

## 🎨 前端显示效果

### 场景 1: 新域名（未配置 DNS）

```
┌─────────────────────────────────────┐
│ 🌐 test.example.com                 │
│    当前 IP: (空)                     │
│                                     │
│    ⏳ 等待首次更新                  │
│    更新于：(空)                      │
└─────────────────────────────────────┘
```

**解读**: DNS 查询无结果，域名还未配置 A 记录

---

### 场景 2: 已配置的域名

```
┌─────────────────────────────────────┐
│ 🌐 www.baidu.com                    │
│    当前 IP: ✅ 180.101.49.44         │
│                                     │
│    ✅ 解析正常                      │
│    更新于：2026-03-03 17:32:47      │
└─────────────────────────────────────┘
```

**解读**: DNS 查询成功，显示真实的解析 IP 和查询时间

---

### 场景 3: 公网 IP 卡片（双栈显示）

```
┌─────────────────────────────────────┐
│ 🌍 公网 IP 地址          [🔄 刷新]  │
│                                     │
│         123.45.67.89                │
│                                     │
│    ┌──────────────────────────┐    │
│    │ IPv4: 123.45.67.89       │    │
│    │ IPv6: 240e:xxx:xxx:xxx   │    │
│    └──────────────────────────┘    │
└─────────────────────────────────────┘
```

**解读**: 同时显示 IPv4 和 IPv6 地址

---

## 🔧 技术实现细节

### DNS 查询超时控制

```go
// 设置 5 秒超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 使用系统 DNS 解析器
ips, err := net.DefaultResolver.LookupIPAddr(ctx, domainName)
```

### IPv4 优先策略

```go
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
```

### 错误处理

```go
if err != nil || len(ips) == 0 {
    // DNS 解析失败或无结果
    return "", ""
}
```

---

## 🚀 后续优化建议

### 1. DNS 缓存（可选）

虽然现在是实时查询，但可以考虑添加短期缓存提高性能：

```go
var dnsCache = make(map[string]DNSResult)

func queryDomainDNS(domainName string) (string, string) {
    // 检查缓存（5 分钟内有效）
    if result, ok := dnsCache[domainName]; ok {
        if time.Since(result.QueryTime) < 5*time.Minute {
            return result.IP, result.QueryTime.Format("2006-01-02 15:04:05")
        }
    }
    
    // 执行 DNS 查询
    ip, queryTime := queryFromDNS(domainName)
    
    // 更新缓存
    dnsCache[domainName] = DNSResult{
        IP:        ip,
        QueryTime: queryTime,
    }
    
    return ip, queryTime.Format("2006-01-02 15:04:05")
}
```

### 2. 自定义 DNS 服务器

```go
// 使用公共 DNS 服务器（如 8.8.8.8）
resolver := &net.Resolver{
    PreferGo: true,
    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
        d := net.Dialer{Timeout: time.Second * 5}
        return d.DialContext(ctx, network, "8.8.8.8:53")
    },
}

ips, err := resolver.LookupIPAddr(ctx, domainName)
```

### 3. 并发查询多个域名

```go
func GetDomainStatus() ([]types.DomainStatus, error) {
    domains, _ := a.domainRepo.List()
    
    var wg sync.WaitGroup
    results := make(chan types.DomainStatus, len(domains))
    
    for _, domain := range domains {
        wg.Add(1)
        go func(d Domain) {
            defer wg.Done()
            ip, time := queryDomainDNS(d.Domain)
            results <- buildStatus(d, ip, time)
        }(domain)
    }
    
    wg.Wait()
    close(results)
    
    // 收集结果
    var statuses []types.DomainStatus
    for result := range results {
        statuses = append(statuses, result)
    }
    
    return statuses, nil
}
```

---

## 📝 修改的文件

1. ✅ [`internal/app/handlers_monitor.go`](../../internal/app/handlers_monitor.go)
   - 重写 `GetDomainStatus()` 方法
   - 新增 `queryDomainDNS()` 方法
   - 实现实时 DNS 查询逻辑

2. ✅ [`internal/domain/repository.go`](../../internal/domain/repository.go)
   - 删除旧的 `GetStatus()` 方法

3. ✅ [`scripts/test_dns_query.go`](../../scripts/test_dns_query.go)
   - 新增 DNS 实时查询测试脚本

4. ✅ [`docs/DNS_REALTIME_QUERY.md`](./DNS_REALTIME_QUERY.md)
   - 详细说明文档

---

## ✅ 总结

### 核心改进

1. **实时性**: 域名 IP 不再依赖数据库，而是实时从 DNS 服务商获取
2. **准确性**: 显示的就是域名当前实际指向的 IP，便于用户对比
3. **及时性**: 公网 IP 变化后，调用 DNS API 立即可见
4. **双栈支持**: IPv4 和 IPv6 同时显示，不会遗漏

### 用户体验提升

- ✅ 用户可以立即看到域名当前的真实 IP
- ✅ 方便对比公网 IP 和域名解析 IP 是否一致
- ✅ 发现问题可以立即刷新，无需等待定时任务
- ✅ IPv4/IPv6信息完整显示，一目了然

---

**修复日期**: 2026-03-03  
**相关文档**: [`scripts/test_dns_query.go`](../scripts/test_dns_query.go)  
**后端实现**: [`internal/app/handlers_monitor.go`](../../internal/app/handlers_monitor.go)
