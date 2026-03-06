package provider

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AliyunProvider 阿里云 DNS 服务提供商
type AliyunProvider struct {
	*BaseProvider
	accessKeyID     string
	accessKeySecret string
}

// NewAliyunProvider 创建阿里云 DNS 提供商实例
func NewAliyunProvider(accessKeyID, accessKeySecret string) *AliyunProvider {
	return &AliyunProvider{
		BaseProvider:    NewBaseProvider("Aliyun"),
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
	}
}

// UpdateRecord 更新 DNS 记录（A 记录）
func (p *AliyunProvider) UpdateRecord(domain, ip string) error {
	// 分离主域名和子域名
	domainName, rr := p.parseDomain(domain)
	
	// 1. 查询现有的解析记录 ID
	recordID, err := p.describeDomainRecord(domainName, rr)
	if err != nil {
		return fmt.Errorf("查询解析记录失败：%w", err)
	}
	
	// 2. 根据记录是否存在进行更新或创建
	var updateErr error
	if recordID != "" {
		// 记录存在，更新
		updateErr = p.updateDomainRecord(recordID, rr, ip)
		if updateErr != nil {
			return fmt.Errorf("更新 DNS 记录失败（RecordID: %s）：%w", recordID, updateErr)
		}
		fmt.Printf("[Aliyun] 成功更新 DNS 记录：%s -> %s (RecordID: %s)\n", domain, ip, recordID)
	} else {
		// 记录不存在，创建
		updateErr = p.addDomainRecord(domainName, rr, ip)
		if updateErr != nil {
			return fmt.Errorf("创建 DNS 记录失败：%w", updateErr)
		}
		fmt.Printf("[Aliyun] 成功创建 DNS 记录：%s -> %s\n", domain, ip)
	}
	
	return nil
}

