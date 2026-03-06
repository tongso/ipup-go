package app

import (
	"context"
	"fmt"
	
	"ipup-go/internal/config"
	"ipup-go/internal/database"
	"ipup-go/internal/domain"
	"ipup-go/internal/log"
	"ipup-go/internal/monitor"
)

// App Wails 应用结构体
type App struct {
	ctx          context.Context
	db           *database.Database
	configMgr    *config.SettingsStorage
	domainRepo   *domain.Repository
	logger       *log.Logger
	monitorSvc   *monitor.MonitorService
}

// NewApp 创建新的 App 实例
func NewApp() *App {
	return &App{}
}

// Startup 应用启动时初始化
func (a *App) Startup(ctx context.Context) error {
	a.ctx = ctx
	
	// 初始化数据库
	var err error
	a.db, err = database.NewDatabase("./ipup.db")
	if err != nil {
		return fmt.Errorf("初始化数据库失败：%w", err)
	}
	
	// 创建表
	if err := a.db.CreateTables(); err != nil {
		return fmt.Errorf("创建数据库表失败：%w", err)
	}
	
	// 初始化默认设置
	if err := a.db.InitDefaults(); err != nil {
		return fmt.Errorf("初始化默认设置失败：%w", err)
	}
	
	// 初始化各模块
	db := a.db.GetDB()
	a.configMgr = config.NewSettingsStorage(db)
	a.domainRepo = domain.NewRepository(db)
	a.logger = log.NewLogger(db)
	
	// 加载设置并应用时区
	settings, err := a.configMgr.Load()
	if err == nil && settings.Timezone != "" {
		a.logger.SetTimezone(settings.Timezone)
	}
	
	// 初始化监控服务
	checker := monitor.NewChecker("") // API endpoint 从配置加载
	a.monitorSvc = monitor.NewMonitorService(a.domainRepo, a.logger, checker)
	
	a.addLog("info", "", "应用初始化完成")
	return nil
}

// Shutdown 应用关闭时清理
func (a *App) Shutdown(ctx context.Context) error {
	// 停止监控服务
	if a.monitorSvc != nil {
		a.monitorSvc.Stop()
	}
	
	// 关闭数据库连接
	if a.db != nil {
		return a.db.Close()
	}
	
	return nil
}

// addLog 添加日志（内部方法）
func (a *App) addLog(level, domain, message string) {
	if a.logger != nil {
		a.logger.Add(level, domain, message)
	}
}
