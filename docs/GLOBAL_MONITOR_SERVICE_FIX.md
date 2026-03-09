# 全局定时任务架构重构

## 🎯 问题描述

**用户需求**：程序只要开着，定时任务就要一直执行，不受页面切换的影响。

**原有问题**：
1. ❌ 定时器在前端的 `DomainList.vue` 组件中创建
2. ❌ 当用户切换到其他页面（如日志、设置）时，组件销毁，定时器被清除
3. ❌ 返回域名管理页面时，定时器重新创建，导致更新中断或重复

---

## 🔍 问题分析

### 修复前的架构（❌ 错误）

```
前端 DomainList.vue 组件
    ↓
onMounted() → 创建定时器
    ↓
setInterval(UpdateDomainDNS)
    ↓
onUnmounted() → 清除定时器
    ↓
用户切换页面 → 组件销毁 → 定时器停止 ❌
```

**问题场景**：
```
用户在"域名管理"页面
    ↓
定时器正常工作（每 5 分钟更新一次）
    ↓
用户点击"日志查看"标签
    ↓
DomainList.vue 组件销毁
    ↓
定时器被清除 ❌
    ↓
域名不再自动更新 ❌
```

### 修复后的架构（✅ 正确）

```
后端 Go MonitorService（全局服务）
    ↓
Startup() → 启动监控服务
    ↓
为每个域名创建独立 goroutine + ticker
    ↓
持续运行，不受前端页面影响 ✅
    ↓
Shutdown() → 停止所有定时器
```

**优势**：
- ✅ 定时器在后端 Go 服务中运行
- ✅ 应用启动时自动创建，关闭时自动销毁
- ✅ 不受前端页面切换影响
- ✅ 每个域名独立的定时器和 goroutine

---

## ✅ 修复方案

### 核心改动总览

| 模块 | 修改内容 | 目的 |
|------|---------|------|
| **后端** | `internal/monitor/checker.go` | 重构 MonitorService，支持多定时器 |
| **后端** | `internal/app/app.go` | Startup 时自动启动监控服务 |
| **后端** | `internal/app/handlers_domain.go` | 配置变更时刷新定时器 |
| **前端** | `frontend/src/components/DomainList.vue` | 移除定时器逻辑，只负责 UI |

---

### 1. 后端修改：MonitorService 重构

**文件**: `internal/monitor/checker.go`

#### 修改前（单一全局定时器）

```go
type MonitorService struct {
    domainRepo   *domain.Repository
    logger       *log.Logger
    checker      *Checker
    isRunning    bool
    stopChan     chan struct{}
}

// Start 使用统一的全局间隔
func (m *MonitorService) Start(interval int) {
    ticker := time.NewTicker(time.Duration(interval) * time.Second)
    go func() {
        for {
            select {
            case <-ticker.C:
                m.checkAllDomains()  // 所有域名用同一个间隔
            case <-m.stopChan:
                return
            }
        }
    }()
}
```

#### 修改后（多定时器独立管理）

```go
type MonitorService struct {
    domainRepo   *domain.Repository
    logger       *log.Logger
    checker      *Checker
    isRunning    bool
    stopChan     chan struct{}
    timers       map[int]*time.Ticker           // 每个域名的独立定时器
    timerStopChans map[int]chan struct{}        // 每个定时器的停止通道
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
```

**关键改进**：
1. ✅ 每个域名有独立的 `ticker` 和 `stopChan`
2. ✅ 每个域名在独立的 goroutine 中运行
3. ✅ 支持动态添加/删除/修改定时器
4. ✅ 应用启动时自动创建，关闭时自动销毁

---

### 2. 后端修改：应用启动时自动启动监控服务

**文件**: `internal/app/app.go`

#### 修改前

