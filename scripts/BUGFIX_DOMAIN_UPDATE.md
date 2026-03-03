# 域名更新 BUG 修复验证指南

## 🐛 问题描述

**BUG**: 域名修改后不能立即更新在界面上，包括域名列表和监控面板上

## 🔍 问题原因

### 后端问题
在 `internal/domain/repository.go` 的 `Update` 方法中：
```go
// ❌ 错误的代码（未更新 domain 字段）
UPDATE domains 
SET provider = ?, token = ?, interval = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?
```

**缺少了 `domain` 字段的更新**，导致即使前端传递了新的域名，数据库中的域名也不会改变。

### 前端问题
1. **DomainList.vue**: 虽然调用了 `loadDomains()`，但没有通知其他组件
2. **StatusPanel.vue**: 只依赖定时刷新（5 分钟），不会立即响应域名变化

## ✅ 修复方案

### 1. 后端修复

**文件**: `internal/domain/repository.go`

```go
// ✅ 正确的代码（包含所有字段更新）
func (r *Repository) Update(domain types.Domain) error {
	updateSQL := `
	UPDATE domains 
	SET domain = ?, provider = ?, token = ?, interval = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP 
	WHERE id = ?
	`
	
	_, err := r.db.Exec(updateSQL, domain.Domain, domain.Provider, domain.Token, domain.Interval, domain.Enabled, domain.ID)
	if err != nil {
		return fmt.Errorf("更新域名失败：%w", err)
	}
	
	return nil
}
```

**关键变更**: 添加 `domain = ?` 到 SET 子句中

---

### 2. 前端优化

#### DomainList.vue - 添加事件通知

**修改位置**: `saveEdit()`, `addDomain()`, `deleteDomain()` 方法

```typescript
// 保存编辑后
await loadDomains() // 重新加载列表
window.dispatchEvent(new CustomEvent('domains-updated')) // 通知其他组件
```

**作用**: 当域名发生变化时，通过自定义事件通知所有监听的组件

---

#### StatusPanel.vue - 监听更新事件

**修改位置**: `onMounted()` 生命周期

```typescript
onMounted(() => {
  refreshStatus() // 初始加载
  
  // 每 5 分钟自动刷新
  const interval = setInterval(refreshStatus, 5 * 60 * 1000)
  
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

**作用**: 当收到 `domains-updated` 事件时，立即刷新状态面板

---

## 🧪 验证步骤

### 方法 1: 使用测试脚本验证

```bash
# 运行域名更新测试
go run scripts/test_domain_update.go
```

**预期输出**:
```
🔍 测试域名更新功能...
============================================================

📋 当前域名列表:
  [3] ✅ | test1.xx.com | Cloudflare

🧪 测试更新域名 ID: 3 (test1.xx.com)
✅ 更新成功，影响行数：1

📋 验证更新后的域名:
  域名 ID 3 现在是：updated-1772528856.test1.xx.com

✅ 域名更新测试通过！
```

---

### 方法 2: 手动界面测试

#### 步骤 1: 启动应用
```bash
wails dev
```

#### 步骤 2: 测试域名列表更新
1. 打开 **域名管理** 页面
2. 点击任意域名的 **✏️ 编辑** 按钮
3. 修改域名（例如：从 `example.com` 改为 `new-example.com`）
4. 点击 **保存修改**
5. ✅ **验证**: 域名列表应立即显示新域名 `new-example.com`

#### 步骤 3: 测试监控面板更新
1. 在 **域名管理** 页面修改一个域名
2. 切换到 **状态监控** 页面
3. ✅ **验证**: 域名状态列表中应立即显示修改后的域名

#### 步骤 4: 测试添加域名
1. 在 **域名管理** 页面添加新域名
2. 切换到 **状态监控** 页面
3. ✅ **验证**: 应立即显示新域名的状态

#### 步骤 5: 测试删除域名
1. 在 **域名管理** 页面删除一个域名
2. 切换到 **状态监控** 页面
3. ✅ **验证**: 已删除的域名应从状态列表中消失

---

## 📊 测试检查清单

- [ ] **后端测试**
  - [ ] 运行 `go run scripts/test_domain_update.go` 通过
  - [ ] 数据库中 domain 字段正确更新
  - [ ] updated_at 时间戳正确更新

- [ ] **前端测试 - 域名列表**
  - [ ] 修改域名后立即显示新名称
  - [ ] 添加域名后立即出现在列表中
  - [ ] 删除域名后立即从列表中移除

- [ ] **前端测试 - 状态监控**
  - [ ] 修改域名后状态面板立即更新
  - [ ] 添加域名后状态面板立即显示
  - [ ] 删除域名后状态面板立即移除

- [ ] **跨页面测试**
  - [ ] 在域名管理页修改 → 切换到状态监控页应立即更新
  - [ ] 在状态监控页刷新 → 应显示最新的域名数据

---

## 🎯 验收标准

### ✅ 必须满足
1. 修改域名后，域名列表**立即**显示新域名（无需手动刷新）
2. 修改域名后，状态监控面板**立即**显示新域名（无需等待 5 分钟）
3. 添加/删除域名后，两个页面都立即同步

### ✅ 性能要求
- 界面响应时间 < 200ms
- 数据库查询时间 < 100ms
- 事件传播延迟 < 50ms

---

## 🔧 故障排查

### 如果域名列表不更新

1. **检查浏览器控制台**: 查看是否有 JavaScript 错误
2. **检查网络请求**: 确认 ListDomains API 调用成功
3. **检查数据库**: 运行验证脚本确认数据已更新

### 如果状态监控不更新

1. **检查事件监听**: 在浏览器控制台输入以下代码测试事件：
   ```javascript
   window.dispatchEvent(new CustomEvent('domains-updated'))
   ```
2. **检查 API 调用**: 确认 GetDomainStatus 返回最新数据
3. **检查定时器**: 确认没有多个定时器冲突

---

## 📝 技术细节

### 事件机制说明

```
用户操作 → 调用后端 API → 更新数据库 → 
↓
前端 loadDomains() → 更新本地状态 → dispatchEvent('domains-updated')
↓
StatusPanel 监听到事件 → 调用 refreshStatus() → 获取最新数据 → 更新 UI
```

### 数据流

```
Vue Component (Form) 
  ↓
Backend API (UpdateDomain)
  ↓
Repository (Update SQL)
  ↓
SQLite Database
  ↓
Repository (List SQL with COALESCE for NULL handling)
  ↓
Backend API (ListDomains)
  ↓
Vue Component (Reactive Update)
  ↓
Custom Event (domains-updated)
  ↓
Other Components Listen & Refresh
```

---

## 🚀 后续优化建议

1. **使用状态管理**: 考虑引入 Pinia/Vuex 统一管理应用状态
2. **乐观更新**: 在 UI 上立即显示更改，后台异步同步数据库
3. **WebSocket 实时推送**: 多标签页场景下使用 WebSocket 同步状态
4. **防抖处理**: 对频繁的操作进行防抖，减少请求次数

---

## 📅 修复记录

- **修复日期**: 2026-03-03
- **修复内容**: 
  - 后端：添加 domain 字段到 Update SQL
  - 前端：添加事件通知机制实现跨组件同步
- **测试脚本**: `scripts/test_domain_update.go`
- **相关文档**: `docs/BUGFIX_DOMAIN_UPDATE.md`

---

**✅ 修复完成，请按照上述步骤进行验证！**
