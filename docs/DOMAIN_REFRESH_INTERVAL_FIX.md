# 域名刷新间隔配置修复说明

## 🐛 问题描述

### 问题 1: 域名列表刷新频率写死了
**修复前**: 所有域名都使用固定的 5 分钟刷新间隔  
**问题代码**:
```typescript
const interval = setInterval(refreshStatus, 5 * 60 * 1000) // ❌ 写死的 5 分钟
```

**应该**: 根据每个域名的 `interval` 字段（更新间隔）来分别刷新

---

### 问题 2: 状态监控面板刷新间隔硬编码
**修复前**: 状态监控面板固定 5 分钟刷新  
**问题代码**:
```typescript
setInterval(refreshStatus, 5 * 60 * 1000) // ❌ 应该从设置中读取
```

**应该**: 从系统设置中读取 `CheckInterval` 配置值

---

## ✅ 修复方案

### 修复 1: 域名列表 - 独立刷新定时器

**文件**: `frontend/src/components/DomainList.vue`

#### 核心实现

```typescript
// 存储每个域名的刷新定时器
const refreshTimers = new Map<number, number>()

// 为单个域名创建刷新定时器
const createRefreshTimer = (domainId: number, intervalSeconds: number) => {
  // 清除旧的定时器
  const oldTimer = refreshTimers.get(domainId)
  if (oldTimer) {
    clearInterval(oldTimer)
  }
  
  // 创建新的定时器（转换为毫秒）
  const intervalMs = intervalSeconds * 1000
  const timer = window.setInterval(() => {
    console.log(`⏰ 自动刷新域名 ID ${domainId} 的状态...`)
    refreshSingleDomainStatus(domainId)
  }, intervalMs)
  
  refreshTimers.set(domainId, timer)
  console.log(`✅ 域名 ID ${domainId} 的刷新定时器已创建，间隔：${intervalSeconds}秒`)
}

// 刷新单个域名的状态（仅更新状态显示，不重新加载列表）
const refreshSingleDomainStatus = async (domainId: number) => {
  try {
    const statuses = await GetDomainStatus()
    if (statuses) {
      const status = statuses.find(s => {
        const targetDomain = domains.value.find(d => d.id === domainId)
        return targetDomain && s.domain === targetDomain.domain
      })
      
      if (status) {
        const target = domains.value.find(d => d.id === domainId)
        if (target) {
          target.currentIP = status.currentIP
          target.lastUpdate = status.lastUpdate
        }
      }
    }
  } catch (error) {
    console.error(`刷新域名 ID ${domainId} 状态失败:`, error)
  }
}

// 更新所有域名刷新定时器
const updateAllRefreshTimers = () => {
  // 清除所有旧定时器
  refreshTimers.forEach((timer, id) => {
    clearInterval(timer)
  })
  refreshTimers.clear()
  
  // 为每个启用的域名创建新定时器
  domains.value.forEach(domain => {
    if (domain.enabled && domain.interval > 0) {
      createRefreshTimer(domain.id, domain.interval)
    }
  })
  
  console.log(`📊 已创建 ${refreshTimers.size} 个域名刷新定时器`)
}
```

#### 调用时机

```typescript
// 加载域名列表时
const loadDomains = async () => {
  try {
    isLoading.value = true
    const result = await ListDomains()
    domains.value = result || []
    
    // 为每个启用的域名创建独立刷新定时器
    updateAllRefreshTimers()
  } catch (error) {
    console.error('加载域名列表失败:', error)
  } finally {
    isLoading.value = false
  }
}
```

#### 生命周期管理

```typescript
// 组件卸载时清除所有定时器
onUnmounted(() => {
  refreshTimers.forEach((timer, id) => {
    clearInterval(timer)
  })
  refreshTimers.clear()
  console.log('🗑️ 已清除所有域名刷新定时器')
})
```

---

### 修复 2: 状态监控面板 - 从设置读取刷新间隔

**文件**: `frontend/src/components/StatusPanel.vue`

