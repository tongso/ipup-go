# 阿里云 DNS API 详细参考

## 📋 API 接口清单

### 官方文档
- **主文档**：https://help.aliyun.com/document_detail/29739.html
- **API 列表**：https://next.api.aliyun.com/product/Dns

---

## 🔍 查询类接口

### DescribeDomainRecords

**功能：** 查询域名下的解析记录列表

**请求参数：**
| 参数名 | 类型 | 必需 | 说明 |
|--------|------|------|------|
| Action | String | 是 | 固定值：DescribeDomainRecords |
| DomainName | String | 是 | 主域名（如：jizhangxia.com） |
| RRKeyWord | String | 否 | 子域名关键字（如：ddns） |
| TypeKeyWord | String | 否 | 记录类型（如：A、CNAME、MX） |
| PageNumber | Integer | 否 | 页码，默认 1 |
| PageSize | Integer | 否 | 每页数量，默认 20 |

**返回字段：**
```json
{
  "TotalCount": 1,
  "PageNumber": 1,
  "PageSize": 20,
  "DomainRecords": {
    "Record": [
      {
        "RecordId": "2029489663920487424",
        "RR": "ddns",
        "Type": "A",
        "Value": "116.230.251.7",
        "Status": "ENABLE",
        "TTL": 600,
        "Priority": 10,
        "Line": "default",
        "Weight": 1,
        "DomainName": "jizhangxia.com"
      }
    ]
  }
}
```

**调用示例：**
```bash
curl "https://alidns.aliyuncs.com/?Action=DescribeDomainRecords&DomainName=jizhangxia.com&RRKeyWord=ddns&TypeKeyWord=A&Format=JSON&Version=2015-01-09&AccessKeyId=xxx&Signature=xxx"
```

---

### DescribeDomainRecordInfo

**功能：** 根据 RecordId 查询单条记录详情

**请求参数：**
| 参数名 | 类型 | 必需 | 说明 |
|--------|------|------|------|
| Action | String | 是 | DescribeDomainRecordInfo |
| RecordId | String | 是 | 解析记录 ID |

**返回示例：**
```json
{
  "RecordId": "2029489663920487424",
  "RR": "ddns",
  "Type": "A",
  "Value": "116.230.251.7",
  "Status": "ENABLE",
  "TTL": 600
}
```

---

## ✏️ 修改类接口

### ModifyDomainRecord ✅ **推荐**

**功能：** 修改现有解析记录的属性

**请求参数：**
| 参数名 | 类型 | 必需 | 说明 |
|--------|------|------|------|
| Action | String | 是 | ModifyDomainRecord |
| RecordId | String | 是 | 要更新的记录 ID |
| RR | String | 是 | 子域名（必须与 RecordId 匹配） |
| Type | String | 是 | 记录类型（A、CNAME、MX 等） |
| Value | String | 是 | 新的记录值（IP 地址） |
| TTL | Integer | 否 | 生存时间（秒），默认 600 |
| Priority | Integer | 否 | MX 记录优先级 |

**返回示例：**
```json
{
  "RecordId": "2029489663920487424",
  "RequestId": "AF2A3E5A-A88B-5A12-9370-D14F2FA51DC5"
}
```

**✅ 正确用法：**
```go
params := map[string]string{
    "Action":     "ModifyDomainRecord",
    "RecordId":   "2029489663920487424",
    "RR":         "ddns",
    "Type":       "A",
    "Value":      "116.230.251.100",  // 新 IP
    "TTL":        "600",
}
```

**❌ 常见错误：**
```go
// 错误 1：缺少 RR 参数
params := map[string]string{
    "Action":   "ModifyDomainRecord",
    "RecordId": "2029489663920487424",
    "Value":    "116.230.251.100",
    // ❌ 缺少 RR 和 Type，会报错 "RR is mandatory for this action"
}

// 错误 2：使用 UpdateDomainRecord
params["Action"] = "UpdateDomainRecord"
// ❌ 会返回 "The DNS record already exists"
```

---

### UpdateDomainRecord ❌ **不推荐**

**⚠️ 警告：** 此接口实际行为是**创建**而非更新！

**问题描述：**
- 即使提供了 `RecordId`，仍然要求提供 `RR`、`Type`、`Value`
- 如果记录已存在，会返回 `DomainRecordDuplicate` 错误
- 官方文档描述模糊，容易误导

**错误示例：**
```go
// ❌ 错误：尝试更新记录
params := map[string]string{
    "Action":   "UpdateDomainRecord",
    "RecordId": "2029489663920487424",
    "RR":       "ddns",
    "Type":     "A",
    "Value":    "116.230.251.7",
}

// 返回：
// Code: DomainRecordDuplicate
// Message: The DNS record already exists.
```

