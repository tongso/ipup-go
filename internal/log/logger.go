package log

import (
	"database/sql"
	"fmt"
	"time"
	
	"ipup-go/pkg/types"
)

// Logger 日志记录器
type Logger struct {
	db       *sql.DB
	timezone string // 时区设置
}

// NewLogger 创建日志记录器
func NewLogger(db *sql.DB) *Logger {
	return &Logger{
		db:       db,
		timezone: "Local", // 默认使用本地时区
	}
}

// SetTimezone 设置时区
func (l *Logger) SetTimezone(timezone string) {
	if timezone == "" || timezone == "Local" {
		l.timezone = "Local"
	} else {
		l.timezone = timezone
	}
}

// GetTimezone 获取当前时区
func (l *Logger) GetTimezone() string {
	return l.timezone
}

// formatTimestamp 格式化时间戳为指定时区
func (l *Logger) formatTimestamp(t time.Time) string {
	var loc *time.Location
	var err error
	
	if l.timezone == "Local" {
		loc = time.Local
	} else {
		loc, err = time.LoadLocation(l.timezone)
		if err != nil {
			// 如果时区加载失败，回退到本地时区
			loc = time.Local
		}
	}
	
	// 转换为指定时区并格式化
	return t.In(loc).Format("2006-01-02 15:04:05")
}

// Add 添加日志记录
func (l *Logger) Add(level, domain, message string) error {
	insertSQL := `
	INSERT INTO logs (timestamp, level, domain, message, created_at) 
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	
	// 存储 UTC 时间到数据库
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	
	_, err := l.db.Exec(insertSQL, timestamp, level, domain, message)
	if err != nil {
		return fmt.Errorf("插入日志失败：%w", err)
	}
	
	return nil
}

// Get 获取日志列表（返回已转换时区的时间）
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
		
		// 将 UTC 时间转换为指定时区显示
		entry.Timestamp = l.convertTimezone(entry.Timestamp)
		
		logs = append(logs, entry)
	}
	
	return logs, nil
}

// convertTimezone 将 UTC 时间转换为用户配置的时区
func (l *Logger) convertTimezone(utcTimestamp string) string {
	var utcTime time.Time
	var err error
	
	// 尝试多种时间格式
	formats := []string{
		"2006-01-02 15:04:05",           // 标准格式
		"2006-01-02T15:04:05Z",          // ISO 8601 格式（带 Z）
		"2006-01-02T15:04:05",           // ISO 8601 格式（不带 Z）
		time.RFC3339,                     // RFC3339 格式
	}
	
	// 尝试每种格式直到解析成功
	for _, format := range formats {
		utcTime, err = time.Parse(format, utcTimestamp)
		if err == nil {
			break
		}
	}
	
	// 如果所有格式都解析失败，直接返回原值
	if err != nil {
		return utcTimestamp
	}
	
	// 根据配置的时区转换
	var loc *time.Location
	if l.timezone == "" || l.timezone == "Local" {
		loc = time.Local
	} else {
		loc, err = time.LoadLocation(l.timezone)
		if err != nil {
			// 时区加载失败，回退到本地时区
			loc = time.Local
		}
	}
	
	// UTC 时间转换为指定时区并格式化
	return utcTime.In(loc).Format("2006-01-02 15:04:05")
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
		
		// 将 UTC 时间转换为指定时区
		displayTime := l.convertTimezone(timestamp)
		
		result += fmt.Sprintf("[%s] [%s] %s: %s\n", displayTime, level, domain, message)
	}
	
	return result, nil
}