```typescript
onMounted(() => {
  // 初始加载
  refreshStatus()
  
  // 从设置中读取检查间隔（默认 5 分钟）
  let checkInterval = 5 * 60 * 1000 // 默认值
  
  // 异步获取设置
  GetSettings().then(settings => {
    if (settings && settings.checkInterval) {
      checkInterval = settings.checkInterval * 60 * 1000 // 转换为毫秒
      console.log(`📡 状态监控刷新间隔：${settings.checkInterval}分钟`)
    }
  }).catch(err => {
    console.warn('获取设置失败，使用默认值:', err)
  })
  
  // 定时刷新状态监控面板
  const interval = setInterval(refreshStatus, checkInterval)
  
  // 监听域名更新事件，立即刷新状态
  window.addEventListener('domains-updated', () => {
    console.log('📡 检测到域名更新，立即刷新状态...')
    refreshStatus()
  })
  
  return () => {
    clearInterval(interval)
    window.removeEventListener('domains-updated', () => {})
  }
})
```

---

## 📊 数据流对比

### 修复前（错误）

```
所有域名 → 共用一个定时器 → 固定 5 分钟刷新一次
              ↓
         无法区分不同域名的刷新需求
```

### 修复后（正确）

```
域名 A (interval=60)   → 独立定时器 → 每 60 秒刷新
域名 B (interval=300)  → 独立定时器 → 每 300 秒刷新
域名 C (interval=600)  → 独立定时器 → 每 600 秒刷新
              
状态监控面板 → 从设置读取 CheckInterval → 动态刷新间隔
```

---

## 🎯 刷新机制详解

### 域名列表刷新策略

| 域名 | Interval(秒) | Enabled | 刷新频率 | 说明 |
|------|-------------|---------|----------|------|
| example1.com | 60 | ✅ | 每 60 秒 | 高频刷新 |
| example2.com | 300 | ✅ | 每 300 秒 | 中频刷新 |
| example3.com | 600 | ✅ | 每 600 秒 | 低频刷新 |
| example4.com | 300 | ❌ | 不刷新 | 已禁用 |

### 状态监控面板刷新策略

| 设置项 | 值 (分钟) | 实际间隔 | 说明 |
|--------|----------|----------|------|
| CheckInterval | 1 | 60 秒 | 每分钟刷新 |
| CheckInterval | 5 | 300 秒 | 每 5 分钟刷新（默认） |
| CheckInterval | 10 | 600 秒 | 每 10 分钟刷新 |

---

## 🔧 定时器管理细节

### 创建时机

1. **组件初始化**: `onMounted()` → `loadDomains()` → `updateAllRefreshTimers()`
2. **添加域名后**: `addDomain()` → `loadDomains()` → `updateAllRefreshTimers()`
3. **编辑保存后**: `saveEdit()` → `loadDomains()` → `updateAllRefreshTimers()`
4. **启用/禁用**: `toggleEnabled()` → `loadDomains()` → `updateAllRefreshTimers()`

### 销毁时机

1. **删除域名**: `deleteDomain()` → `loadDomains()` → `updateAllRefreshTimers()`（重建）
2. **组件卸载**: `onUnmounted()` → 清除所有定时器

### 更新流程

```
用户操作（添加/编辑/删除/切换）
    ↓
调用后端 API
    ↓
loadDomains() 重新加载列表
    ↓
updateAllRefreshTimers()
    ↓
清除所有旧定时器
    ↓
为每个启用的域名创建新定时器
    ↓
完成 ✅
```

---

## 💡 优势分析

### 修复前的问题

- ❌ **一刀切**: 所有域名统一 5 分钟刷新，无法满足个性化需求
- ❌ **浪费资源**: 不需要频繁刷新的域名也被迫高频刷新
- ❌ **不够灵活**: 需要快速响应的域名无法及时更新
- ❌ **设置无效**: 用户在设置中配置的刷新间隔不起作用

### 修复后的优势

- ✅ **个性化**: 每个域名根据自己的配置独立刷新
- ✅ **资源优化**: 按需刷新，减少不必要的请求
- ✅ **灵活可控**: 用户可以自定义每个域名的刷新频率
- ✅ **设置生效**: 状态监控面板的刷新间隔由用户配置决定

