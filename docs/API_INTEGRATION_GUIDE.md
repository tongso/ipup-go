# 第三方 API 集成指南

## 📋 概述

本项目集成了多家 DNS 服务商的 API，实现动态 DNS 解析功能。本文档详细说明各 API 的集成规范和注意事项。

---

## 🔶 阿里云 DNS API

### API 版本
```
Version: 2015-01-09
BaseURL: https://alidns.aliyuncs.com/
```

### 签名算法

#### 1. 签名生成步骤
```go
// 步骤 1：参数按字典序排序
keys := sort.Strings(params)

// 步骤 2：构建规范查询字符串（键值分别 URL 编码）
canonicalizedQueryString := ""
for k, v := range params {
    canonicalizedQueryString += url.QueryEscape(k) + "=" + url.QueryEscape(v) + "&"
}

// 步骤 3：构建待签名字符串（关键：对整个字符串再次 URL 编码！）
stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString)

// 步骤 4：HMAC-SHA1 签名
signature = base64.StdEncoding.EncodeToString(
    hmac_sha1(accessKeySecret + "&", stringToSign)
)
```

#### 2. ⚠️ 关键注意事项

**双重编码问题：**
```
❌ 错误：Timestamp=2026-03-05T10:24:15Z
         ↓ url.QueryEscape() (第一次)
   Timestamp%3D2026-03-05T10%3A24%3A15Z
         ↓ url.QueryEscape() (第二次 - 错误！)
   Timestamp%253D2026-03-05T10%253A24%253A15Z  ← 签名验证失败

✅ 正确：Timestamp=2026-03-05T10:24:15Z
         ↓ url.QueryEscape() (第一次)
   Timestamp%3D2026-03-05T10%3A24%3A15Z
         ↓ 直接使用（不再次编码）
   GET&%2F&Timestamp%3D2026-03-05T10%3A24%3A15Z  ← 签名验证通过
```

### API 接口详解

#### 1. DescribeDomainRecords - 查询解析记录列表

**用途：** 查询域名下的所有解析记录

**请求参数：**
```go
params := map[string]string{
    "Action":        "DescribeDomainRecords",
    "DomainName":    "jizhangxia.com",      // 主域名（必需）
    "RRKeyWord":     "ddns",                // 子域名关键字（可选）
    "TypeKeyWord":   "A",                   // 记录类型（可选）
    "PageNumber":    "1",
    "PageSize":      "20",
    // ... 公共参数
}
```

**返回示例：**
```json
{
  "TotalCount": 1,
  "DomainRecords": {
    "Record": [{
      "RecordId": "2029489663920487424",
      "RR": "ddns",
      "Type": "A",
      "Value": "116.230.251.7",
      "Status": "ENABLE",
      "TTL": 600,
      "DomainName": "jizhangxia.com"
    }]
  }
}
```

**代码位置：** `internal/provider/aliyun.go:112-157`

---

#### 2. ModifyDomainRecord - 修改解析记录 ✅ **推荐使用**

**用途：** 更新现有解析记录的值（IP、TTL 等）

**请求参数：**
```go
params := map[string]string{
    "Action":      "ModifyDomainRecord",    // ✅ 正确的更新接口
    "RecordId":    "2029489663920487424",   // 必需：要更新的记录 ID
    "RR":          "ddns",                  // 必需：子域名
    "Type":        "A",                     // 必需：记录类型
    "Value":       "116.230.251.7",         // 必需：新的 IP 地址
    "TTL":         "600",                   // 可选
    // ... 公共参数
}
```

**返回示例：**
```json
{
  "RecordId": "2029489663920487424",
  "RequestId": "AF2A3E5A-A88B-5A12-9370-D14F2FA51DC5"
}
```

**⚠️ 重要说明：**
- 必须提供 `RecordId`、`RR`、`Type`、`Value` 四个参数
- 即使有 RecordId，RR 和 Type 也是必需的
- 只传递需要修改的字段（Value、TTL）

**代码位置：** `internal/provider/aliyun.go:185-210`

---

#### 3. UpdateDomainRecord - ❌ 不推荐使用

**问题：** 此接口实际是**创建操作**，不是更新操作！

**错误示例：**
```go
// ❌ 错误：会返回 "The DNS record already exists"
params := map[string]string{
    "Action":   "UpdateDomainRecord",
    "RecordId": "2029489663920487424",
    "RR":       "ddns",
    "Type":     "A",
    "Value":    "116.230.251.7",
}

// 返回错误：
// Code: DomainRecordDuplicate
// Message: The DNS record already exists.
```

**结论：** 不要使用此接口进行更新操作！

---

#### 4. AddDomainRecord - 添加解析记录

**用途：** 创建新的 DNS 记录

**请求参数：**
```go
params := map[string]string{
    "Action":     "AddDomainRecord",
    "DomainName": "jizhangxia.com",
    "RR":         "ddns",
    "Type":       "A",
    "Value":      "116.230.251.7",
    "TTL":        "600",
    // ... 公共参数
}
```

**返回示例：**
```json
{
  "RecordId": "2029489663920487425",
  "RequestId": "xxx-xxx-xxx"
}
```

---

### 完整调用流程

```go
// 1. 查询现有记录
recordID, err := provider.describeDomainRecord("jizhangxia.com", "ddns")

if recordID != "" {
    // 2a. 记录存在，更新
    err = provider.updateDomainRecord(recordID, "ddns", newIP)
} else {
    // 2b. 记录不存在，创建
    err = provider.addDomainRecord("jizhangxia.com", "ddns", newIP)
}
```

**代码位置：** `internal/provider/aliyun.go:35-64`

---

### 错误处理

#### 常见错误码

| 错误码 | 说明 | 解决方案 |
|--------|------|---------|
| `DomainRecordDuplicate` | 记录已存在 | 使用 ModifyDomainRecord 而非 AddDomainRecord |
| `MissingParameter` | 缺少必需参数 | 检查 RR、Type 等是否提供 |
| `InvalidAccessKeyId.NotFound` | AccessKey 无效 | 检查密钥配置 |
| `SignatureDoesNotMatch` | 签名不匹配 | 检查签名算法（双重编码问题） |
| `DomainNameNotFound` | 域名不存在 | 检查域名是否正确 |

#### 错误日志示例
```
[Aliyun] API 响应状态码：400
[Aliyun] API 响应原始数据：{"Code":"DomainRecordDuplicate","Message":"The DNS record already exists."}
[Aliyun] API 错误消息：The DNS record already exists.
```

---

### 调试技巧

#### 1. 启用详细日志
```go
fmt.Printf("[Aliyun] 规范查询字符串：%s\n", canonicalizedQueryString)
fmt.Printf("[Aliyun] 待签名字符串：%s\n", stringToSign)
fmt.Printf("[Aliyun] 请求参数：%+v\n", debugParams)
fmt.Printf("[Aliyun] API 响应原始数据：%s\n", string(body))
```

#### 2. 验证签名
比较本地计算的 `stringToSign` 与阿里云服务器计算的字符串是否一致。

#### 3. 检查参数
确保所有必需参数都已提供，特别是 `RecordId`、`RR`、`Type`。

---

## 🌐 其他 DNS 服务商

### 腾讯云 DNSPod
*待补充*

### Cloudflare
*待补充*

### GoDaddy
*待补充*

---

## 📚 参考资料

- [阿里云 DNS API 文档](https://help.aliyun.com/document_detail/29739.html)
- [阿里云 API 调试工具](https://api.aliyun.com/troubleshoot)
- [HMAC-SHA1 签名算法](https://tools.ietf.org/html/rfc2104)

---

*最后更新：2026-03-06*
