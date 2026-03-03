package provider

// Provider DNS 提供商接口
type Provider interface {
	// Name 返回提供商名称
	Name() string
	
	// UpdateRecord 更新 DNS 记录
	UpdateRecord(domain, token, ip string) error
	
	// GetRecord 获取当前 DNS 记录
	GetRecord(domain, token string) (string, error)
}

// BaseProvider 基础提供商结构（可复用）
type BaseProvider struct {
	name string
}

// NewBaseProvider 创建基础提供商
func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{name: name}
}

// Name 返回提供商名称
func (b *BaseProvider) Name() string {
	return b.name
}
