package provider

import "fmt"

// ProviderFactory 提供商工厂函数
type ProviderFactory func(domain, token, accessKeyID, accessKeySecret string) (Provider, error)

// GetProvider 根据服务商名称获取对应的 Provider 实例
func GetProvider(providerName, domain, token, accessKeyID, accessKeySecret string) (Provider, error) {
	switch providerName {
	case "Aliyun":
		if accessKeyID == "" || accessKeySecret == "" {
			return nil, fmt.Errorf("阿里云需要 AccessKey ID 和 AccessKey Secret")
		}
		return NewAliyunProvider(accessKeyID, accessKeySecret), nil
	
	case "Cloudflare":
		if token == "" {
			return nil, fmt.Errorf("Cloudflare 需要 API Token")
		}
		// TODO: 实现 Cloudflare provider
		return nil, fmt.Errorf("Cloudflare provider 尚未实现")
	
	default:
		return nil, fmt.Errorf("不支持的服务商：%s", providerName)
	}
}