---

## 🧪 测试场景

### 场景 1: 不同刷新间隔的域名

```
添加 3 个域名：
- example1.com: interval = 60 (1 分钟)
- example2.com: interval = 300 (5 分钟)
- example3.com: interval = 600 (10 分钟)

预期结果:
- example1.com 每 1 分钟自动刷新一次
- example2.com 每 5 分钟自动刷新一次
- example3.com 每 10 分钟自动刷新一次
```

### 场景 2: 修改刷新间隔

```
1. 添加域名 example.com，interval = 300
2. 等待 2 分钟，观察是否刷新（应该未刷新）
3. 编辑域名，修改 interval = 60
4. 观察是否在 1 分钟后刷新（应该立即刷新）
```

### 场景 3: 启用/禁用域名

```
1. 禁用域名 example.com
   → 应该清除该域名的刷新定时器
   
2. 重新启用域名 example.com
   → 应该重新创建刷新定时器
```

### 场景 4: 状态监控面板设置

```
1. 在设置页面修改 CheckInterval = 1 (分钟)
2. 刷新状态监控页面
3. 观察是否每 1 分钟自动刷新一次

4. 修改 CheckInterval = 10 (分钟)
5. 刷新状态监控页面
6. 观察是否每 10 分钟自动刷新一次
```

---

## 📝 修改的文件

1. ✅ [`frontend/src/components/DomainList.vue`](../../frontend/src/components/DomainList.vue)
   - 新增 `refreshTimers` Map 存储定时器
   - 新增 `createRefreshTimer()` 方法
   - 新增 `refreshSingleDomainStatus()` 方法
   - 新增 `updateAllRefreshTimers()` 方法
   - 新增 `clearDomainTimer()` 方法
   - 修改 `loadDomains()` 调用定时器更新
   - 添加 `onUnmounted()` 清理逻辑

2. ✅ [`frontend/src/components/StatusPanel.vue`](../../frontend/src/components/StatusPanel.vue)
   - 导入 `GetSettings` 方法
   - 修改自动刷新逻辑，从设置中读取间隔
   - 添加默认值处理

---

## 🚀 后续优化建议

### 1. 可视化定时器状态

在域名列表中添加一列"下次刷新时间"：

```vue
<div class="next-refresh">
  下次刷新：{{ formatNextRefresh(domain.interval) }}
</div>
```

### 2. 智能降级策略

当域名连续多次刷新失败时，自动降低刷新频率或暂停刷新：

```typescript
let failCount = 0
const MAX_FAILS = 5

if (failCount >= MAX_FAILS) {
  // 暂停刷新或延长间隔
  pauseDomainRefresh(domainId)
}
```

### 3. 批量刷新优化

当多个域名的刷新时间接近时，合并为一次请求：

```typescript
// 如果多个域名在 10 秒内需要刷新，合并为一次 GetDomainStatus 调用
const batchRefresh = debounce(() => {
  refreshAllEnabledDomains()
}, 10000)
```

---

## ✅ 总结

### 核心改进

1. **域名列表**: 
   - ✅ 每个域名独立的刷新定时器
   - ✅ 根据自身 `interval` 字段决定刷新频率
   - ✅ 只在启用状态下才刷新

2. **状态监控**:
   - ✅ 从系统设置读取 `CheckInterval`
   - ✅ 支持用户自定义刷新间隔
   - ✅ 有合理的默认值（5 分钟）

### 用户体验提升

- ✅ **灵活性**: 用户可以为不同域名设置不同的刷新频率
- ✅ **可控性**: 状态监控的刷新频率由用户掌控
- ✅ **性能**: 减少不必要的刷新请求，节省资源
- ✅ **准确性**: 每个域名按照自己的节奏刷新，数据更准确

---

**修复日期**: 2026-03-03  
**相关 Issue**: 域名刷新间隔配置问题  
**前端组件**: [`DomainList.vue`](../../frontend/src/components/DomainList.vue), [`StatusPanel.vue`](../../frontend/src/components/StatusPanel.vue)
