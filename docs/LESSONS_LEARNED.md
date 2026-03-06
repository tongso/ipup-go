# 踩坑记录和经验总结

## 📅 2026-03-06 阿里云 DNS API 集成踩坑记

### 🎯 问题背景

在实现阿里云 DNS 动态更新功能时，遇到了一系列签名验证和 API 调用错误。本文档详细记录了问题排查过程和解决方案。

---

## 🐛 问题一：签名验证失败

### 错误现象
```
[Aliyun] API 响应状态码：400
API 错误：Specified signature is not matched with our calculation.
server string to sign is: GET&%2F&AccessKeyId%3D...%26Timestamp%3D2026-03-05T10%253A24%253A15Z
```

### 根本原因

**URL 双重编码问题！**

错误的签名生成逻辑：
```go
// ❌ 错误代码
canonicalizedQueryString := ""
for k, v := range params {
    canonicalizedQueryString += url.QueryEscape(k) + "=" + url.QueryEscape(v) + "&"
}

// 错误：对已经编码的字符串再次编码！
stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString)
```

**问题链路：**
```
原始时间戳："2026-03-05T10:24:15Z"
    ↓ url.QueryEscape() (第一次)
"2026-03-05T10%3A24:15Z"
    ↓ url.QueryEscape() (第二次 - 错误！)
"2026-03-05T10%253A24:15Z"  ← %3A 被编码成 %253A
```

### 解决方案

```go
// ✅ 正确代码
func (p *AliyunProvider) generateSignature(params map[string]string) string {
    // 1. 排序参数
    keys := sort.Strings(keys)
    
    // 2. 构建规范查询字符串（键值分别编码）
    var canonicalizedQueryString string
    for i, k := range keys {
        if i > 0 {
            canonicalizedQueryString += "&"
        }
        canonicalizedQueryString += url.QueryEscape(k) + "=" + url.QueryEscape(params[k])
    }
    
    // 3. ✅ 关键修复：对整个字符串进行 URL 编码（只此一次！）
    stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString)
    
    // 4. HMAC-SHA1 签名
    h := hmac.New(sha1.New, []byte(p.accessKeySecret+"&"))
    h.Write([]byte(stringToSign))
    signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
    
    return signature
}
```

### 经验教训

✅ **签名生成规则：**
1. 参数键值分别 URL 编码 → 构建规范查询字符串
2. 对整个规范查询字符串 URL 编码 → 用于待签名字符串
3. **不要**对已编码的部分再次编码！

---

## 🐛 问题二：RR is mandatory for this action

### 错误现象
```
[Aliyun] API 响应状态码：400
API 错误：RR is mandatory for this action.
```

### 根本原因

尝试简化参数，移除了 RR 和 Type：
```go
// ❌ 错误代码
params := map[string]string{
    "Action":   "UpdateDomainRecord",
    "RecordId": "2029489663920487424",
    "Value":    "116.230.251.7",
    // ❌ 缺少 RR 和 Type
}
```

### 解决方案

```go
// ✅ 正确代码
params := map[string]string{
    "Action":     "ModifyDomainRecord",
    "RecordId":   "2029489663920487424",
    "RR":         "ddns",              // ✅ 必需
    "Type":       "A",                 // ✅ 必需
    "Value":      "116.230.251.7",     // ✅ 必需
    "TTL":        "600",               // 可选
}
```

### 经验教训

✅ **即使提供了 RecordId，RR 和 Type 仍然是必需的！**

---

## 🐛 问题三：The DNS record already exists

### 错误现象
```
[Aliyun] API 响应状态码：400
Code: DomainRecordDuplicate
Message: The DNS record already exists.
```

### 根本原因

**使用了错误的 API！** `UpdateDomainRecord` 实际上是创建操作，不是更新操作。

```go
// ❌ 错误代码
params := map[string]string{
    "Action":   "UpdateDomainRecord",  // ❌ 这个接口是创建操作！
    "RecordId": "2029489663920487424",
    "RR":       "ddns",
    "Type":     "A",
    "Value":    "116.230.251.7",
}

// 返回：DomainRecordDuplicate 错误
```

### 解决方案

```go
// ✅ 正确代码
params := map[string]string{
    "Action":   "ModifyDomainRecord",  // ✅ 正确的更新接口
    "RecordId": "2029489663920487424",
    "RR":       "ddns",
    "Type":     "A",
    "Value":    "116.230.251.7",
}
```

### 经验教训

✅ **API 命名陷阱：**
- `UpdateDomainRecord` ≠ 更新操作（实际是创建）
- `ModifyDomainRecord` = 修改现有记录 ✅

✅ **永远使用 `ModifyDomainRecord` 进行更新操作！**

---

## 🐛 问题四：ModifyDomainRecord 返回 404

### 错误现象
```
[Aliyun] API 响应状态码：404
API 错误：Specified api is not found
```

### 根本原因

当时签名计算逻辑有 bug（见问题一），导致阿里云无法识别 API。

### 解决方案

修复签名生成逻辑后，404 错误自动消失。

### 经验教训

✅ **404 不一定是 API 不存在，可能是签名错误导致服务器无法解析！**

