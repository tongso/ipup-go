package database

import (
	"database/sql"
	"fmt"
	
	_ "modernc.org/sqlite"
)

// Database 数据库连接管理器
type Database struct {
	db *sql.DB
}

// NewDatabase 创建新的数据库连接
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败：%w", err)
	}
	
	d := &Database{db: db}
	
	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败：%w", err)
	}
	
	return d, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetDB 获取底层 sql.DB 对象
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// CreateTables 创建数据库表
func (d *Database) CreateTables() error {
	// 创建域名配置表（包含阿里云 AccessKey 字段）
	domainTable := `
	CREATE TABLE IF NOT EXISTS domains (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL UNIQUE,
		provider TEXT NOT NULL,
		token TEXT NOT NULL DEFAULT '',
		access_key_id TEXT NOT NULL DEFAULT '',
		access_key_secret TEXT NOT NULL DEFAULT '',
		interval INTEGER NOT NULL DEFAULT 300,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		current_ip TEXT,
		last_update DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	
	// 创建系统设置表
	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	
	// 创建日志表
	logsTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		level TEXT NOT NULL,
		domain TEXT,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	
	// 创建索引
	indexSQL := `
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_domains_enabled ON domains(enabled);
	`
	
	// 执行创建表的 SQL
	if _, err := d.db.Exec(domainTable); err != nil {
		return fmt.Errorf("创建 domains 表失败：%w", err)
	}
	
	if _, err := d.db.Exec(settingsTable); err != nil {
		return fmt.Errorf("创建 settings 表失败：%w", err)
	}
	
	if _, err := d.db.Exec(logsTable); err != nil {
		return fmt.Errorf("创建 logs 表失败：%w", err)
	}
	
	if _, err := d.db.Exec(indexSQL); err != nil {
		return fmt.Errorf("创建索引失败：%w", err)
	}
	
	return nil
}

// InitDefaults 初始化默认设置
func (d *Database) InitDefaults() error {
	defaultSettings := []struct {
		key   string
		value string
	}{
		{"autoStart", "true"},
		{"checkInterval", "300"},
		{"retryCount", "3"},
		{"retryDelay", "10"},
		{"logLevel", "info"},
		{"timezone", "Local"}, // 默认使用本地时区
		{"notifySuccess", "false"},
		{"notifyError", "true"},
		{"proxy", ""},
		{"apiEndpoint", "https://api.ipify.org"},
	}
	
	for _, setting := range defaultSettings {
		insertSQL := `INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`
		if _, err := d.db.Exec(insertSQL, setting.key, setting.value); err != nil {
			return fmt.Errorf("插入默认设置失败：%w", err)
		}
	}
	
	return nil
}
