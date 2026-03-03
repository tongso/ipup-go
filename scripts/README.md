# 开发脚本目录

本目录包含项目开发过程中使用的各种辅助脚本和测试工具。

## 📁 脚本说明

### 🔧 db_migrate.go
**数据库迁移工具**
- 用于执行数据库结构变更和数据迁移
- 支持数据库备份和恢复功能
- **运行方式**: `go run scripts/db_migrate.go`

### ✅ verify_db.go
**数据库验证工具**
- 检查 SQLite 数据库文件完整性
- 验证表结构和数据记录
- 显示 settings、domains、logs 表的详细状态
- **运行方式**: `go run scripts/verify_db.go`

### 🧪 app_test.go
**应用单元测试**
- 包含核心功能的自动化测试用例
- 用于回归测试和功能验证
- **运行方式**: `go test -v ./scripts/app_test.go`

### 🌐 test_ip.go
**IP 获取功能测试**
- 测试 IPv4/IPv6 双栈网络支持
- 验证多个公共 IP 查询 API
- **运行方式**: `go run scripts/test_ip.go`

## 🚀 使用示例

```bash
# 验证数据库状态
go run scripts/verify_db.go

# 测试 IP 获取功能
go run scripts/test_ip.go

# 运行单元测试
go test -v ./scripts/*.go

# 备份数据库
go run scripts/db_migrate.go
```

## ⚠️ 注意事项

1. 所有脚本均使用 `//go:build ignore` 构建标签，不会参与主应用的编译
2. 运行脚本前请确保已安装必要的 Go 依赖（如 `modernc.org/sqlite`）
3. 数据库操作脚本默认在项目根目录运行，以确保正确的相对路径
4. 建议定期使用 `verify_db.go` 检查数据库健康状态

## 📝 添加新脚本

当需要添加新的辅助脚本时：
1. 将脚本放在此目录
2. 在脚本开头添加 `//go:build ignore`
3. 更新此 README 说明脚本用途
4. 确保脚本有清晰的注释和使用说明
