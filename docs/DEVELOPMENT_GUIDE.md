# 开发环境搭建指南

## 🎯 快速开始

本文档指导你从零开始搭建 myproject 的开发环境。

---

## 📋 前置要求

### 必需软件

| 软件 | 版本要求 | 下载地址 | 备注 |
|------|---------|---------|------|
| Go | 1.23+ | https://golang.org/dl/ | 后端开发 |
| Node.js | 16+ | https://nodejs.org/ | 前端构建 |
| Git | 最新 | https://git-scm.com/ | 版本控制 |

### 可选工具

- **VS Code** - 推荐编辑器
- **Go 插件** - Go 语言支持
- **Volar** - Vue 3 + TypeScript 支持

---

## 🚀 安装步骤

### Step 1: 安装 Go

#### Windows
```powershell
# 下载并安装 Go 1.23
# 访问 https://golang.org/dl/ 下载 Windows 安装包
# 运行安装包，按提示完成安装

# 验证安装
go version
# 输出：go version go1.23.x windows/amd64
```

#### macOS
```bash
# 使用 Homebrew
brew install go@1.23

# 验证安装
go version
```

#### Linux
```bash
# 下载并解压
wget https://golang.org/dl/go1.23.x.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.x.linux-amd64.tar.gz

# 添加到 PATH
export PATH=$PATH:/usr/local/go/bin

# 验证安装
go version
```

---

### Step 2: 安装 Wails CLI

```bash
# 全局安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 验证安装
wails version
# 输出：v2.11.0
```

**⚠️ 注意：** 确保 `$GOPATH/bin` 在系统 PATH 中

#### Windows 添加 GOPATH 到 PATH
```powershell
# PowerShell（临时）
$env:Path += ";$HOME\go\bin"

# 永久添加：系统属性 → 环境变量 → Path → 新建 → %USERPROFILE%\go\bin
```

---

### Step 3: 克隆项目

```bash
# 克隆项目到本地
git clone <your-repo-url>
cd myproject
```

---

### Step 4: 安装后端依赖

```bash
# 下载 Go 模块依赖
go mod download

# 验证依赖完整性
go mod verify
```

---

### Step 5: 安装前端依赖

```bash
# 进入前端目录
cd frontend

# 安装 npm 依赖
npm install

# 或使用 pnpm（如果项目使用 pnpm）
pnpm install

# 返回项目根目录
cd ..
```

---

### Step 6: 启动开发环境

```bash
# 在项目根目录执行
wails dev
```

**预期输出：**
```
  ___   __     ___   ___ 
 / _ \ / /__  / _ \ / _ \
/ //_/ / / _ \/ //_/ // /
\____/_/_//_/____/\____/ 

Version: v2.11.0
...

[DEV] Server started on http://localhost:34115
```

**首次启动可能需要几分钟**（编译和初始化）。

---

## 🔧 配置开发环境

### VS Code 推荐插件

1. **Go** (官方)
   - 代码补全、调试、格式化
   
2. **Volar** (Vue 3 + TypeScript)
   - Vue 文件语法高亮、智能提示
   
3. **ESLint**
   - 前端代码检查

### settings.json 配置

在项目根目录创建 `.vscode/settings.json`：

```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "editor.formatOnSave": true,
    "[vue]": {
        "editor.defaultFormatter": "Vue.volar"
    },
    "[typescript]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode"
    },
    "[go]": {
        "editor.defaultFormatter": "golang.go"
    }
}
```

---

## 📁 项目结构概览

```
myproject/
├── cmd/                 # 应用启动入口
├── internal/            # 内部业务逻辑
│   ├── app/            # API 暴露层
│   ├── config/         # 配置管理
│   ├── database/       # 数据库操作
│   ├── domain/         # 域名管理
│   ├── log/            # 日志系统
│   ├── monitor/        # 监控服务
│   ├── notify/         # 通知服务
│   └── provider/       # 第三方服务提供商
├── pkg/                # 公共包
│   └── types/          # 类型定义
├── scripts/            # 工具和测试脚本
├── docs/               # 项目文档
├── frontend/           # 前端代码
│   ├── src/
│   │   ├── components/ # Vue 组件
│   │   ├── App.vue     # 主组件
│   │   └── main.ts     # 入口文件
│   ├── wailsjs/        # Wails 绑定（自动生成）
│   └── package.json
├── main.go             # 程序入口
├── app.go              # 应用主结构
├── wails.json          # Wails 配置
└── go.mod              # Go 模块配置
```

---

## 🏃 常用开发命令

### 开发模式
```bash
wails dev
```
- 热重载前端代码
- 自动重启后端
- 监听文件变化

### 生产构建
```bash
wails build
```
- 编译为独立可执行文件
- 前端打包嵌入二进制
- 输出到 `build/bin/`

### 生成绑定
```bash
wails generate module
```
- 重新生成前端 TypeScript 类型
- Go 方法变更后必须执行

### 运行测试
```bash
# Go 单元测试
go test ./...

# 运行测试脚本
go run scripts/test_*.go
```

---

## 🐛 常见问题

### Q1: `wails dev` 启动失败

**错误：** `command not found: wails`

**解决：**
```bash
# 确认 GOPATH/bin 在 PATH 中
echo $GOPATH/bin
echo $PATH

# Windows PowerShell
$env:Path += ";$HOME\go\bin"
```

---

### Q2: 前端依赖安装失败

**错误：** `npm ERR! network timeout`

**解决：**
```bash
# 切换淘宝镜像
npm config set registry https://registry.npmmirror.com

# 重新安装
rm -rf node_modules package-lock.json
npm install
```

---

### Q3: Go 模块下载慢

**错误：** `dial tcp: lookup proxy.golang.org: no such host`

**解决：**
```bash
# 设置国内代理
go env -w GOPROXY=https://goproxy.cn,direct

# 或禁用代理（如果可以访问）
go env -w GOPROXY=direct
```

---

### Q4: SQLite 编译错误

**错误：** `gcc not found` 或 `CGO_ENABLED=0`

**解决：**
```bash
# Windows: 安装 TDM-GCC 或 MinGW
# macOS: 安装 Xcode Command Line Tools
xcode-select --install

# Linux: 安装 build-essential
sudo apt-get install build-essential

# 确保 CGO 启用
go env CGO_ENABLED
# 应该输出：1
```

---

### Q5: 前端调用 Go 方法失败

**错误：** `window.go.app is undefined`

**解决：**
1. 确保 `wails dev` 正在运行
2. 检查 Go 方法是否首字母大写（公开）
3. 重新生成绑定：`wails generate module`
4. 刷新浏览器页面

---

## 📚 下一步

完成环境搭建后，建议阅读：

1. **[项目开发规范](./PROJECT_SPECIFICATIONS.md)** - 了解编码规范
2. **[架构说明](./ARCHITECTURE_OVERVIEW.md)** - 理解整体架构
3. **[API 集成指南](./API_INTEGRATION_GUIDE.md)** - 学习第三方 API 集成
4. **[QUICKSTART.md](../QUICKSTART.md)** - 快速上手指南

---

## 🆘 获取帮助

遇到问题时：

1. 查看 [LESSONS_LEARNED.md](./LESSONS_LEARNED.md) - 踩坑记录
2. 搜索 GitHub Issues
3. 加入 Wails Discord 社区
4. 查看官方文档：https://wails.io/docs/

---

*最后更新：2026-03-06*
