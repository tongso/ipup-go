package notify

// Notifier 通知器接口
type Notifier interface {
	// Name 返回通知器名称
	Name() string
	
	// Send 发送通知
	Send(title, message string) error
}

// EventType 事件类型
type EventType int

const (
	EventTypeIPChanged EventType = iota + 1 // IP 变更
	EventTypeUpdateSuccess                  // 更新成功
	EventTypeUpdateFailed                   // 更新失败
	EventTypeSystemError                    // 系统错误
)

// Event 事件结构
type Event struct {
	Type    EventType
	Title   string
	Message string
	Data    map[string]interface{}
}

// BaseNotifier 基础通知器
type BaseNotifier struct {
	name string
}

// NewBaseNotifier 创建基础通知器
func NewBaseNotifier(name string) *BaseNotifier {
	return &BaseNotifier{name: name}
}

// Name 返回通知器名称
func (b *BaseNotifier) Name() string {
	return b.name
}
