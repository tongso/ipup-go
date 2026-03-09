# 域名自动更新功能修复

## 🐛 问题描述

**现象**：域名设置了更新间隔（如 5 分钟），但程序并没有自动更新 DNS 解析记录。

**原因**：前端定时器只调用了 `GetDomainStatus()` 刷新状态显示，但没有实际触发 DNS 更新操作。

---

## 🔍 问题分析

### 修复前的逻辑

```typescript
// ❌ 错误的实现
const createRefreshTimer = (domainId: number, intervalSeconds: number) => {
  const timer = window.setInterval(() => {
    console.log(`⏰ 自动刷新域名 ID ${domainId} 的状态...`)
    // 只刷新状态显示，不实际更新 DNS
    refreshSingleDomainStatus(domainId)
  }, intervalMs)
}
```

**问题**：
1. 只调用 `GetDomainStatus()` 查询并显示状态
2. 不会调用 `UpdateDomainDNS()` 实际更新 DNS
3. 即使 IP 变化了，也不会触发阿里云 API 更新

### 数据流对比

#### 修复前（错误）

```
定时器触发
    ↓
调用 GetDomainStatus()
    ↓
仅更新 UI 显示
    ↓
❌ 不会实际更新 DNS 记录
```

#### 修复后（正确）

```
定时器触发
    ↓
检查域名是否启用
    ↓
调用 UpdateDomainDNS()
    ↓
1. 获取公网IP
2. 比较 IP 是否变化
3. 如果变化，调用阿里云 API 更新
4. 记录日志
    ↓
刷新 UI 状态显示
    ↓
✅ 完成 DNS 更新
```

---

## ✅ 修复方案

### 修改文件

**文件**: `frontend/src/components/DomainList.vue`

### 核心改动

#### 1. 导入 UpdateDomainDNS 方法

```typescript
// 修改前
import { ListDomains, AddDomain, UpdateDomain, DeleteDomain, ToggleDomain, GetDomainStatus } from '../../wailsjs/go/app/App'

// 修改后
import { ListDomains, AddDomain, UpdateDomain, DeleteDomain, ToggleDomain, GetDomainStatus, UpdateDomainDNS } from '../../wailsjs/go/app/App'
```

#### 2. 修改定时器回调函数

```typescript
const createRefreshTimer = (domainId: number, intervalSeconds: number) => {
  const intervalMs = intervalSeconds * 1000
  const timer = window.setInterval(async () => {
    console.log(`⏰ 自动刷新域名 ID ${domainId} 的状态...`)
    
    // 找到对应的域名配置
    const domain = domains.value.find(d => d.id === domainId)
    if (!domain || !domain.enabled) {
      console.log(`⚠️ 域名 ID ${domainId} 未启用或不存在，跳过更新`)
      return
    }
    
    try {
      // ✅ 调用后端 API 实际更新 DNS 解析
      console.log(`🔄 开始自动更新域名：${domain.domain}`)
      const result = await UpdateDomainDNS(domainId)
      
      // 显示成功通知
      notifySuccess(`${domain.domain}: ${result}`)
      
      // 更新本地状态
      await refreshSingleDomainStatus(domainId)
      
      // 触发自定义事件，通知其他组件
      window.dispatchEvent(new CustomEvent('domains-updated'))
      
    } catch (error) {
      const errorMsg = (error as Error).message
      
      console.error(`❌ 自动更新域名 ID ${domainId} 失败:`, errorMsg)
      
      // 智能错误处理
      if (errorMsg.includes('IP 未变化') || errorMsg.includes('跳过更新')) {
        notifyInfo(`${domain.domain}: ${errorMsg}`)
      } else {
        notifyError(`${domain.domain}: 自动更新失败 - ${errorMsg}`)
      }
    }
  }, intervalMs)
  
  refreshTimers.set(domainId, timer)
  console.log(`✅ 域名 ID ${domainId} 的刷新定时器已创建，间隔：${intervalSeconds}秒`)
}
```

---

## 🎯 关键改进点

### 1. 实际触发 DNS 更新

- ✅ 调用 `UpdateDomainDNS(domainId)` 方法
- ✅ 执行完整的更新流程：获取 IP → 比较 → 更新数据库 → 调用阿里云 API

