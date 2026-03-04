package app

import (
	"fmt"
	
	"ipup-go/pkg/types"
)

// ==================== 日志管理 API ====================

// GetLogs 获取日志列表
func (a *App) GetLogs(level, keyword string) ([]types.LogEntry, error) {
	logs, err := a.logger.Get(level, keyword, 100)
	if err != nil {
		return nil, fmt.Errorf("查询日志失败：%w", err)
	}
	
	// 调试输出
	fmt.Printf("📝 查询到 %d 条日志\n", len(logs))
	for i, log := range logs {
		fmt.Printf("   [%d] %s - %s - %s: %s\n", i, log.Timestamp, log.Level, log.Domain, log.Message)
	}
	
	return logs, nil
}

// ClearLogs 清空所有日志
func (a *App) ClearLogs() error {
	if err := a.logger.Clear(); err != nil {
		return fmt.Errorf("清空日志失败：%w", err)
	}
	
	a.addLog("info", "", "日志已清空")
	return nil
}

// ExportLogs 导出日志为文本
func (a *App) ExportLogs() (string, error) {
	content, err := a.logger.Export()
	if err != nil {
		return "", fmt.Errorf("导出日志失败：%w", err)
	}
	return content, nil
}
