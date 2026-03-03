# 域名状态显示说明

## 📋 状态监控面板字段说明

在 **状态监控** 页面的域名状态列表中，每个域名显示以下信息：

### 1️⃣ 当前 IP (`currentIP`)

**预期显示**: 
- ✅ **DDNS 更新后的公网 IP 地址**（例如：`123.45.67.89`）
- ⏳ **空值**（新添加的域名，还未执行 DDNS 更新）

**数据来源**: 
- 数据库字段：`domains.current_ip`
- 更新时机：当后台 DDNS 检查服务发现 IP 变化并成功更新 DNS 记录后

**示例**:
```
新添加的域名 → 当前 IP: (空)
              ↓
          执行 DDNS 更新
              ↓
已更新的域名 → 当前 IP: 123.45.67.89
```

---

### 2️⃣ 更新时间 (`lastUpdate`)

**预期显示**:
- ✅ **最后一次成功更新 DDNS 的时间**（例如：`2026-03-03 17:30:45`）
- ⏳ **空值**（从未更新过）

**数据来源**:
- 数据库字段：`domains.last_update`
- 更新时机：每次成功更新 DNS 记录后自动设置

**示例**:
```
新添加的域名 → 更新时间：(空)
              ↓
          执行 DDNS 更新
              ↓
已更新的域名 → 更新时间：2026-03-03 17:30:45
```

---

### 3️⃣ 状态指示器 (`status`)

**三种可能的状态**:

#### ✅ success（解析正常）
- **条件**: `currentIP` 不为空 **且** `lastUpdate` 不为空
- **显示**: 绿色圆点 + "解析正常"
- **含义**: 域名已成功更新到最新的公网 IP

#### ⏳ pending（等待更新）
- **条件**: `currentIP` 为空 **或** `lastUpdate` 为空
- **显示**: 橙色圆点 + "等待首次更新" 或 "已获取 IP，等待更新 DNS"
- **含义**: 域名配置存在，但还未执行 DDNS 更新

#### ❌ error（更新失败）
- **条件**: DDNS 更新过程中发生错误（未来实现）
- **显示**: 红色圆点 + 错误信息
- **含义**: 更新失败，需要检查配置或网络

---

## 🎯 实际效果对比

### 场景 1: 新添加的域名

```
┌─────────────────────────────────────┐
│ 🌐 test.example.com                 │
│    当前 IP: ⏳ (空，等待更新)        │
│                                     │
│    ⏳ 等待首次更新                  │
│    更新于：⏳ (空，未更新)           │
└─────────────────────────────────────┘
```

**解读**: 
- 域名刚添加，还未执行 DDNS 更新
- 系统会在后台自动检测并更新（如果实现了定时任务）
- 用户也可以手动触发刷新

---

### 场景 2: 已更新的域名

```
┌─────────────────────────────────────┐
│ 🌐 test.example.com                 │
│    当前 IP: ✅ 123.45.67.89          │
│                                     │
│    ✅ 解析正常                      │
│    更新于：✅ 2026-03-03 17:30:45   │
└─────────────────────────────────────┘
```

**解读**:
- DDNS 已成功执行
- 域名解析到正确的公网 IP
- 显示了最后的更新时间

---

## 🔧 后端逻辑

### GetStatus() 方法的状态判断

```go
// 查询数据库
querySQL := `
SELECT domain, COALESCE(current_ip, ''), COALESCE(last_update, '')
FROM domains
WHERE enabled = 1
ORDER BY domain
`

// 状态判断逻辑
if currentIP == "" {
    status = "pending"
    message = "等待首次更新"
} else if lastUpdate == "" {
    status = "pending"
    message = "已获取 IP，等待更新 DNS"
} else {
    status = "success"
    message = "解析正常"
}
```

---

## 📊 测试验证

### 运行测试脚本

```bash
go run scripts/test_domain_status.go
```

### 预期输出

```
🔍 测试域名状态显示逻辑...
============================================================

📊 启用域名的状态分析:
-----------------------------------------------------------

🌐 域名：test.example.com
   当前 IP: ⏳ (空，等待更新)
   最后更新：⏳ (空，未更新)
   状态：⏳ - 等待首次更新

============================================================
📈 统计信息:
   总域名数：1
   ✅ 解析正常：0
   ⏳ 等待更新：1

💡 提示：部分域名等待 DDNS 更新
   - 新添加的域名会显示「等待首次更新」
   - 系统会在后台自动执行 DDNS 更新
   - 更新后这里会显示实际的 IP 地址和更新时间
```

---

## 🚀 完整工作流程

### 1. 用户添加域名

```
前端提交表单
    ↓
后端保存到数据库
    ↓
current_ip = NULL
last_update = NULL
    ↓
状态监控显示：⏳ 等待首次更新
```

### 2. 后台 DDNS 检查（未来实现）

