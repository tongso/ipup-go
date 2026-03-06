package types

// Domain 域名配置结构
type Domain struct {
	ID              int    `json:"id"`
	Domain          string `json:"domain"`
	Provider        string `json:"provider"`
	Token           string `json:"token"`            // 通用 Token（如 Cloudflare）
	AccessKeyID     string `json:"accessKeyID"`      // 阿里云 AccessKey ID
	AccessKeySecret string `json:"accessKeySecret"`  // 阿里云 AccessKey Secret
	Interval        int    `json:"interval"`
	Enabled         bool   `json:"enabled"`
	CurrentIP       string `json:"currentIP"`
	LastUpdate      string `json:"lastUpdate"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// DomainStatus 域名状态结构
type DomainStatus struct {
	ID           int    `json:"id"`             // 域名 ID
	Domain       string `json:"domain"`
	CurrentIP    string `json:"currentIP"`
	LastUpdate   string `json:"lastUpdate"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	Provider     string `json:"provider"`     // DNS 提供商
	LastAPICall  string `json:"lastApiCall"`  // 最近一次 API 调用时间
	APIStatus    string `json:"apiStatus"`    // API 调用状态：success/error/pending
	APIMessage   string `json:"apiMessage"`   // API 调用消息
}

// Settings 系统设置结构
type Settings struct {
	AutoStart      bool   `json:"autoStart"`
	CheckInterval  int    `json:"checkInterval"`
	RetryCount     int    `json:"retryCount"`
	RetryDelay     int    `json:"retryDelay"`
	LogLevel       string `json:"logLevel"`
	Timezone       string `json:"timezone"`       // 时区设置
	NotifySuccess  bool   `json:"notifySuccess"`
	NotifyError    bool   `json:"notifyError"`
	Proxy          string `json:"proxy"`
	APIEndpoint    string `json:"apiEndpoint"`
}

// LogEntry 日志条目结构
type LogEntry struct {
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Domain    string `json:"domain"`
	Message   string `json:"message"`
}

// IPInfo IP 信息结构
type IPInfo struct {
	PublicIP   string `json:"publicIP"`   // IPv4 地址（主要显示）
	IPv4       string `json:"ipv4"`       // IPv4 地址
	IPv6       string `json:"ipv6"`       // IPv6 地址
	Location   string `json:"location"`
	ISP        string `json:"isp"`
}
