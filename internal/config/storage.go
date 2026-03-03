package config

import (
	"database/sql"
	"fmt"
	
	"ipup-go/pkg/types"
)

// SettingsStorage 设置存储管理器
type SettingsStorage struct {
	db *sql.DB
}

// NewSettingsStorage 创建设置存储管理器
func NewSettingsStorage(db *sql.DB) *SettingsStorage {
	return &SettingsStorage{db: db}
}

// Save 保存所有设置
func (ss *SettingsStorage) Save(settings types.Settings) error {
	settingsMap := map[string]interface{}{
		"autoStart":     settings.AutoStart,
		"checkInterval": settings.CheckInterval,
		"retryCount":    settings.RetryCount,
		"retryDelay":    settings.RetryDelay,
		"logLevel":      settings.LogLevel,
		"notifySuccess": settings.NotifySuccess,
		"notifyError":   settings.NotifyError,
		"proxy":         settings.Proxy,
		"apiEndpoint":   settings.APIEndpoint,
	}
	
	for key, value := range settingsMap {
		if err := ss.saveSingle(key, value); err != nil {
			return err
		}
	}
	
	return nil
}

// saveSingle 保存单个设置项
func (ss *SettingsStorage) saveSingle(key string, value interface{}) error {
	var valueStr string
	switch v := value.(type) {
	case bool:
		if v {
			valueStr = "true"
		} else {
			valueStr = "false"
		}
	case int:
		valueStr = fmt.Sprintf("%d", v)
	case string:
		valueStr = v
	}
	
	updateSQL := `
	INSERT OR REPLACE INTO settings (key, value, updated_at) 
	VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	
	if _, err := ss.db.Exec(updateSQL, key, valueStr); err != nil {
		return fmt.Errorf("保存设置 %s 失败：%w", key, err)
	}
	
	return nil
}

// Load 加载所有设置
func (ss *SettingsStorage) Load() (types.Settings, error) {
	settings := types.Settings{
		AutoStart:      true,
		CheckInterval:  300,
		RetryCount:     3,
		RetryDelay:     10,
		LogLevel:       "info",
		NotifySuccess:  false,
		NotifyError:    true,
		Proxy:          "",
		APIEndpoint:    "https://api.ipify.org",
	}
	
	querySQL := `SELECT key, value FROM settings`
	
	rows, err := ss.db.Query(querySQL)
	if err != nil {
		return settings, fmt.Errorf("加载设置失败：%w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		
		switch key {
		case "autoStart":
			settings.AutoStart = (value == "true")
		case "checkInterval":
			fmt.Sscanf(value, "%d", &settings.CheckInterval)
		case "retryCount":
			fmt.Sscanf(value, "%d", &settings.RetryCount)
		case "retryDelay":
			fmt.Sscanf(value, "%d", &settings.RetryDelay)
		case "logLevel":
			settings.LogLevel = value
		case "notifySuccess":
			settings.NotifySuccess = (value == "true")
		case "notifyError":
			settings.NotifyError = (value == "true")
		case "proxy":
			settings.Proxy = value
		case "apiEndpoint":
			settings.APIEndpoint = value
		}
	}
	
	return settings, nil
}

// Reset 重置为默认设置
func (ss *SettingsStorage) Reset() error {
	// 删除所有设置
	_, err := ss.db.Exec("DELETE FROM settings")
	if err != nil {
		return fmt.Errorf("删除设置失败：%w", err)
	}
	
	// 重新插入默认设置
	defaultSettings := []struct {
		key   string
		value string
	}{
		{"autoStart", "true"},
		{"checkInterval", "300"},
		{"retryCount", "3"},
		{"retryDelay", "10"},
		{"logLevel", "info"},
		{"notifySuccess", "false"},
		{"notifyError", "true"},
		{"proxy", ""},
		{"apiEndpoint", "https://api.ipify.org"},
	}
	
	insertSQL := `INSERT INTO settings (key, value) VALUES (?, ?)`
	for _, setting := range defaultSettings {
		if _, err := ss.db.Exec(insertSQL, setting.key, setting.value); err != nil {
			return fmt.Errorf("插入默认设置失败：%w", err)
		}
	}
	
	return nil
}
