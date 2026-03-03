# ipup-go - 数据库验证快速指南

## 🚀 快速验证步骤

### 方法一：运行验证脚本（推荐）

```bash
# 编译并运行验证脚本
go run verify_db.go
```

这个脚本会自动检查:
- ✅ 数据库文件是否存在
- ✅ settings 表中的所有设置
- ✅ domains 表中的域名配置
- ✅ logs 表中的最近日志
- ✅ 数据统计信息

### 方法二：使用 SQLite 命令行

```bash
# 打开数据库
sqlite3 ipup.db

# 查看所有设置
SELECT * FROM settings;

# 查看启用的域名
SELECT domain, current_ip, last_update FROM domains WHERE enabled = 1;

# 查看最近 10 条日志
SELECT timestamp, level, domain, message FROM logs ORDER BY timestamp DESC LIMIT 10;

# 退出
.quit
```

## 🔍 问题排查

### 如果 settings 表为空

**说明设置没有保存到数据库**

可能原因:
1. ❌ 前端没有调用 `SaveSettings` API
2. ❌ 后端 `SaveSettings` 方法执行失败
3. ❌ 数据库连接问题

**调试步骤**:
```bash
# 1. 启动应用时查看控制台输出
wails dev

# 2. 修改设置并保存

# 3. 查看是否有错误日志

# 4. 立即检查数据库
go run verify_db.go
```

### 如果域名状态不显示

**说明 GetDomainStatus 没有读到数据**

可能原因:
1. ❌ domains 表中没有数据
2. ❌ 域名的 enabled = 0 (未启用)
3. ❌ 前端没有调用 `GetDomainStatus` API

**调试步骤**:
```bash
# 1. 检查域名是否启用
sqlite3 ipup.db "SELECT domain, enabled FROM domains;"

# 2. 如果没有域名，先添加一个
# 在界面上进入"域名管理" → 添加域名 → 确保勾选"启用"

# 3. 再次检查
go run verify_db.go
```

## ✅ 正常情况下的输出

运行 `go run verify_db.go` 应该看到类似输出:

```
✅ 数据库文件存在：./ipup.db
============================================================

📋 检查 settings 表:
  autoStart            | true       | 2026-03-03 14:05:23
  apiEndpoint          | https://...| 2026-03-03 14:05:23
  checkInterval        | 300        | 2026-03-03 14:05:23
  logLevel             | info       | 2026-03-03 14:05:23
  notifyError          | true       | 2026-03-03 14:05:23
  notifySuccess        | false      | 2026-03-03 14:05:23
  proxy                |            | 2026-03-03 14:05:23
  retryCount           | 3          | 2026-03-03 14:05:23
  retryDelay           | 10         | 2026-03-03 14:05:23
  ✅ 共 9 条设置记录

🌐 检查 domains 表:
  [✅] example.com | Cloudflare | IP: 192.168.1.1
  [❌] test.com    | Aliyun     | IP: 
  ✅ 共 2 个域名 (1 个启用)

📝 检查 logs 表 (最近 10 条):
  2026-03-03 14:05:23 | ℹ️ | : 系统设置已保存
  2026-03-03 14:04:15 | ✅ | example.com: IP 地址更新成功
  2026-03-03 14:03:08 | ℹ️ | : 添加了域名 example.com
  ✅ 最近 3 条日志

============================================================
📊 统计信息:
  Settings: 9 条
  Domains:  2 个
  Logs:     3 条

✅ 数据库检查完成!
```

## 🛠️ 如果数据不存在

### 初始化数据库

如果你刚启动应用，数据库可能是空的。请按以下步骤操作:

1. **添加域名配置**:
   - 进入"域名管理"标签页
   - 点击"添加域名"
   - 填写信息并保存

2. **修改系统设置**:
   - 进入"设置"标签页
   - 修改任意设置
   - 点击"保存设置"

3. **产生日志**:
   - 上述操作会自动产生日志
   - 可以在"日志查看"中查看

4. **再次验证**:
   ```bash
   go run verify_db.go
   ```

## 📝 代码已经正确实现

### ✅ SaveSettings 方法
位置：[app.go](file://d:\go\wails\myproject\app.go#L501-L545)
- 将设置转换为 key-value 对
- 使用 `INSERT OR REPLACE` 保存到数据库
- 记录操作日志

### ✅ LoadSettings 方法
位置：[app.go](file://d:\go\wails\myproject\app.go#L548-L602)
- 从 database 查询所有设置
- 将字符串转换回对应类型
- 返回完整的 Settings 对象

### ✅ GetDomainStatus 方法
位置：[app.go](file://d:\go\wails\myproject\app.go#L346-L377)
- 查询所有 `enabled = 1` 的域名
- 返回域名状态列表
- 包含 IP 地址和最后更新时间

### ✅ 前端集成
- [Settings.vue](file://d:\go\wails\myproject\frontend\src\components\Settings.vue) - 已调用 `LoadSettings` 和 `SaveSettings`
- [StatusPanel.vue](file://d:\go\wails\myproject\frontend\src\components\StatusPanel.vue) - 已调用 `GetDomainStatus`

---

**建议**: 先运行 `go run verify_db.go` 查看数据库实际情况，再根据输出判断问题所在。
