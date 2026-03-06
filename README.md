# ipup-go - DDNS 动态域名解析系统

## 🎉 项目简介

ipup-go 是一个基于 Wails + Vue 的现代化 DDNS（动态域名解析）管理工具，使用 SQLite 数据库持久化存储配置。

**✨ 最新进展**：已完成阿里云 DNS 服务商集成！

## 🎉 已完成的工作

### ✅ 前端组件

1. **主框架** ([`App.vue`](d:\go\wails\myproject\frontend\src\App.vue))
   - 四个功能标签页切换
   - 现代化渐变紫色主题
   - 响应式布局

2. **状态监控** ([`StatusPanel.vue`](d:\go\wails\myproject\frontend\src\components\StatusPanel.vue))
   - 公网 IP 地址显示
   - IP 归属地和 ISP 信息
   - 域名解析状态列表
   - 自动刷新机制

3. **域名管理** ([`DomainList.vue`](d:\go\wails\myproject\frontend\src\components\DomainList.vue))
   - 添加/编辑/删除域名配置
   - ✅ **支持多个 DNS 提供商（阿里云、Cloudflare 等）**
   - ✅ **根据服务商动态显示配置项（Token / AccessKey）**
   - 启用/禁用开关
   - Token 安全存储

4. **日志查看** ([`Logs.vue`](d:\go\wails\myproject\frontend\src\components\Logs.vue))
   - 实时日志记录
   - 级别筛选和搜索
   - 导出功能
   - 统计面板

5. **设置** ([`Settings.vue`](d:\go\wails\myproject\frontend\src\components\Settings.vue))
   - 基本设置
   - 重试策略
   - 通知配置
   - 网络设置

### ✅ 后端 API 框架

已创建完整的 Go 后端接口 ([`app.go`](d:\go\wails\myproject\app.go))：
- 获取公网 IP
- ✅ **域名增删改查（支持阿里云 AccessKey 配置）**
- 日志管理
- 设置持久化
- DDNS 核心功能

### ✅ TypeScript 绑定

已更新前端类型定义 ([`models.ts`](d:\go\wails\myproject\frontend\wailsjs\go\models.ts))：
- ✅ 添加 `accessKeyID` 和 `accessKeySecret` 字段
- 保持前后端类型同步

## 🚀 立即运行

### 方法 1: 开发模式（推荐）

```bash
# 在项目根目录执行
wails dev
```

这将启动开发服务器，你可以访问 http://localhost:34115 预览界面。

### 方法 2: 生产构建

```bash
# 构建可执行文件
wails build
```

构建完成后会在 `build/bin` 目录生成可执行文件。

### ✅ 阿里云 DNS 集成 ⭐ NEW

完整实现了阿里云 DNS 服务提供商：

**核心文件**：
- [`internal/provider/aliyun.go`](d:\go\wails\myproject\internal\provider\aliyun.go) - 阿里云 DNS API 实现
- [`internal/provider/factory.go`](d:\go\wails\myproject\internal\provider\factory.go) - 提供商工厂方法
- [`pkg/types/domain.go`](d:\go\wails\myproject\pkg\types\domain.go) - 类型定义（支持多服务商）
- [`internal/domain/repository.go`](d:\go\wails\myproject\internal\domain\repository.go) - 数据访问层

**功能特性**：
- ✅ 支持 AccessKey ID 和 AccessKey Secret 验证
- ✅ 完整的 DNS A 记录增删改查
- ✅ HMAC-SHA1 签名算法
- ✅ 自动识别域名结构（根域名/子域名）
- ✅ 错误处理和日志记录
- ✅ 可扩展架构，轻松添加新服务商

**文档资源**：
- [`docs/ALIYUN_DNS_GUIDE.md`](d:\go\wails\myproject\docs\ALIYUN_DNS_GUIDE.md) - 完整配置指南
- [`docs/ALIYUN_QUICK_REFERENCE.md`](d:\go\wails\myproject\docs\ALIYUN_QUICK_REFERENCE.md) - 快速参考卡片
- [`docs/ALIYUN_IMPLEMENTATION.md`](d:\go\wails\myproject\docs\ALIYUN_IMPLEMENTATION.md) - 实现总结

**测试脚本**：
- [`scripts/test_aliyun_dns.go`](d:\go\wails\myproject\scripts\test_aliyun_dns.go) - API 调用测试
- [`scripts/db_migrate.go`](d:\go\wails\myproject\scripts\db_migrate.go) - 数据库迁移工具

## 📋 下一步工作

### 需要实现的核心逻辑

后端 Go 代码中已预留了所有必要的接口，但具体实现需要你完成：

1. **IP 检测服务** (`GetPublicIP`)
   ```go
   // 可以使用第三方 API
   // https://api.ipify.org
   // https://ip-api.com/json
   ```

2. **DDNS 更新逻辑** (`CheckAndUpdate`)
   ```go
   // 根据域名提供商调用对应的 API
   // Cloudflare, Aliyun, Tencent 等
   ```

3. **定时任务** 
   ```go
   // 使用 time.Ticker 实现定期检查
   // 根据设置的间隔执行 CheckAndUpdate
   ```

4. **DNS 提供商集成**
   ```go
   // Cloudflare API
   // 阿里云 API (Aliyun) - ✅ 已完成
   // 腾讯云 API (Tencent)
   // DNSPod API
   ```

