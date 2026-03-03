package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier Webhook 通知器
type WebhookNotifier struct {
	*BaseNotifier
	webhookURL string
	secret     string
}

// NewWebhookNotifier 创建 Webhook 通知器
func NewWebhookNotifier(webhookURL, secret string) *WebhookNotifier {
	return &WebhookNotifier{
		BaseNotifier: NewBaseNotifier("Webhook"),
		webhookURL:   webhookURL,
		secret:       secret,
	}
}

// Send 发送 Webhook 通知
func (w *WebhookNotifier) Send(title, message string) error {
	payload := map[string]interface{}{
		"title":   title,
		"message": message,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败：%w", err)
	}
	
	req, err := http.NewRequest("POST", w.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败：%w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 如果有密钥，添加签名
	if w.secret != "" {
		// TODO: 添加 HMAC 签名逻辑
		req.Header.Set("X-Signature", w.secret)
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送 Webhook 失败：%w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook 返回错误状态码：%d", resp.StatusCode)
	}
	
	return nil
}
