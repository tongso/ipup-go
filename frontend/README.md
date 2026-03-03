# DDNS 动态域名解析 - 前端界面说明

## 🎨 界面概览

本项目的前端基于 **Vue 3 + TypeScript** 构建，提供了美观、易用的 DDNS 管理界面。

### 功能模块

#### 1. 📊 状态监控
- **公网 IP 显示**: 实时显示当前公网 IP 地址
- **IP 详情**: 显示 IP 归属地和 ISP 信息
- **域名状态**: 展示所有已配置域名的解析状态
- **自动刷新**: 每 5 分钟自动更新一次状态

#### 2. 🌐 域名管理
- **添加域名**: 支持配置多个域名
- **DNS 提供商**: 
  - ☁️ Cloudflare
  - 🔶 阿里云 (Aliyun)
  - 🐧 腾讯云 (Tencent)
  - 🎯 DNSPod
  - 🌟 GoDaddy
- **快速操作**: 启用/禁用、编辑、删除
- **Token 安全**: Token 以脱敏形式显示

#### 3. 📝 日志查看
- **日志级别**: 成功 ✅、信息 ℹ️、警告 ⚠️、错误 ❌
- **筛选功能**: 按级别筛选、关键词搜索
- **统计面板**: 各类日志数量统计
- **导出功能**: 支持导出为 TXT 文件

#### 4. ⚙️ 设置
- **基本设置**: 自动启动、检查间隔
- **重试设置**: 最大重试次数、重试延迟
- **通知设置**: 成功/失败时通知
- **网络设置**: 代理配置、API 端点

## 🎨 设计特色

- **渐变紫色主题**: 现代化的紫蓝渐变配色
- **暗色模式**: 护眼深色背景
- **流畅动画**: 平滑的过渡和悬停效果
- **响应式布局**: 自适应不同窗口大小
- **图标系统**: Emoji 图标增强可读性

## 🚀 开发环境运行

### 前提条件
```bash
# 确保已安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 进入 frontend 目录安装依赖
cd frontend
npm install
```

### 启动开发服务器
```bash
# 在项目根目录执行
wails dev
```

访问 http://localhost:34115 可以在浏览器中预览界面。

### 生产构建
```bash
wails build
```

## 📦 组件结构

```
frontend/src/
├── App.vue              # 主应用组件
├── components/
│   ├── StatusPanel.vue  # 状态监控面板
│   ├── DomainList.vue   # 域名管理列表
│   ├── Logs.vue         # 日志查看器
│   └── Settings.vue     # 设置页面
└── style.css            # 全局样式
```

## 💡 使用提示

1. **首次使用**: 
   - 前往「域名管理」添加你的第一个域名
   - 选择对应的 DNS 提供商
   - 填写 API Token/密钥

2. **查看状态**:
   - 在「状态监控」查看所有域名的解析状态
   - 绿色表示正常，橙色表示等待，红色表示错误

3. **排查问题**:
   - 使用「日志查看」功能查看详细记录
   - 根据错误信息调整配置

4. **优化性能**:
   - 在「设置」中调整检查间隔
   - 配置合适的重试策略

## 🔧 后续开发建议

### 需要实现的后端功能

1. **IP 检测服务**
   ```go
   func GetPublicIP() string
   ```

2. **域名管理 API**
   ```go
   func AddDomain(domain DomainConfig) error
   func UpdateDomain(id int, domain DomainConfig) error
   func DeleteDomain(id int) error
   func ListDomains() []DomainConfig
   ```

3. **DDNS 核心功能**
   ```go
   func StartDDNS(domain DomainConfig)
   func StopDDNS(domainID int)
   func CheckAndUpdate(domain DomainConfig)
   ```

4. **日志系统**
   ```go
   func GetLogs(level string, keyword string) []LogEntry
   func ExportLogs() string
   func ClearLogs()
   ```

5. **设置持久化**
   ```go
   func SaveSettings(settings Settings) error
   func LoadSettings() Settings
   func ResetSettings()
   ```

### 可选增强功能

- 📊 图表统计（IP 变化趋势）
- 🔔 桌面通知
- 📱 移动端适配优化
- 🌍 多语言支持
- 🎨 主题切换
- 📥 配置导入导出

## 📝 注意事项

1. **安全性**: API Token 会加密存储在前端配置中
2. **性能**: 建议检查间隔不少于 300 秒（5 分钟）
3. **兼容性**: 支持主流桌面操作系统（Windows、macOS、Linux）
4. **浏览器**: 开发调试时建议使用现代浏览器（Chrome、Edge、Firefox）

## 🤝 贡献指南

如果你要扩展功能，建议：
1. 保持代码风格一致
2. 添加适当的注释
3. 测试各种边界情况
4. 更新相关文档

---

**开发工具**: Wails v2.11.0 + Vue 3 + TypeScript  
**UI 框架**: 原生 CSS + 自定义组件  
**构建工具**: Vite
