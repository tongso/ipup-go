package log

import (
	"database/sql"
	"fmt"
	
	"ipup-go/pkg/types"
)

// Logger 日志记录器
type Logger struct {
	db *sql.DB
}

// NewLogger 创建日志记录器
func NewLogger(db *sql.DB) *Logger {
	return &Logger{db: db}
}

// Add 添加日志记录
func (l *Logger) Add(level, domain, message string) error {
	insertSQL := `
	INSERT INTO logs (timestamp, level, domain, message, created_at) 
	VALUES (CURRENT_TIMESTAMP, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	
	_, err := l.db.Exec(insertSQL, level, domain, message)
	if err != nil {
		return fmt.Errorf("插入日志失败：%w", err)
	}
	
	return nil
}

// Get 获取日志列表
func (l *Logger) Get(level, keyword string, limit int) ([]types.LogEntry, error) {
	// 处理 level 参数：如果是 'all'，则不过滤级别
	levelParam := ""
	if level != "all" && level != "" {
		levelParam = level
	}
	
	querySQL := `
	SELECT id, timestamp, level, domain, message
	FROM logs
	WHERE (? = '' OR level = ?) AND (? = '' OR domain LIKE ? OR message LIKE ?)
	ORDER BY timestamp DESC
	LIMIT ?
	`
	
	rows, err := l.db.Query(querySQL, levelParam, levelParam, keyword, "%"+keyword+"%", "%"+keyword+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("查询日志失败：%w", err)
	}
	defer rows.Close()
	
	var logs []types.LogEntry
	for rows.Next() {
		var entry types.LogEntry
		if err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.Level, &entry.Domain, &entry.Message); err != nil {
			continue
		}
		logs = append(logs, entry)
	}
	
	return logs, nil
}

// Clear 清空所有日志
func (l *Logger) Clear() error {
	_, err := l.db.Exec("DELETE FROM logs")
	if err != nil {
		return fmt.Errorf("清空日志失败：%w", err)
	}
	return nil
}

// Count 统计日志数量
func (l *Logger) Count() (int, error) {
	var count int
	err := l.db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计日志失败：%w", err)
	}
	return count, nil
}

// Export 导出日志为文本
func (l *Logger) Export() (string, error) {
	querySQL := `
	SELECT timestamp, level, domain, message
	FROM logs
	ORDER BY timestamp DESC
	LIMIT 1000
	`
	
	rows, err := l.db.Query(querySQL)
	if err != nil {
		return "", fmt.Errorf("查询日志失败：%w", err)
	}
	defer rows.Close()
	
	result := "=== IPUP-DDNS Logs ===\n\n"
	for rows.Next() {
		var timestamp, level, domain, message string
		if err := rows.Scan(&timestamp, &level, &domain, &message); err != nil {
			continue
		}
		result += fmt.Sprintf("[%s] [%s] %s: %s\n", timestamp, level, domain, message)
	}
	
	return result, nil
}
