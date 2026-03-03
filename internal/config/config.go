package config

import (
	"database/sql"
	"fmt"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	db *sql.DB
}

// NewConfigManager 创建配置管理器
func NewConfigManager(db *sql.DB) *ConfigManager {
	return &ConfigManager{db: db}
}

// Save 保存配置项
func (cm *ConfigManager) Save(key string, value interface{}) error {
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
	default:
		return fmt.Errorf("不支持的配置类型：%T", value)
	}
	
	updateSQL := `
	INSERT OR REPLACE INTO settings (key, value, updated_at) 
	VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	
	if _, err := cm.db.Exec(updateSQL, key, valueStr); err != nil {
		return fmt.Errorf("保存配置 %s 失败：%w", key, err)
	}
	
	return nil
}

// Load 加载配置项
func (cm *ConfigManager) Load(key string, defaultValue interface{}) (interface{}, error) {
	querySQL := `SELECT value FROM settings WHERE key = ?`
	
	var value string
	err := cm.db.QueryRow(querySQL, key).Scan(&value)
	if err == sql.ErrNoRows {
		return defaultValue, nil
	}
	if err != nil {
		return nil, fmt.Errorf("加载配置 %s 失败：%w", key, err)
	}
	
	// 根据默认值类型转换
	switch defaultValue.(type) {
	case bool:
		return value == "true", nil
	case int:
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result, nil
	case string:
		return value, nil
	}
	
	return value, nil
}

// LoadBool 加载布尔型配置
func (cm *ConfigManager) LoadBool(key string, defaultValue bool) (bool, error) {
	val, err := cm.Load(key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	if b, ok := val.(bool); ok {
		return b, nil
	}
	return defaultValue, nil
}

// LoadInt 加载整型配置
func (cm *ConfigManager) LoadInt(key string, defaultValue int) (int, error) {
	val, err := cm.Load(key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	if i, ok := val.(int); ok {
		return i, nil
	}
	return defaultValue, nil
}

// LoadString 加载字符串配置
func (cm *ConfigManager) LoadString(key string, defaultValue string) (string, error) {
	val, err := cm.Load(key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return defaultValue, nil
}