---

## 📊 完整调试日志示例

### 成功的更新流程
```
[ddns.jizhangxia.com] 开始手动更新 DNS 解析
[ddns.jizhangxia.com] 获取到公网 IP: 116.230.251.7
[ddns.jizhangxia.com] 数据库 IP 已更新：116.230.251.7
[Aliyun] API 请求：DescribeDomainRecords
[Aliyun] 规范查询字符串：AccessKeyId=***&Action=DescribeDomainRecords&DomainName=jizhangxia.com&...
[Aliyun] 待签名字符串：GET&%2F&AccessKeyId%3D***%26Action%3DDescribeDomainRecords%26...
[Aliyun] 请求参数：map[Action:DescribeDomainRecords DomainName:jizhangxia.com ...]
[Aliyun] 请求 URL 长度：340 bytes
[Aliyun] API 响应状态码：200
[Aliyun] API 响应原始数据：{"TotalCount":1,"DomainRecords":{"Record":[...]}}
[Aliyun] 查询到 1 条 DNS 记录
[Aliyun] ✓ 找到匹配的 A 记录：ddns, RecordID: 2029489663920487424, Type: A, Value: 116.230.251.7
[Aliyun] 准备更新 DNS 记录 - RecordID: 2029489663920487424, RR: ddns, Type: A, IP: 116.230.251.7
[Aliyun] 规范查询字符串：AccessKeyId=***&Action=ModifyDomainRecord&...
[Aliyun] 待签名字符串：GET&%2F&AccessKeyId%3D***%26Action%3DModifyDomainRecord%26...
[Aliyun] 请求参数：map[Action:ModifyDomainRecord RecordId:xxx RR:ddns Type:A Value:116.230.251.7 ...]
[Aliyun] 请求 URL 长度：356 bytes
[Aliyun] API 请求：ModifyDomainRecord
[Aliyun] API 响应状态码：200
[Aliyun] API 响应原始数据：{"RecordId":"2029489663920487424","RequestId":"xxx"}
[Aliyun] DNS 记录更新成功 - RecordID: 2029489663920487424
[Aliyun] 成功更新 DNS 记录
[ddns.jizhangxia.com] 成功调用 Aliyun API 更新 DNS 记录：ddns.jizhangxia.com -> 116.230.251.7
```

---

## 🎯 最终解决方案总结

### 1. 正确的签名生成
```go
// 步骤 1：参数排序
keys := sort.Strings(keys)

// 步骤 2：键值分别编码，构建规范查询字符串
canonicalizedQueryString += url.QueryEscape(k) + "=" + url.QueryEscape(v)

// 步骤 3：整体编码（关键！）
stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString)

// 步骤 4：HMAC-SHA1 签名
signature = base64.StdEncoding.EncodeToString(hmac_sha1(secret+"&", stringToSign))
```

### 2. 正确的 API 选择
| 操作 | 正确的 API | 错误的 API |
|------|-----------|-----------|
| 查询记录 | DescribeDomainRecords | - |
| 更新记录 | ✅ ModifyDomainRecord | ❌ UpdateDomainRecord |
| 创建记录 | AddDomainRecord | - |
| 删除记录 | DeleteDomainRecord | - |

### 3. 必需的参数
```go
// 更新操作必须提供
params := map[string]string{
    "Action":     "ModifyDomainRecord",
    "RecordId":   "记录 ID",    // 必需
    "RR":         "子域名",     // 必需（即使有 RecordId）
    "Type":       "A",          // 必需
    "Value":      "新 IP",      // 必需
    "TTL":        "600",        // 可选
}
```

---

## 💡 最佳实践建议

### 1. 调试技巧
- ✅ 添加详细的调试日志输出
- ✅ 打印规范查询字符串和待签名字符串
- ✅ 显示 API 响应原始数据
- ✅ 隐藏敏感信息（AccessKeyId、Signature）

### 2. 错误处理
```go
response, err := provider.sendRequest(params)
if err != nil {
    errMsg := err.Error()
    
    switch {
    case strings.Contains(errMsg, "already exists"):
        // 使用 ModifyDomainRecord 而非 AddDomainRecord
    case strings.Contains(errMsg, "RR is mandatory"):
        // 补充 RR 和 Type 参数
    case strings.Contains(errMsg, "SignatureDoesNotMatch"):
        // 检查签名算法（双重编码问题）
    case strings.Contains(errMsg, "api is not found"):
        // 检查 API 名称或签名
    }
}
```

### 3. 代码组织
- ✅ 将签名生成逻辑封装为独立方法
- ✅ 使用统一的 sendRequest 方法
- ✅ 分离查询和更新逻辑
- ✅ 添加完善的错误处理

---

## 📚 参考资料

- [阿里云 DNS API 官方文档](https://help.aliyun.com/document_detail/29739.html)
- [API 集成指南](./API_INTEGRATION_GUIDE.md)
- [项目开发规范](./PROJECT_SPECIFICATIONS.md)

---

*记录日期：2026-03-06*  
*问题解决耗时：约 2 小时*  
*关键突破：发现 UpdateDomainRecord 实际是创建操作*