```go
// Startup 应用启动时初始化
func (a *App) Startup(ctx context.Context) error {
    // ... 初始化数据库和配置 ...
    
    // 初始化监控服务
    checker := monitor.NewChecker("")
    a.monitorSvc = monitor.NewMonitorService(a.domainRepo, a.logger, checker)
    
    // ❌ 需要手动调用 StartDDNS() 才能启动
    a.addLog("info", "", "应用初始化完成")
    return nil
}
```

#### 修改后

```go
// Startup 应用启动时初始化
func (a *App) Startup(ctx context.Context) error {
    // ... 初始化数据库和配置 ...
    
    // 初始化监控服务
    checker := monitor.NewChecker("")
    a.monitorSvc = monitor.NewMonitorService(a.domainRepo, a.logger, checker)
    
    // ✅ 自动启动监控服务（为每个启用的域名创建独立定时器）
    a.monitorSvc.Start()
    
    a.addLog("info", "", "应用初始化完成，监控服务已启动")
    return nil
}
```

**优势**：
- ✅ 应用一启动就自动开始监控
- ✅ 不需要前端手动触发
- ✅ 完全自动化

---

### 3. 后端修改：配置变更时刷新定时器

**文件**: `internal/app/handlers_domain.go`

#### 新增逻辑

```go
// UpdateDomain 更新域名配置
func (a *App) UpdateDomain(d types.Domain) error {
    if err := a.domainRepo.Update(d); err != nil {
        return fmt.Errorf("更新域名失败：%w", err)
    }
    
    a.addLog("info", d.Domain, fmt.Sprintf("更新域名配置：%s", d.Domain))
    
    // ✅ 刷新监控服务的定时器（根据新的 interval 重新创建）
    if a.monitorSvc != nil {
        a.monitorSvc.RefreshTimers()
    }
    
    return nil
}

// DeleteDomain 删除域名配置
func (a *App) DeleteDomain(id int) error {
    if err := a.domainRepo.Delete(id); err != nil {
        return fmt.Errorf("删除域名失败：%w", err)
    }
    
    a.addLog("info", "", fmt.Sprintf("删除域名配置 (ID: %d)", id))
    
    // ✅ 刷新监控服务的定时器（删除对应的定时器）
    if a.monitorSvc != nil {
        a.monitorSvc.RefreshTimers()
    }
    
    return nil
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
    
    // ✅ 刷新监控服务的定时器（启用则创建，禁用则删除）
    if a.monitorSvc != nil {
        a.monitorSvc.RefreshTimers()
    }
    
    return newStatus, nil
}
```

**作用**：
- ✅ 当域名配置变更时，自动重新加载定时器
- ✅ 修改 interval → 按新间隔重新创建定时器
- ✅ 删除域名 → 删除对应定时器
- ✅ 切换启用状态 → 创建或删除定时器

---

### 4. 前端修改：移除定时器逻辑

**文件**: `frontend/src/components/DomainList.vue`

#### 修改前

```typescript
import { ref, reactive, watch, onMounted, onUnmounted } from 'vue'

// 存储每个域名的刷新定时器
const refreshTimers = new Map<number, number>()

// 为单个域名创建刷新定时器
const createRefreshTimer = (domainId: number, intervalSeconds: number) => {
  const oldTimer = refreshTimers.get(domainId)
  if (oldTimer) {
    clearInterval(oldTimer)
  }
  
  const intervalMs = intervalSeconds * 1000
  const timer = window.setInterval(async () => {
    console.log(`⏰ 自动刷新域名 ID ${domainId} 的状态...`)
    
    const domain = domains.value.find(d => d.id === domainId)
    if (!domain || !domain.enabled) {
      return
    }
    
    try {
      const result = await UpdateDomainDNS(domainId)
      notifySuccess(`${domain.domain}: ${result}`)
      await refreshSingleDomainStatus(domainId)
      window.dispatchEvent(new CustomEvent('domains-updated'))
    } catch (error) {
      // 错误处理
    }
  }, intervalMs)
  
  refreshTimers.set(domainId, timer)
}

// 组件卸载时清除所有定时器
onUnmounted(() => {
  refreshTimers.forEach((timer, id) => {
    clearInterval(timer)
  })
  refreshTimers.clear()
  console.log('🗑️ 已清除所有域名刷新定时器')
})
```