**结论：** 始终使用 `ModifyDomainRecord` 进行更新操作！

---

## ➕ 创建类接口

### AddDomainRecord

**功能：** 添加新的 DNS 解析记录

**请求参数：**
| 参数名 | 类型 | 必需 | 说明 |
|--------|------|------|------|
| Action | String | 是 | AddDomainRecord |
| DomainName | String | 是 | 主域名 |
| RR | String | 是 | 子域名 |
| Type | String | 是 | 记录类型 |
| Value | String | 是 | 记录值 |
| TTL | Integer | 否 | 生存时间 |

**返回示例：**
```json
{
  "RecordId": "2029489663920487425",
  "RequestId": "xxx-xxx-xxx"
}
```

---

## 🗑️ 删除类接口

### DeleteDomainRecord

**功能：** 删除指定的 DNS 解析记录

**请求参数：**
| 参数名 | 类型 | 必需 | 说明 |
|--------|------|------|------|
| Action | String | 是 | DeleteDomainRecord |
| RecordId | String | 是 | 要删除的记录 ID |

**返回示例：**
```json
{
  "RecordId": "2029489663920487424",
  "RequestId": "xxx-xxx-xxx"
}
```

---

## 🔐 公共参数

所有 API 请求都需要以下公共参数：

| 参数名 | 类型 | 必需 | 说明 |
|--------|------|------|------|
| AccessKeyId | String | 是 | 访问密钥 ID |
| Format | String | 是 | 返回格式：JSON |
| SignatureMethod | String | 是 | 签名方式：HMAC-SHA1 |
| SignatureNonce | String | 是 | 唯一随机数 |
| SignatureVersion | String | 是 | 签名版本：1.0 |
| Timestamp | String | 是 | UTC 时间戳 |
| Version | String | 是 | API 版本：2015-01-09 |
| Signature | String | 是 | 签名结果 |

---

## 🎯 最佳实践总结

### 1. 正确的 API 选择

| 操作场景 | 应使用的 API | 避免使用的 API |
|---------|-------------|---------------|
| 查询记录列表 | DescribeDomainRecords | - |
| 查询单条记录 | DescribeDomainRecordInfo | - |
| **更新记录值** | ✅ **ModifyDomainRecord** | ❌ UpdateDomainRecord |
| 添加新记录 | AddDomainRecord | - |
| 删除记录 | DeleteDomainRecord | - |

### 2. 参数传递规范

**✅ 推荐做法：**
```go
// 更新操作：提供完整参数
params := map[string]string{
    "Action":     "ModifyDomainRecord",
    "RecordId":   "2029489663920487424",  // 必需
    "RR":         "ddns",                  // 必需
    "Type":       "A",                     // 必需
    "Value":      "116.230.251.100",       // 必需
    "TTL":        "600",                    // 可选
}
```

**❌ 错误做法：**
```go
// 只提供 RecordId 和 Value
params := map[string]string{
    "Action":     "ModifyDomainRecord",
    "RecordId":   "2029489663920487424",
    "Value":      "116.230.251.100",
    // ❌ 缺少 RR 和 Type，会报错
}
```

### 3. 签名生成注意事项

**关键步骤：**
1. 参数按字典序排序
2. 键和值分别 URL 编码
3. 用 `&` 连接构建规范查询字符串
4. 对整个字符串再次 URL 编码（用于待签名字符串）
5. HMAC-SHA1 计算签名

**双重编码陷阱：**
```
原始值：2026-03-05T10:24:15Z
第一次编码：%3A (冒号)
第二次编码：%253A (错误！不应该再次编码)

正确做法：只对最终的 canonicalizedQueryString 整体编码一次
```

### 4. 错误处理策略

```go
response, err := provider.sendRequest(params)
if err != nil {
    // 检查错误类型
    if strings.Contains(err.Error(), "already exists") {
        // 记录已存在 → 应该使用 ModifyDomainRecord
    } else if strings.Contains(err.Error(), "RR is mandatory") {
        // 缺少 RR 参数 → 补充 RR 和 Type
    } else if strings.Contains(err.Error(), "SignatureDoesNotMatch") {
        // 签名错误 → 检查签名算法
    }
}
```

---

## 📖 参考资料

- [阿里云 DNS API 官方文档](https://help.aliyman.com/document_detail/29739.html)
- [API 调试工具](https://api.aliyun.com/troubleshoot)
- [OpenAPI Explorer](https://next.api.aliyun.com/api-tools/sdk/Dns)

---

*最后更新：2026-03-06*
