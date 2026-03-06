# 阿里云 DNS 配置快速参考

## 🔑 获取 AccessKey（3 分钟）

1. 访问：https://ram.console.aliyun.com/
2. 创建用户 → 输入用户名 → 勾选"编程访问"
3. 保存显示的 AccessKey ID 和 Secret
4. 添加权限：搜索"AliyunDNSFullAccess" → 确定

## ⚙️ 配置步骤（1 分钟）

### 方式一：Web 界面（推荐）

```
域名管理 → ➕ 添加域名 → 
  域名：www.yourdomain.com
  服务商：🔶 阿里云
  AccessKey ID: LTAI5t...
  AccessKey Secret: xxxxxxxxxxxxxxxx
  更新间隔：300（秒）
  ✓ 立即启用
→ 添加
```

### 方式二：测试脚本验证

```powershell
# PowerShell
$env:ALIYUN_ACCESS_KEY_ID="LTAI5t..."
$env:ALIYUN_ACCESS_KEY_SECRET="xxxxx"
$env:TEST_DOMAIN="www.yourdomain.com"
go run scripts/test_aliyun_dns.go
```

## 📋 常用配置示例

| 场景 | 域名填写 | 说明 |
|------|---------|------|
| 根域名 | `example.com` | 直接解析到主域名 |
| WWW 子域 | `www.example.com` | 最常见的网站地址 |
| 动态 DNS | `home.example.com` | 家庭宽带常用 |
| 多级子域 | `nas.home.example.com` | 复杂场景 |

## ⏱️ 更新间隔建议

| 网络环境 | 推荐间隔 | 说明 |
|---------|---------|------|
| 家庭宽带（动态 IP） | 300-600 秒 | IP 变化频繁 |
| 企业专线（静态 IP） | 3600 秒 + | IP 基本不变 |
| 数据中心 | 1800 秒 + | 较为稳定 |

## 🔍 验证是否生效

### Windows
```cmd
nslookup www.yourdomain.com
```

### macOS / Linux
```bash
dig www.yourdomain.com +short
```

### 在线工具
- [站长工具](http://tool.chinaz.com/dns/)
- [DNS 传播检查](https://dnschecker.org/)

## ❗ 常见错误速查

| 错误提示 | 原因 | 解决方案 |
|---------|------|---------|
| AccessKey 无效 | 密钥错误或未授权 | 检查拼写，确认 RAM 权限 |
| 域名不存在 | 域名未在阿里云托管 | 检查域名拼写或转移 DNS |
| API 调用失败 | 网络问题或欠费 | 检查网络和账户状态 |
| 触发限流 | 更新过于频繁 | 增大更新间隔至 300 秒 + |

## 🛡️ 安全提醒

✅ **推荐做法**：
- 使用 RAM 用户，不用主账号
- 仅授予 DNS 管理权限
- 定期轮换 AccessKey
- 不在公开场合泄露

❌ **禁止行为**：
- 将 AccessKey 上传到代码仓库
- 在公共论坛分享密钥
- 使用他人提供的 AccessKey
- 长期不更换密钥

## 📞 获取帮助

1. **查看日志**：应用内"日志"标签页
2. **官方文档**：docs/ALIYUN_DNS_GUIDE.md
3. **API 文档**：https://help.aliyun.com/document_detail/29739.html
4. **技术支持**：提交工单或联系运维团队

---

💡 **提示**：首次配置后，建议等待 5-10 分钟再验证 DNS 是否生效。
