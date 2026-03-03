# 项目目录结构说明

## 📁 完整目录树

```
myproject/
├── frontend/                 # 前端项目目录
│   ├── src/
│   │   ├── components/      # Vue 组件
│   │   ├── App.vue          # 主应用组件
│   │   └── main.ts          # 入口文件
│   ├── wailsjs/             # Wails 生成的绑定
│   └── README.md            # 前端开发文档
│
├── internal/                # 内部业务逻辑（不可外部引用）
│   ├── app/                 # API 层（暴露给前端）
│   ├── config/              # 配置管理
│   ├── database/            # 数据库封装
│   ├── domain/              # 域名仓库
│   ├── log/                 # 日志模块
│   ├── monitor/             # 监控服务
│   ├── notify/              # 通知模块
│   └── provider/            # DNS 提供商
│
├── pkg/                     # 公共可复用包
│   ├── types/               # 类型定义
│   └── utils/               # 工具函数
│
├── scripts/                 # 开发辅助脚本
│   ├── db_migrate.go        # 数据库迁移工具
│   ├── verify_db.go         # 数据库验证工具
│   ├── test_ip.go           # IP 功能测试脚本
│   ├── app_test.go          # 单元测试
│   └── README.md            # 脚本使用说明
│
├── docs/                    # 项目文档
│   ├── QUICKSTART.md        # 快速启动指南
│   ├── QUICK_VERIFY.md      # 快速验证指南
│   ├── SQLITE_GUIDE.md      # SQLite 集成指南
│   ├── SQLITE_INTEGRATION.md # SQLite 集成详细文档
│   └── README.md            # 文档目录说明
│
├── main.go                  # 主程序入口
├── app.go                   # Wails 应用配置
├── README.md                # 项目主文档
├── wails.json               # Wails 配置文件
└── go.mod                   # Go 模块依赖
```

## 🎯 目录职责

### `internal/` - 业务逻辑层
- **app/**: API 暴露层，直接对接前端调用
- **config/**: 系统设置管理和持久化
- **database/**: 数据库连接和表结构管理
- **domain/**: 域名配置的 CRUD 操作
- **log/**: 日志记录功能
- **monitor/**: DDNS 监控和定时检查
- **notify/**: Webhook 通知推送
- **provider/**: DNS 服务商接口实现

### `pkg/` - 公共包
- **types/**: 所有业务实体类型定义
  - `Domain`: 域名配置结构
  - `Settings`: 系统设置结构
  - `LogEntry`: 日志条目结构
- **utils/**: 通用工具函数
  - `ip.go`: IP 地址相关工具

### `scripts/` - 开发工具
存放不參與正式編譯的輔助腳本：
- 数据库迁移和备份
- 功能测试和验证
- 单元测试
- 数据导入导出

### `docs/` - 技术文档
存放项目开发过程中的技术文档：
- 快速入门指南
- 技术集成文档
- 最佳实践说明
- 问题排查指南

### `frontend/` - 前端资源
- Vue 3 + TypeScript 组件
- Wails JS 绑定
- Vite 构建配置
- 静态资源文件

## 📋 文件组织原则

1. **单一职责**: 每个目录只负责一个特定领域
2. **依赖分层**: 
   - `main` → `app` → `internal/*` → `database`
   - `internal` 和 `pkg` 不依赖 `app`
3. **类型安全**: 所有业务类型统一在 `pkg/types` 定义
4. **可测试性**: 测试脚本独立在 `scripts/`，不影响主应用编译
5. **文档分离**: 
   - 核心文档在项目根目录（README.md）
   - 开发文档在 `docs/`
   - 前端文档在 `frontend/README.md`

## 🚀 开发工作流

```bash
# 1. 开发模式运行
wails dev

# 2. 运行测试脚本
go run scripts/verify_db.go
go run scripts/test_ip.go

# 3. 运行单元测试
go test -v ./scripts/*.go

# 4. 构建生产版本
wails build

# 5. 查看文档
cat docs/README.md
cat scripts/README.md
```

## 🔒 访问控制

- `internal/`: Go 编译器强制限制，只能在项目内部使用
- `pkg/`: 可以被任何外部项目引用
- `scripts/`: 独立的可执行脚本，不参与主应用编译
