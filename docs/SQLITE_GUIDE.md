# ipup - SQLite 集成指南

## 📋 概述

本指南详细介绍如何在 ipup-go 项目中使用 SQLite 数据库。

## 🎯 核心功能

- ✅ 自动创建 `ipup.db` 数据库文件
- ✅ 创建三个核心表：domains、settings、logs
- ✅ 创建必要的索引优化查询性能
- ✅ 初始化默认系统设置

## ✅ 已完成的集成

### 1. 数据库初始化
- ✅ 自动创建 `ddns.db` 数据库文件
- ✅ 创建三个核心表：domains、settings、logs
- ✅ 创建必要的索引优化查询性能
- ✅ 初始化默认系统设置

### 2. 数据存储方案

#### 域名配置 (domains 表)
```go
domain := DomainConfig{
    Domain:   "example.com",
    Provider: "Cloudflare",
    Token:    "your_api_token",
    Interval: 300,  // 5 分钟
    Enabled:  true,
}
err := app.AddDomain(domain)
```

**特性：**
- 🔒 域名唯一性约束
- 📝 自动记录创建和更新时间
- 🔄 支持 IP 地址追踪
- ⚡ 启用/禁用快速切换

#### 系统设置 (settings 表)
```go
// 加载设置
settings := app.LoadSettings()

// 保存设置
settings.CheckInterval = 600
err := app.SaveSettings(settings)

// 重置为默认值
err := app.ResetSettings()
```

**默认设置项：**
- 自动启动：true
- 检查间隔：300 秒
- 重试次数：3 次
- 重试延迟：10 秒
- 日志级别：info
- 通知策略：仅失败时通知

#### 日志记录 (logs 表)
```go
// 查询日志（支持级别筛选和关键词搜索）
logs := app.GetLogs("error", "example.com")

// 导出日志
content := app.ExportLogs()

// 清空日志
err := app.ClearLogs()
```

**日志级别：**
- ℹ️ info - 普通信息
- ⚠️ warning - 警告
- ❌ error - 错误
- ✅ success - 成功

### 3. 线程安全保证

所有数据库操作都使用了读写锁保护：

```go
// 读操作 - 允许多个并发读取
func (a *App) ListDomains() []DomainConfig {
    a.mu.RLock()
    defer a.mu.RUnlock()
    // ...
}

// 写操作 - 独占锁，保证写入安全
func (a *App) AddDomain(domain DomainConfig) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    // ...
}
```

### 4. 连接池优化

```go
db.SetMaxOpenConns(1)  // SQLite 不支持高并发写入
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(time.Hour)
```

## 📋 使用方法

### 前端调用示例

#### 添加域名
``typescript
import { AddDomain } from '../wailsjs/go/main/App'

const domain = {
  domain: 'example.com',
  provider: 'Cloudflare',
  token: 'sk_xxxxxxxxxxxx',
  interval: 300,
  enabled: true
}

try {
  await AddDomain(domain)
  console.log('添加成功')
} catch (error) {
  console.error('添加失败:', error)
}
```

#### 获取域名列表
```typescript
import { ListDomains } from '../wailsjs/go/main/App'

const domains = await ListDomains()
console.log('域名列表:', domains)
```

#### 保存设置
```typescript
import { SaveSettings } from '../wailsjs/go/main/App'

const settings = {
  autoStart: true,
  checkInterval: 300,
  retryCount: 3,
  retryDelay: 10,
  logLevel: 'info',
  notifySuccess: false,
  notifyError: true,
  proxy: '',
  apiEndpoint: 'https://api.ipify.org'
}

await SaveSettings(settings)
```

#### 查询日志
```typescript
import { GetLogs } from '../wailsjs/go/main/App'

// 查询所有日志
const allLogs = await GetLogs('all', '')

// 只查询错误日志
const errorLogs = await GetLogs('error', '')

// 搜索包含特定关键词的日志
const searchLogs = await GetLogs('all', 'example.com')
```

## 🔧 数据库管理

### 查看数据库文件

使用 SQLite 命令行工具：

```
# 打开数据库
sqlite3 ipup.db

# 查看所有表
.tables

# 查看域名配置
SELECT * FROM domains;

# 查看系统设置
SELECT * FROM settings;

# 查看最近 10 条日志
SELECT * FROM logs ORDER BY timestamp DESC LIMIT 10;

# 统计日志数量
SELECT level, COUNT(*) as count FROM logs GROUP BY level;

# 删除所有日志
DELETE FROM logs;

# 备份数据库
.backup ddns_backup.db
```

### 数据库文件位置

- **开发环境**: `./ipup.db`（相对于可执行文件）
- **生产环境**: 应用数据目录下的 `ddns.db`

### 备份策略

```
# Windows PowerShell
Copy-Item .\ddns.db .\ddns_backup_$(Get-Date -Format "yyyyMMdd_HHmmss").db

# Linux/Mac
cp ddns.db ddns_backup_$(date +%Y%m%d_%H%M%S).db
```

## 🧪 测试

运行测试验证所有数据库功能：

```
# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestDatabaseOperations

# 运行并发测试
go test -v -run TestConcurrentAccess
```

## 📊 数据库性能优化

### 已实现的优化

1. **索引优化**
   - `idx_logs_timestamp`: 加速按时间排序
   - `idx_logs_level`: 加速按级别筛选
   - `idx_domains_enabled`: 加速查询启用的域名

2. **查询限制**
   - 日志查询最多返回 1000 条
   - 避免全表扫描

3. **连接池**
   - 限制最大连接数为 1（SQLite 特性）
   - 保持最小连接活跃

### 建议的维护操作

```
-- 定期清理旧日志（保留最近 1000 条）
DELETE FROM logs 
WHERE id NOT IN (
  SELECT id FROM (
    SELECT id FROM logs ORDER BY timestamp DESC LIMIT 1000
  )
);

-- 更新统计信息
ANALYZE;

-- 优化数据库文件
VACUUM;
```

## ⚠️ 注意事项

### 数据安全

1. **Token 加密**: 目前 Token 以明文存储在生产数据库中
   - 建议：在生产环境中使用加密存储
   - 可以使用 AES 加密后存储

2. **SQL 注入防护**: 
   - ✅ 所有查询都使用参数化查询
   - ✅ 防止 SQL 注入攻击

3. **并发安全**:
   - ✅ 已实现读写锁
   - ✅ 避免数据库锁定问题

### 性能考虑

1. **批量操作**: 大量插入时使用事务
2. **定期清理**: 定期清理旧日志和数据
3. **备份策略**: 重要操作前自动备份

## 🔍 故障排查

### 常见问题

#### 1. 数据库锁定
```
错误：database is locked
解决：确保所有 rows.Close() 都被正确调用
```

#### 2. 无法写入
```
错误：attempt to write a readonly database
解决：检查文件权限，确保有写权限
```

#### 3. 连接失败
```
错误：unable to open database file
解决：检查路径是否正确，目录是否存在
```

### 调试技巧

在 [app.go](file://d:\go\wails\myproject\app.go) 中添加调试输出：

```go
func (a *App) addLog(level string, domain string, message string) {
    fmt.Printf("[DEBUG] 添加日志：%s [%s] %s\n", level, domain, message)
    // ...
}
```

## 📚 扩展阅读

- [SQLite 官方文档](https://www.sqlite.org/docs.html)
- [go-sqlite3 驱动文档](https://github.com/mattn/go-sqlite3)
- [Wails 数据绑定](https://wails.io/docs/guides/data-binding)

---

**集成完成日期**: 2026-03-03  
**数据库版本**: v1.0  
**最后更新**: 2026-03-03