5. **数据持久化**
   ```go
   // 使用 JSON 文件或 SQLite 存储配置
   // 保存域名列表、设置、日志等
   ```

### 推荐的实现顺序

1. ✅ **第一步**: 先运行界面看看效果
2. 🔲 **第二步**: 实现 `GetPublicIP()` 方法
3. 🔲 **第三步**: 实现域名配置的增删改查
4. 🔲 **第四步**: 集成一个 DNS 提供商（建议从 Cloudflare 开始）
5. 🔲 **第五步**: 实现定时检查和更新逻辑
6. 🔲 **第六步**: 完善日志记录和设置持久化

## 💡 使用提示

### 测试界面

即使后端逻辑还未实现，你也可以：
- 查看所有界面的布局和交互
- 点击各个按钮和表单
- 体验动画和过渡效果
- 调整设置项

### 调试技巧

在浏览器中打开开发者工具（F12），可以：
- 查看控制台输出
- 检查网络请求
- 调试 Vue 组件
- 查看存储的数据

### 修改样式

所有组件都使用了 scoped CSS，你可以：
- 修改颜色方案
- 调整字体大小
- 更改间距和布局
- 添加自定义动画

## 📁 文件结构

```
myproject/
├── app.go                          # Go 后端主逻辑
├── main.go                         # 程序入口
├── wails.json                      # Wails 配置
├── go.mod                          # Go 模块依赖
│
├── internal/                       # 内部业务逻辑
│   ├── app/                        # API 层 (暴露给前端)
│   ├── config/                     # 配置管理
│   ├── database/                   # 数据库封装
│   ├── domain/                     # 域名仓库
│   ├── log/                        # 日志模块
│   ├── monitor/                    # 监控服务
│   └── provider/                   # DNS 提供商接口
│
├── pkg/                            # 公共可复用包
│   ├── types/                      # 类型定义
│   └── utils/                      # 工具函数
│
├── frontend/                       # 前端资源
│   ├── src/
│   │   ├── App.vue                 # 主应用组件
│   │   ├── components/
│   │   │   ├── StatusPanel.vue     # 状态监控
│   │   │   ├── DomainList.vue      # 域名管理
│   │   │   ├── Logs.vue            # 日志查看
│   │   │   └── Settings.vue        # 系统设置
│   │   └── style.css               # 全局样式
│   ├── wailsjs/go/main/
│   │   ├── App.d.ts                # TypeScript 类型定义
│   │   └── App.js                  # JavaScript 绑定
│   └── README.md                   # 前端说明文档
│
├── scripts/                        # 开发辅助脚本 ⭐ NEW
│   ├── README.md                   # 脚本使用说明
│   ├── app_test.go                 # 单元测试
│   ├── test_ip.go                  # IP 功能测试
│   ├── verify_db.go                # 数据库验证工具
│   └── db_migrate.go               # 数据库迁移工具
│
└── docs/                           # 技术文档 ⭐ NEW
    ├── README.md                   # 文档目录索引
    ├── QUICKSTART.md               # 快速启动指南
    ├── QUICK_VERIFY.md             # 快速验证指南
    ├── SQLITE_GUIDE.md             # SQLite 集成指南
    ├── SQLITE_INTEGRATION.md       # SQLite 集成详细文档
    ├── DIRECTORY_STRUCTURE.md      # 目录结构说明
    ├── DEVELOPMENT_RESOURCES.md    # 开发资源管理规范
    └── PROJECT_ORGANIZATION_COMPLETE.md  # 项目整理报告
```

**💡 提示**: 
- `scripts/` 目录存放开发过程中的测试和辅助脚本
- `docs/` 目录存放技术文档和使用指南
- 详细的目录说明请查看 [`docs/DIRECTORY_STRUCTURE.md`](docs/DIRECTORY_STRUCTURE.md)

## 📨 界面截图预期

运行后你将看到：

1. **状态监控页面**
   - 大字体的公网 IP 地址
   - IP 归属地信息
   - 域名状态卡片列表

2. **域名管理页面**
   - 域名卡片网格布局
   - 添加/编辑弹窗
   - 启用/禁用开关

3. **日志页面**
   - 彩色编码的日志条目
   - 筛选和搜索栏
   - 统计卡片

4. **设置页面**
   - 分组设置项
   - 开关和输入框
   - 保存/重置按钮

## 🔧 常见问题

### Q: 为什么界面上看不到真实数据？
A: 因为后端方法目前返回的是模拟数据。你需要实现具体的业务逻辑。

### Q: 如何连接真实的 DNS 提供商 API？
A: 在 `app.go` 中的 `CheckAndUpdate` 方法里，根据 `domain.Provider` 字段调用对应的 API。

### Q: 可以自定义主题颜色吗？
A: 可以！修改各组件 `<style>` 标签中的渐变色即可。

### Q: 支持移动端吗？
A: 当前为桌面应用设计，但布局是响应式的，可以在小窗口中良好显示。

## 📞 需要帮助？

如果你在实现过程中遇到问题，可以：
- 查看 Wails 官方文档：https://wails.io
- 参考 Vue 3 文档：https://vuejs.org
- 检查 Go 代码语法：`go build`
- 查看浏览器控制台错误信息

---

**祝你开发顺利！** 🎉
