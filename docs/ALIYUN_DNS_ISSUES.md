# 阿里云DNS API 踩坑记录

## 问题 1：DomainRecordDuplicate 错误

### 问题描述
调用 `UpdateDomainRecord` API 时返回错误：
```json
{
  "Code": "DomainRecordDuplicate",
  "Message": "The DNS record already exists."
}
```

### 根本原因
当尝试更新的 DNS 记录值（IP 地址）与现有记录**完全相同时**，阿里云 API 会返回此错误，认为这是重复操作。

**场景还原：**
- 当前公网IP：`116.230.251.7`
- DNS 记录中的 IP：`116.230.251.7`
- 由于 IP 没有变化，直接调用 `UpdateDomainRecord` 会导致错误

### 解决方案

#### ✅ 优化策略：IP 变化检测

在调用更新 API 之前，先查询当前的 DNS 记录值，如果 IP 没有变化则跳过更新：

```go
func (p *AliyunProvider) UpdateRecord(domain, ip string) error {
    // 1. 查询现有的解析记录 ID 和当前 IP
    recordID, currentIP, err := p.describeDomainRecord(domainName, rr)
    
    // 2. 检查 IP 是否有变化
    if recordID != "" && currentIP == ip {
        fmt.Printf("✓ IP 无变化，跳过更新：%s -> %s\n", domain, ip)
        return nil // 直接返回成功
    }
    
    // 3. IP 有变化，执行更新
    return p.updateDomainRecord(recordID, rr, ip)
}
```

#### 修改点

1. **describeDomainRecord 方法签名变更**
   ```go
   // 修改前
   func (p *AliyunProvider) describeDomainRecord(domainName, rr string) (string, error)
   
   // 修改后
   func (p *AliyunProvider) describeDomainRecord(domainName, rr string) (string, string, error)
   // 返回值：recordId, currentIP, error
   ```

2. **增加 IP 比较逻辑**
   - 从查询结果中获取当前 IP 值
   - 与目标 IP 进行比较
   - 相同则跳过，不同才更新

### 官方文档参考

- **错误码说明**：https://help.aliyun.com/zh/dns/api-alidns-2015-01-09-errorcodes
  - `DomainRecordDuplicate`: 解析记录已存在
  
- **API 文档**：https://help.aliyun.com/zh/dns/api-alidns-2015-01-09-updatedomainrecord

### 最佳实践

1. **避免不必要的 API 调用**
   - 先查询后比较，减少 API 调用次数
   - 降低触发速率限制的风险

2. **优雅处理重复操作**
   - 将"无变化"视为成功，而不是错误
   - 记录日志便于调试

3. **参考其他开发者的实现**
   - https://gitee.com/zhangruhong/aliyunddns
   - https://m.blog.csdn.net/wandersky0822/article/details/112724046

---

## 问题 2：API 名称错误

### 问题描述
最初代码中使用了错误的 API 名称 `ModifyDomainRecord`。

### 正确做法
根据阿里云官方文档，正确的 API 名称是 **`UpdateDomainRecord`**。

### 参考链接
- 官方文档：https://help.aliyun.com/zh/dns/api-alidns-2015-01-09-updatedomainrecord

---

## 经验总结

### 1. 必须依据官方文档
- 不能依赖猜测或记忆
- 通过 OpenAPI Explorer 实时验证
- 在代码中添加官方文档链接注释

### 2. 智能降级策略
- 优先调用新版 API
- 失败时自动回退到旧版兼容接口
- 确保核心功能可用

### 3. 参数优化
- 仅传递官方文档声明的必填和可选参数
- 避免添加非必要字段
- 使用唯一标识符（如 RecordId）定位资源

### 4. 日志可观测性
- 所有 API 调用生成详细日志
- 记录请求参数、响应状态、错误信息
- 便于问题排查

---

**最后更新时间**：2026-03-09