// GetRecord 获取当前 DNS 记录
func (p *AliyunProvider) GetRecord(domain string) (string, error) {
	domainName, rr := p.parseDomain(domain)
	
	recordID, err := p.describeDomainRecord(domainName, rr)
	if err != nil {
		return "", err
	}
	
	if recordID == "" {
		return "", fmt.Errorf("未找到解析记录")
	}
	
	// 获取记录详情
	params := map[string]string{
		"Action":    "DescribeDomainRecordInfo",
		"RecordId":  recordID,
		"Timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	
	response, err := p.sendRequest(params)
	if err != nil {
		return "", err
	}
	
	value, ok := response["Value"].(string)
	if !ok {
		return "", fmt.Errorf("无法获取解析值")
	}
	
	return value, nil
}

// parseDomain 解析域名，返回主域名和 RR 前缀
func (p *AliyunProvider) parseDomain(fullDomain string) (domainName, rr string) {
	parts := strings.Split(fullDomain, ".")
	if len(parts) <= 2 {
		return fullDomain, "@"
	}
	
	// 取最后两部分作为主域名
	domainName = parts[len(parts)-2] + "." + parts[len(parts)-1]
	
	// 前面的部分作为 RR
	rr = strings.Join(parts[:len(parts)-2], ".")
	if rr == "" {
		rr = "@"
	}
	
	return domainName, rr
}

// describeDomainRecord 查询解析记录 ID
func (p *AliyunProvider) describeDomainRecord(domainName, rr string) (string, error) {
	params := map[string]string{
		"Action":       "DescribeDomainRecords",
		"DomainName":   domainName,
		"RRKeyWord":    rr,
		"TypeKeyWord":  "A",
		"Timestamp":    time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	
	response, err := p.sendRequest(params)
	if err != nil {
		return "", fmt.Errorf("API 请求失败：%w", err)
	}
	
	// 检查是否有错误响应
	if errMsg, ok := response["Message"].(string); ok {
		return "", fmt.Errorf("API 错误：%s", errMsg)
	}
	
	domainRecords, ok := response["DomainRecords"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("解析响应失败：DomainRecords 字段不存在或类型错误")
	}
	
	recordList, ok := domainRecords["Record"].([]interface{})
	if !ok {
		return "", fmt.Errorf("解析响应失败：Record 字段类型错误")
	}
	
	fmt.Printf("[Aliyun] 查询到 %d 条 DNS 记录\n", len(recordList))
	
	if len(recordList) == 0 {
		return "", nil // 没有找到记录，返回空字符串
	}
	
	// 找到精确匹配的记录（包括 RR 和 Type）
	for _, record := range recordList {
		r, ok := record.(map[string]interface{})
		if !ok {
			continue
		}
		
		rRR, rrOk := r["RR"].(string)
		rType, typeOk := r["Type"].(string)
		
		if rrOk && typeOk && rRR == rr && rType == "A" {
			if recordID, ok := r["RecordId"].(string); ok {
				fmt.Printf("[Aliyun] ✓ 找到匹配的 A 记录：%s, RecordID: %s, Type: %s, Value: %s\n", 
					rr, recordID, rType, r["Value"])
				return recordID, nil
			}
		} else {
			fmt.Printf("[Aliyun] ✗ 跳过记录 - RR: %v, Type: %v (需要 RR=%s, Type=A)\n", rRR, rType, rr)
		}
	}
	
	fmt.Printf("[Aliyun] 未找到匹配的 A 记录：%s\n", rr)
	return "", nil // 没有找到匹配的记录
}

// addDomainRecord 添加解析记录
func (p *AliyunProvider) addDomainRecord(domainName, rr, ip string) error {
	params := map[string]string{
		"Action":    "AddDomainRecord",
		"DomainName": domainName,
		"RR":        rr,
		"Type":      "A",
		"Value":     ip,
		"TTL":       "600",
		"Timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	
	_, err := p.sendRequest(params)
	return err
}

// updateDomainRecord 更新解析记录
func (p *AliyunProvider) updateDomainRecord(recordID, rr, ip string) error {
	// 使用 ModifyDomainRecord API（正确的更新接口）
	params := map[string]string{
		"Action":     "ModifyDomainRecord",
		"RecordId":   recordID,
		"RR":         rr,
		"Type":       "A",
		"Value":      ip,
		"TTL":        "600",
		"Timestamp":  time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	
	fmt.Printf("[Aliyun] 准备更新 DNS 记录 - RecordID: %s, RR: %s, Type: A, IP: %s\n", recordID, rr, ip)
	
	response, err := p.sendRequest(params)
	if err != nil {
		return err
	}
	
	// 检查响应中是否有 RecordId
	if respRecordID, ok := response["RecordId"].(string); ok {
		fmt.Printf("[Aliyun] DNS 记录更新成功 - RecordID: %s\n", respRecordID)
	}
	
	return nil
}

// sendRequest 发送阿里云 API 请求
func (p *AliyunProvider) sendRequest(params map[string]string) (map[string]interface{}, error) {
	// 添加公共参数
	params["AccessKeyId"] = p.accessKeyID
	params["Format"] = "JSON"
	params["SignatureMethod"] = "HMAC-SHA1"
	params["SignatureVersion"] = "1.0"
	params["SignatureNonce"] = strconv.FormatInt(time.Now().UnixNano(), 10)
	params["Version"] = "2015-01-09"
	
	// 生成签名
	signature := p.generateSignature(params)
	params["Signature"] = signature
	
	// 构建请求 URL
	baseURL := "https://alidns.aliyuncs.com/"
	queryParams := url.Values{}
	for k, v := range params {
		queryParams.Set(k, v)
	}
	
	reqURL := baseURL + "?" + queryParams.Encode()
	
	// 调试输出（隐藏敏感信息）
	debugParams := make(map[string]string)
	for k, v := range params {
		if k == "AccessKeyId" || k == "Signature" {
			debugParams[k] = "***HIDDEN***"
		} else {
			debugParams[k] = v
		}
	}
	fmt.Printf("[Aliyun] API 请求：%s\n", debugParams["Action"])
	fmt.Printf("[Aliyun] 请求参数：%+v\n", debugParams)
	fmt.Printf("[Aliyun] 请求 URL 长度：%d bytes\n", len(reqURL))
	
	// 发送 HTTP GET 请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败：%w", err)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%w", err)
	}
	
	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败：%w", err)
	}
	
	// 调试输出响应
	fmt.Printf("[Aliyun] API 响应状态码：%d\n", resp.StatusCode)
	fmt.Printf("[Aliyun] API 响应原始数据：%s\n", string(body))
	
	// 检查错误
	if errMsg, ok := result["Message"].(string); ok {
		fmt.Printf("[Aliyun] API 错误消息：%s\n", errMsg)
		return nil, fmt.Errorf("API 错误：%s", errMsg)
	}
	
	return result, nil
}

// generateSignature 生成阿里云 API 签名
func (p *AliyunProvider) generateSignature(params map[string]string) string {
	// 1. 排序参数（按字典序）
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// 2. 构建规范查询字符串（对键和值分别进行 URL 编码）
	var canonicalizedQueryString string
	for i, k := range keys {
		if i > 0 {
			canonicalizedQueryString += "&"
		}
		// 对键和值都进行 URL 编码
		canonicalizedQueryString += url.QueryEscape(k) + "=" + url.QueryEscape(params[k])
	}
	
	// 3. 构建待签名字符串
	// 注意：canonicalizedQueryString 已经是编码后的形式，直接使用
	stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString)
	
	fmt.Printf("[Aliyun] 规范查询字符串：%s\n", canonicalizedQueryString)
	fmt.Printf("[Aliyun] 待签名字符串：%s\n", stringToSign)
	
	// 4. 计算签名（HMAC-SHA1）
	h := hmac.New(sha1.New, []byte(p.accessKeySecret+"&"))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	return signature
}
