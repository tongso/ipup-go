package app

import (
	"fmt"
	
	"ipup-go/pkg/types"
)

// ==================== 系统设置 API ====================

// GetSettings 获取系统设置
func (a *App) GetSettings() (types.Settings, error) {
	settings, err := a.configMgr.Load()
	if err != nil {
		return types.Settings{}, fmt.Errorf("加载设置失败：%w", err)
	}
	return settings, nil
}

// SaveSettings 保存系统设置
func (a *App) SaveSettings(settings types.Settings) error {
	if err := a.configMgr.Save(settings); err != nil {
		return fmt.Errorf("保存设置失败：%w", err)
	}
	
	a.addLog("info", "", "系统设置已保存")
	return nil
}

// LoadSettings 加载系统设置（已废弃，使用 GetSettings）
func (a *App) LoadSettings() (types.Settings, error) {
	return a.GetSettings()
}

// ResetSettings 重置系统设置为默认值
func (a *App) ResetSettings() error {
	if err := a.configMgr.Reset(); err != nil {
		return fmt.Errorf("重置设置失败：%w", err)
	}
	
	a.addLog("info", "", "系统设置已重置为默认值")
	return nil
}