### 2. 智能错误处理

| 错误类型 | 处理方式 | 用户体验 |
|---------|---------|---------|
| IP 未变化 | 显示信息提示（蓝色） | 友好，知道是正常情况 |
| 网络错误 | 显示错误通知（红色） | 清晰，便于排查 |
| API 权限错误 | 显示错误通知（红色） | 明确，指导修复 |

### 3. 状态同步优化

```typescript
// 更新成功后
await refreshSingleDomainStatus(domainId)  // 刷新本地状态
window.dispatchEvent(new CustomEvent('domains-updated'))  // 通知其他组件
```

**效果**：
- DomainList.vue：立即更新显示
- StatusPanel.vue：监听事件后立即刷新监控面板

---

## 🧪 测试验证步骤

### 场景 1: 不同刷新间隔的域名

**步骤**：
1. 添加域名 `ddns.jizhangxia.com`，设置 `interval = 300`（5 分钟）
2. 添加域名 `test.example.com`，设置 `interval = 60`（1 分钟）
3. 观察控制台日志和通知

**预期结果**：
```
✅ 域名 ID 1 的刷新定时器已创建，间隔：300 秒
✅ 域名 ID 2 的刷新定时器已创建，间隔：60 秒

⏰ 自动刷新域名 ID 2 的状态...
🔄 开始自动更新域名：test.example.com
[Aliyun] ✓ IP 无变化，跳过更新：test.example.com -> 116.230.251.7
ℹ️ test.example.com: IP 未变化，跳过更新

⏰ 自动刷新域名 ID 1 的状态...
🔄 开始自动更新域名：ddns.jizhangxia.com
[Aliyun] IP 已变化：116.230.251.7 -> 123.45.67.89，开始更新...
[Aliyun] 成功更新 DNS 记录：ddns.jizhangxia.com -> 123.45.67.89
✅ ddns.jizhangxia.com: 成功调用 Aliyun API 更新 DNS 记录
```

### 场景 2: 禁用域名

**步骤**：
1. 禁用域名 `ddns.jizhangxia.com`
2. 等待 5 分钟

**预期结果**：
```
⚠️ 域名 ID 1 未启用或不存在，跳过更新
```

**验证**：禁用的域名不会触发自动更新

### 场景 3: 修改刷新间隔

**步骤**：
1. 编辑域名，将 `interval` 从 300 改为 60
2. 保存后观察定时器是否重建

**预期结果**：
```
🗑️ 已清除域名 ID 1 的刷新定时器
✅ 域名 ID 1 的刷新定时器已创建，间隔：60 秒
```

**验证**：修改间隔后，旧定时器被清除，新定时器按新间隔创建

### 场景 4: IP 变化时的完整流程

**步骤**：
1. 手动修改本地公网IP（模拟变化）
2. 等待定时器触发

**预期结果**：
```
🔄 开始自动更新域名：ddns.jizhangxia.com
[Aliyun] IP 已变化：116.230.251.7 -> 新 IP，开始更新...
[Aliyun] ✓ 找到匹配的 A 记录：ddns, RecordID: xxx, Value: 旧 IP
[Aliyun] 成功更新 DNS 记录：ddns.jizhangxia.com -> 新 IP
✅ ddns.jizhangxia.com: 成功调用 Aliyun API 更新 DNS 记录
```

---

## 📊 定时器管理细节

### 创建时机

| 操作 | 是否重建定时器 | 说明 |
|------|--------------|------|
| 页面加载 | ✅ 是 | `onMounted()` → `loadDomains()` |
| 添加域名 | ✅ 是 | `addDomain()` → `loadDomains()` |
| 编辑保存 | ✅ 是 | `saveEdit()` → `loadDomains()` |
| 删除域名 | ✅ 是 | `deleteDomain()` → `loadDomains()` |
| 启用/禁用 | ✅ 是 | `toggleEnabled()` → `loadDomains()` |

### 销毁时机

| 操作 | 是否清除定时器 | 说明 |
|------|--------------|------|
| 页面卸载 | ✅ 是 | `onUnmounted()` 清除所有 |
| 删除域名 | ✅ 是 | `updateAllRefreshTimers()` 重建时清除 |
| 修改间隔 | ✅ 是 | `createRefreshTimer()` 中先清除旧的 |

