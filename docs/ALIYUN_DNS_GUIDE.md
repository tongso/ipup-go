# 阿里云 DNS 配置指南

## 概述

本项目支持使用阿里云 DNS 服务进行动态域名解析（DDNS）。本文档将指导您完成配置过程。

## 前置条件

1. 拥有阿里云账号
2. 拥有一个已在阿里云注册或托管的域名
3. 已获取阿里云 AccessKey 凭证

## 获取 AccessKey

### 步骤 1: 登录阿里云控制台

访问 [阿里云控制台](https://ram.console.aliyun.com/)

### 步骤 2: 创建 RAM 用户（推荐）

为了安全起见，建议创建专用的 RAM 用户而不是直接使用主账号：

1. 访问 RAM 访问控制控制台
2. 点击"用户" > "创建用户"
3. 输入用户名（如：ddns-updater）
4. 选择"编程访问"作为访问方式
5. 点击"确定"

### 步骤 3: 获取 AccessKey

1. 创建成功后，系统会显示 AccessKey ID 和 AccessKey Secret
2. **重要**：请立即保存 AccessKey Secret，关闭页面后将无法再次查看

### 步骤 4: 授权权限

为 RAM 用户授予管理云解析 DNS 的权限：

1. 在用户列表中找到刚创建的用户
2. 点击"添加权限"
3. 搜索并选择"AliyunDNSFullAccess"（云解析 DNS 完全管理权限）
4. 点击"确定"

## 在系统中配置

### 方法一：通过 Web 界面

1. 打开应用，进入"域名管理"标签页
2. 点击"➕ 添加域名"
3. 填写以下信息：
   - **域名**: 您的完整域名（如：www.example.com 或 example.com）
   - **服务提供商**: 选择"🔶 阿里云"
   - **AccessKey ID**: 填入刚才获取的 AccessKey ID
   - **AccessKey Secret**: 填入 AccessKey Secret
   - **更新间隔**: 设置检查间隔（建议不少于 300 秒）
   - **立即启用此域名**: 勾选以立即开始监控
4. 点击"添加"按钮

### 方法二：通过配置文件

（如果支持配置文件方式，在此说明）

## 支持的域名格式

- 根域名：`example.com`
- 子域名：`www.example.com`
- 多级子域名：`home.server.example.com`

系统会自动识别域名结构并正确配置 DNS 记录。

## 测试验证

### 使用测试脚本

```bash
# 设置环境变量
export ALIYUN_ACCESS_KEY_ID="your_access_key_id"
export ALIYUN_ACCESS_KEY_SECRET="your_access_key_secret"
export TEST_DOMAIN="your_domain.com"

# 运行测试脚本
go run scripts/test_aliyun_dns.go
```

### 手动验证

1. 在阿里云 DNS 控制台查看解析记录是否已更新
2. 使用 `nslookup` 或 `dig` 命令验证：
   ```bash
   nslookup your_domain.com
   ```

## 常见问题

### Q1: AccessKey 安全吗？

A: AccessKey 会以加密形式存储在本地数据库中。建议使用 RAM 用户而非主账号，并仅授予最小必要权限。

### Q2: 更新频率应该设多少？

A: 建议设置为 300 秒（5 分钟）或更长。过于频繁的更新可能触发阿里云 API 限流。

### Q3: 支持 IPv6 吗？

A: 当前版本主要支持 IPv4（A 记录）。IPv6（AAAA 记录）支持正在开发中。

### Q4: 如何查看更新日志？

A: 在应用的"日志"标签页可以查看详细的更新记录和错误信息。

## API 限制

阿里云 DNS API 有以下限制：

- 单用户 QPS（每秒查询率）：10
- 建议设置合理的更新间隔，避免触发限流

## 故障排除

### 问题：提示"AccessKey 无效"

**解决方案**：
1. 检查 AccessKey ID 和 Secret 是否正确
2. 确认 RAM 用户已被授予 DNS 管理权限
3. 确认 AccessKey 未被禁用或删除

### 问题：提示"域名不存在"

**解决方案**：
1. 确认域名已在阿里云 DNS 中配置
2. 检查域名拼写是否正确
3. 等待域名传播完成（可能需要几分钟到几小时）

### 问题：更新失败但无明确错误

**解决方案**：
1. 查看应用日志获取详细错误信息
2. 检查网络连接是否正常
3. 确认阿里云账户状态正常（无欠费）

## 相关资源

- [阿里云 DNS API 文档](https://help.aliyun.com/document_detail/29739.html)
- [RAM 用户指南](https://help.aliyun.com/document_detail/93718.html)
- [AccessKey 最佳实践](https://help.aliyun.com/document_detail/102607.html)

## 技术支持

如有问题，请查看应用日志或联系技术支持团队。