```
定时任务触发（例如每 5 分钟）
    ↓
获取当前公网 IP
    ↓
对比数据库中的 current_ip
    ↓
如果不同 → 调用 DNS 提供商 API 更新
    ↓
更新成功后：
    - current_ip = 新 IP
    - last_update = 当前时间
    ↓
状态监控显示：✅ 解析正常
```

### 3. 用户手动刷新（可选）

```
用户点击「刷新」按钮
    ↓
立即执行 DDNS 检查
    ↓
更新数据库
    ↓
界面立即显示最新状态
```

---

## ❓ 常见问题

### Q1: 为什么我刚添加的域名显示「等待首次更新」？

**A**: 这是正常的！新添加的域名还没有执行 DDNS 更新，所以：
- `currentIP` 为空
- `lastUpdate` 为空
- 状态显示为「等待首次更新」

当后台 DDNS 服务执行第一次更新后，就会显示实际的 IP 地址和更新时间。

---

### Q2: 如何让域名从「等待首次更新」变成「解析正常」？

**A**: 需要实现后台 DDNS 定时检查服务：

1. **定时任务**: 每隔一定时间（如 5 分钟）检查所有启用的域名
2. **获取公网 IP**: 调用第三方 API 获取当前公网 IP
3. **对比 IP**: 如果与数据库中的 `currentIP` 不同，则更新 DNS
4. **更新数据库**: 成功后更新 `currentIP` 和 `lastUpdate`

**临时方案**: 可以手动触发刷新（如果提供了此功能）。

---

### Q3: 状态会不会一直卡在「等待首次更新」？

**A**: 不会。一旦实现了后台 DDNS 定时检查服务，系统会自动：
- 定期获取公网 IP
- 更新 DNS 记录
- 刷新数据库中的状态

如果没有实现定时任务，则需要手动触发更新。

---

### Q4: 如果 DNS 更新失败了怎么办？

**A**: （未来实现）会增加错误状态：
- 状态显示：❌ 更新失败
- 错误信息：具体的失败原因（如 Token 无效、网络错误等）
- 重试机制：按照配置的重试次数和延迟自动重试

---

## 📝 数据库表结构

### domains 表相关字段

```sql
CREATE TABLE domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain TEXT NOT NULL,
    provider TEXT NOT NULL,
    token TEXT NOT NULL,
    interval INTEGER DEFAULT 300,
    enabled BOOLEAN DEFAULT 1,
    
    -- 这两个字段由 DDNS 更新服务维护
    current_ip TEXT,           -- 最后一次成功更新时的公网 IP
    last_update TIMESTAMP,     -- 最后一次成功更新的时间
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**字段说明**:
- `current_ip`: 记录最后一次成功更新到 DNS 服务商的公网 IP
- `last_update`: 记录最后一次成功更新的时间

**注意**: 这两个字段**不是**在前端表单中填写的，而是由后台 DDNS 服务自动更新的。

---

## 🎨 前端显示优化建议

### 当前实现

```vue
<div class="domain-ip">
  <span class="label">当前 IP:</span>
  <span class="ip">{{ domain.currentIP }}</span>
</div>
<div class="update-time">
  <small>更新于：{{ domain.lastUpdate }}</small>
</div>
```

### 优化建议（未来）

可以添加更直观的显示：

```vue
<div class="domain-ip">
  <span class="label">当前 IP:</span>
  <span v-if="domain.currentIP" class="ip">{{ domain.currentIP }}</span>
  <span v-else class="ip placeholder">等待更新...</span>
</div>
<div class="update-time">
  <small v-if="domain.lastUpdate">
    更新于：{{ domain.lastUpdate }}
  </small>
  <small v-else class="placeholder">
    尚未更新
  </small>
</div>
```

---

## ✅ 总结

### 预期显示

| 场景 | 当前 IP | 更新时间 | 状态 | 消息 |
|------|---------|----------|------|------|
| 新添加域名 | (空) | (空) | ⏳ pending | 等待首次更新 |
| 已更新域名 | 123.45.67.89 | 2026-03-03 17:30:45 | ✅ success | 解析正常 |
| IP 变化但未更新 | 旧 IP | 旧时间 | ⏳ pending | 等待更新 DNS |
| 更新失败 | 旧 IP | 旧时间 | ❌ error | 更新失败：xxx |

### 核心要点

1. **currentIP 和 lastUpdate 是由后台 DDNS 服务自动更新的**，不是用户手动填写的
2. **新域名显示「等待首次更新」是正常的**，表示还未执行 DDNS 更新
3. **状态判断逻辑已在后端修复**，现在会根据实际数据动态返回正确的状态
4. **需要实现后台定时检查服务**，才能让域名自动从「等待」变为「成功」

---

**修复日期**: 2026-03-03  
**相关文档**: [`scripts/test_domain_status.go`](../scripts/test_domain_status.go)  
**后端实现**: [`internal/domain/repository.go`](../../internal/domain/repository.go#L214-L244)