#### 修改后

```typescript
import { ref, reactive, watch, onMounted } from 'vue'


// 组件挂载时加载数据
onMounted(() => {
  loadDomains()
})

// ✅ 定时器已经移到后端全局管理，前端只负责显示
```

**变化**：
- ❌ 删除所有定时器相关代码
- ❌ 删除 `onUnmounted` 生命周期
- ✅ 只保留数据加载和 UI 展示
- ✅ 定时器由后端全局管理

---

## 📊 架构对比

### 修复前（前端定时器）

```
┌─────────────────────────────────────┐
│  前端 Vue 应用                      │
│  ┌───────────────────────────────┐  │
│  │ DomainList.vue 组件           │  │
│  │  - setInterval()             │  │
│  │  - 定时器                    │  │
│  └───────────────────────────────┘  │
│         ↓ 组件销毁                  │
│         ↓ 定时器清除 ❌              │
└─────────────────────────────────────┘
```

### 修复后（后端全局服务）

```
┌─────────────────────────────────────┐
│  前端 Vue 应用                      │
│  - DomainList.vue (仅 UI)          │
│  - StatusPanel.vue                 │
│  - Logs.vue                        │
└─────────────────────────────────────┘
         ↕ Wails Runtime
┌─────────────────────────────────────┐
│  后端 Go 应用                       │
│  ┌───────────────────────────────┐  │
│  │ MonitorService (全局服务)     │  │
│  │  - Timer 1: example.com (60s) │  │
│  │  - Timer 2: test.com (300s)   │  │
│  │  - Timer 3: demo.com (600s)   │  │
│  │  - 持续运行 ✅                 │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

---

## 🧪 测试验证步骤

### 场景 1: 应用启动后自动开始监控

**步骤**：
1. 启动应用：`wails dev`
2. 查看控制台日志

**预期输出**：
```
启动 IP 监控服务，为每个域名创建独立定时器...
[ddns.jizhangxia.com] 创建定时器，间隔：300 秒
[test.example.com] 创建定时器，间隔：60 秒
已为 2 个域名创建独立定时器
应用初始化完成，监控服务已启动
```

**验证点**：
- ✅ 应用启动时自动创建定时器
- ✅ 每个域名根据自己的 interval 创建定时器

---

### 场景 2: 切换页面不影响定时任务

**步骤**：
1. 在"域名管理"页面添加域名 `test.com`，interval=60
2. 切换到"日志查看"页面
3. 等待 2 分钟
4. 查看日志列表

**预期结果**：
```
[2026-03-09 15:30:00] [success] test.com: 成功调用 Aliyun API 更新 DNS 记录
[2026-03-09 15:29:00] [info] test.com: IP 未变化，跳过更新
[2026-03-09 15:28:00] [info] test.com: IP 未变化，跳过更新
```

**验证点**：
- ✅ 即使不在"域名管理"页面，定时器仍在运行
- ✅ 日志持续生成，证明定时任务正常工作

---

### 场景 3: 修改 interval 后立即生效

**步骤**：
1. 编辑域名，将 interval 从 300 改为 60
2. 保存配置
3. 查看控制台日志

**预期输出**：
```
[ddns.jizhangxia.com] 创建定时器，间隔：60 秒
已为 1 个域名创建独立定时器
```

**验证点**：
- ✅ 旧定时器被清除
- ✅ 新定时器按新间隔创建

---

### 场景 4: 删除域名后定时器停止

**步骤**：
1. 删除域名 `test.com`
2. 查看控制台日志

**预期输出**：
```
[test.com] 停止定时器
已为 0 个域名创建独立定时器
```

**验证点**：
- ✅ 删除域名后，对应定时器停止
- ✅ 内存中没有残留的定时器

---

### 场景 5: 禁用域名后定时器停止

**步骤**：
1. 禁用域名 `ddns.jizhangxia.com`
2. 查看控制台日志

**预期输出**：
```
[ddns.jizhangxia.com] 停止定时器
已为 0 个域名创建独立定时器
```

**验证点**：
- ✅ 禁用的域名不会继续定时更新
- ✅ 重新启用后会重新创建定时器

---

## 💡 优势分析

### 修复前的问题

| 问题 | 影响 |
|------|------|
| ❌ 页面切换导致定时器停止 | 用户必须停留在域名管理页面 |
| ❌ 定时器在组件生命周期内管理 | 组件销毁=定时器销毁 |
| ❌ 前端占用资源 | 浏览器需要维护多个定时器 |
| ❌ 不可靠 | 依赖前端 JavaScript 引擎 |

### 修复后的优势

| 优势 | 说明 |
|------|------|
| ✅ **全局运行** | 定时器在后端 Go 服务中，不受前端影响 |
| ✅ **应用级生命周期** | 启动时创建，关闭时销毁 |
| ✅ **高性能** | Go 的 goroutine 轻量级，资源占用少 |
| ✅ **可靠性高** | 不依赖前端，纯后端实现 |
| ✅ **灵活配置** | 每个域名独立间隔，动态调整 |
| ✅ **易于扩展** | 支持任意数量的域名 |

---

## 📝 相关文件

### 修改的文件

1. ✅ [`internal/monitor/checker.go`](../internal/monitor/checker.go)
   - 重构 MonitorService 结构
   - 新增 `timers` 和 `timerStopChans` 字段
   - 新增 `refreshAllTimers()` 方法
   - 新增 `createTimer()` 方法
   - 修改 `Start()` 和 `Stop()` 方法

2. ✅ [`internal/app/app.go`](../internal/app/app.go)
   - 在 `Startup()` 中自动调用 `monitorSvc.Start()`

3. ✅ [`internal/app/handlers_domain.go`](../internal/app/handlers_domain.go)
   - 在 `UpdateDomain()` 中调用 `RefreshTimers()`
   - 在 `DeleteDomain()` 中调用 `RefreshTimers()`
   - 在 `ToggleDomain()` 中调用 `RefreshTimers()`

4. ✅ [`frontend/src/components/DomainList.vue`](../frontend/src/components/DomainList.vue)
   - 移除所有定时器相关代码
   - 移除 `onUnmounted` 生命周期
   - 简化为纯 UI 组件

---

## 🚀 后续优化建议

### 1. 增加定时器状态查询

提供前端接口查询当前运行的定时器：

```go
// GetTimerStatus 获取定时器状态
func (a *App) GetTimerStatus() ([]types.TimerStatus, error) {
    statuses := []types.TimerStatus{}
    for domainID, ticker := range a.monitorSvc.timers {
        statuses = append(statuses, types.TimerStatus{
            DomainID: domainID,
            Running:  ticker != nil,
            // ...
        })
    }
    return statuses, nil
}
```

### 2. 增加手动刷新定时器按钮

在前端添加"立即刷新定时器"按钮：

```typescript
const refreshTimers = async () => {
  await RefreshDomainTimers()  // 调用后端接口
  notifySuccess('定时器已刷新')
}
```

### 3. 增加定时器调试日志

在开发模式下输出更详细的定时器日志：

```go
if os.Getenv("DEBUG") == "true" {
    fmt.Printf("[DEBUG] [%s] 定时器已创建，下次执行时间：%v\n", 
        d.Domain, time.Now().Add(time.Duration(intervalSeconds)*time.Second))
}
```

---

**修复日期**: 2026-03-09  
**相关 Issue**: 定时任务受页面切换影响  
**影响范围**: 全局监控服务架构重构  
**向后兼容**: ✅ 完全兼容，无需修改现有配置

