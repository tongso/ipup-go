# SQLite 集成 - 快速开始

## 🚀 5 分钟快速上手

### 步骤 1: 下载依赖（必需）

打开终端，在项目根目录执行：

**Windows PowerShell:**
```powershell
$env:GOPROXY = "https://goproxy.cn,direct"
go mod tidy
```

**Linux/Mac:**
```bash
export GOPROXY=https://goproxy.cn,direct
go mod tidy
```

这会下载 SQLite 驱动，通常需要几秒钟。

### 步骤 2: 验证安装

```bash
go build
```

如果没有报错，说明安装成功 ✅

### 步骤 3: 运行测试（可选）

```bash
go test -v
```

看到所有测试通过即为成功。

### 步骤 4: 启动应用

```bash
wails dev
```

应用会自动创建 `ipup.db` 数据库文件！

---

## 📦 已集成的功能

### ✅ 域名配置持久化
- 添加、编辑、删除域名
- 启用/禁用快速切换
- 自动记录更新时间

### ✅ 系统设置存储
- 所有设置项永久保存
- 一键重置为默认值
- 类型安全的数据转换

### ✅ 日志记录系统
- 四种日志级别
- 关键词搜索
- 导出功能
- 自动清理（最多 1000 条）

### ✅ 线程安全保障
- 读写锁保护
- 并发安全
- 资源自动释放

---

## ⚠️ 常见问题

### Q: 下载失败怎么办？

**使用国内镜像加速：**
```bash
# Windows PowerShell
$env:GOPROXY = "https://goproxy.cn,direct"
go mod tidy

# Linux/Mac
export GOPROXY=https://goproxy.cn,direct
go mod tidy
```

### Q: 编译时报错 "CGO required"？

SQLite 需要 CGO 支持，请安装 GCC 编译器：

**Windows:** 下载并安装 [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
**macOS:** `xcode-select --install`
**Linux:** `sudo apt-get install gcc`

### Q: 数据库文件在哪里？

运行应用后，会在可执行文件同目录创建 `ipup.db` 文件。

开发环境：`./ipup.db`

---

## 📖 详细文档

- 📘 [完整集成说明](./SQLITE_INTEGRATION.md)
- 📗 [数据库结构详解](./DATABASE.md)
- 📙 [使用指南](./SQLITE_GUIDE.md)
- 📕 [安装教程](./INSTALL_SQLITE.md)

---

## 💡 下一步

1. **先执行步骤 1** 下载依赖
2. **运行测试** 验证功能正常
3. **启动应用** 体验完整功能
4. **查看文档** 了解更多细节

---

**预计耗时**: 2-5 分钟  
**难度**: ⭐☆☆☆☆ (简单)