### 定时器生命周期

```
组件挂载
    ↓
loadDomains()
    ↓
updateAllRefreshTimers()
    ↓
为每个启用的域名创建定时器
    ↓
定时触发 UpdateDomainDNS()
    ↓
组件卸载
    ↓
清除所有定时器
```

---

## 💡 优势分析

### 修复前的问题

- ❌ **只看不做**：定时器只刷新 UI，不实际更新 DNS
- ❌ **形同虚设**：设置了 interval 但不起作用
- ❌ **需要手动**：必须点击"手动更新"按钮才会更新

### 修复后的优势

- ✅ **名副其实**：定时器真正触发 DNS 更新
- ✅ **自动化**：按照设定的间隔自动更新
- ✅ **智能容错**：IP 不变时显示提示，不报错
- ✅ **即时反馈**：通过通知系统实时反馈状态

---

## 🎨 用户体验提升

### 通知系统

| 场景 | 通知类型 | 颜色 | 文案示例 |
|------|---------|------|---------|
| 更新成功 | Success | 绿色 | ✅ ddns.jizhangxia.com: 成功调用 Aliyun API 更新 DNS 记录 |
| IP 未变化 | Info | 蓝色 | ℹ️ ddns.jizhangxia.com: IP 未变化，跳过更新 |
| 网络错误 | Error | 红色 | ❌ ddns.jizhangxia.com: 自动更新失败 - 获取公网IP 失败 |
| API 错误 | Error | 红色 | ❌ ddns.jizhangxia.com: 自动更新失败 - InvalidAccessKeyID.NotFound |

### 控制台日志

```
✅ 域名 ID 1 的刷新定时器已创建，间隔：300 秒
⏰ 自动刷新域名 ID 1 的状态...
🔄 开始自动更新域名：ddns.jizhangxia.com
[Aliyun] ✓ IP 无变化，跳过更新：ddns.jizhangxia.com -> 116.230.251.7
ℹ️ ddns.jizhangxia.com: IP 未变化，跳过更新
```

---

## 📝 相关文件

### 修改的文件

1. ✅ [`frontend/src/components/DomainList.vue`](../frontend/src/components/DomainList.vue)
   - 导入 `UpdateDomainDNS` 方法
   - 修改 `createRefreshTimer()` 回调函数
   - 添加智能错误处理

### 依赖的后端方法

1. ✅ [`internal/app/handlers_domain.go::UpdateDomainDNS`](../internal/app/handlers_domain.go#L148-L204)
   - 获取公网IP
   - 更新数据库 IP
   - 调用 DNS 提供商 API
   - 记录详细日志

---

## 🚀 后续优化建议

### 1. 批量更新优化

当多个域名的更新时间接近时，合并为一次 IP 获取：

```typescript
const batchUpdateQueue = new Map<number, number>()

const scheduleBatchUpdate = (domainId: number) => {
  batchUpdateQueue.set(domainId, Date.now())
  
  // 防抖：等待 1 秒，收集更多待更新的域名
  setTimeout(async () => {
    if (batchUpdateQueue.size === 0) return
    
    // 一次性获取 IP，然后批量更新
    const domainIds = Array.from(batchUpdateQueue.keys())
    await updateBatchDomains(domainIds)
    
    batchUpdateQueue.clear()
  }, 1000)
}
```

### 2. 失败重试机制

当更新失败时，自动重试：

```typescript
let failCount = 0
const MAX_RETRY = 3
const RETRY_DELAY = 5000 // 5 秒

if (error) {
  failCount++
  if (failCount <= MAX_RETRY) {
    setTimeout(() => {
      UpdateDomainDNS(domainId)
    }, RETRY_DELAY)
  }
}
```

### 3. 可视化下次刷新时间

在 UI 上显示倒计时：

```vue
<div class="next-refresh">
  下次刷新：{{ formatCountdown(domain.interval, domain.lastUpdate) }}
</div>
```

---

**修复日期**: 2026-03-09  
**相关 Issue**: 域名设置了更新间隔但不能自动更新  
**影响范围**: 所有启用的域名自动更新功能
