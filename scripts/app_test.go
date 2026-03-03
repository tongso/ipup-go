package main

import (
	"fmt"
	"testing"
	"time"
)

// TestDatabaseOperations 测试数据库操作
func TestDatabaseOperations(t *testing.T) {
	app := NewApp()
	
	// 初始化数据库（使用内存数据库进行测试）
	if err := app.initDB(); err != nil {
		t.Fatalf("数据库初始化失败：%v", err)
	}
	defer app.closeDB()
	
	fmt.Println("=== 测试域名管理 ===")
	testDomainManagement(app, t)
	
	fmt.Println("\n=== 测试设置管理 ===")
	testSettingsManagement(app, t)
	
	fmt.Println("\n=== 测试日志管理 ===")
	testLogManagement(app, t)
	
	fmt.Println("\n=== 所有测试通过！===")
}

// testDomainManagement 测试域名管理
func testDomainManagement(app *App, t *testing.T) {
	// 1. 添加域名
	domain := DomainConfig{
		Domain:   "test.example.com",
		Provider: "Cloudflare",
		Token:    "test_token_12345",
		Interval: 300,
		Enabled:  true,
	}
	
	if err := app.AddDomain(domain); err != nil {
		t.Errorf("添加域名失败：%v", err)
		return
	}
	fmt.Println("✓ 添加域名成功")
	
	// 2. 查询域名列表
	domains := app.ListDomains()
	if len(domains) != 1 {
		t.Errorf("期望 1 个域名，实际 %d 个", len(domains))
		return
	}
	fmt.Printf("✓ 查询域名列表成功，共 %d 个\n", len(domains))
	
	// 3. 更新域名
	domain.ID = domains[0].ID
	domain.Interval = 600
	if err := app.UpdateDomain(domain); err != nil {
		t.Errorf("更新域名失败：%v", err)
		return
	}
	fmt.Println("✓ 更新域名成功")
	
	// 4. 切换域名状态
	if err := app.ToggleDomain(domain.ID, false); err != nil {
		t.Errorf("切换域名状态失败：%v", err)
		return
	}
	fmt.Println("✓ 切换域名状态成功")
	
	// 5. 获取域名状态
	statuses := app.GetDomainStatus()
	fmt.Printf("✓ 获取域名状态成功，启用中的域名 %d 个\n", len(statuses))
	
	// 6. 删除域名
	if err := app.DeleteDomain(domain.ID); err != nil {
		t.Errorf("删除域名失败：%v", err)
		return
	}
	fmt.Println("✓ 删除域名成功")
}

// testSettingsManagement 测试设置管理
func testSettingsManagement(app *App, t *testing.T) {
	// 1. 加载默认设置
	settings := app.LoadSettings()
	if !settings.AutoStart {
		t.Error("期望 AutoStart 为 true")
	}
	if settings.CheckInterval != 300 {
		t.Errorf("期望 CheckInterval 为 300，实际 %d", settings.CheckInterval)
	}
	fmt.Println("✓ 加载默认设置成功")
	
	// 2. 保存自定义设置
	newSettings := Settings{
		AutoStart:      false,
		CheckInterval:  600,
		RetryCount:     5,
		RetryDelay:     15,
		LogLevel:       "debug",
		NotifySuccess:  true,
		NotifyError:    true,
		Proxy:          "http://proxy.example.com:8080",
		APIEndpoint:    "https://custom.api.com",
	}
	
	if err := app.SaveSettings(newSettings); err != nil {
		t.Errorf("保存设置失败：%v", err)
		return
	}
	fmt.Println("✓ 保存设置成功")
	
	// 3. 验证设置已保存
	loadedSettings := app.LoadSettings()
	if loadedSettings.AutoStart != false {
		t.Error("期望 AutoStart 为 false")
	}
	if loadedSettings.CheckInterval != 600 {
		t.Errorf("期望 CheckInterval 为 600，实际 %d", loadedSettings.CheckInterval)
	}
	fmt.Println("✓ 验证设置正确")
	
	// 4. 重置设置
	if err := app.ResetSettings(); err != nil {
		t.Errorf("重置设置失败：%v", err)
		return
	}
	fmt.Println("✓ 重置设置成功")
	
	// 5. 验证重置后的设置
	defaultSettings := app.LoadSettings()
	if !defaultSettings.AutoStart {
		t.Error("期望重置后 AutoStart 为 true")
	}
	fmt.Println("✓ 验证默认设置正确")
}

// testLogManagement 测试日志管理
func testLogManagement(app *App, t *testing.T) {
	// 1. 添加测试日志
	app.addLog("info", "test.com", "这是一条测试日志")
	app.addLog("error", "test.com", "这是一条错误日志")
	app.addLog("warning", "example.com", "这是一条警告日志")
	app.addLog("success", "test.com", "这是一条成功日志")
	time.Sleep(100 * time.Millisecond) // 等待写入完成
	fmt.Println("✓ 添加测试日志成功")
	
	// 2. 查询所有日志
	logs := app.GetLogs("all", "")
	if len(logs) < 4 {
		t.Errorf("期望至少 4 条日志，实际 %d 条", len(logs))
		return
	}
	fmt.Printf("✓ 查询所有日志成功，共 %d 条\n", len(logs))
	
	// 3. 按级别查询日志
	errorLogs := app.GetLogs("error", "")
	if len(errorLogs) < 1 {
		t.Error("期望至少 1 条错误日志")
	}
	fmt.Printf("✓ 按级别查询成功，错误日志 %d 条\n", len(errorLogs))
	
	// 4. 关键词搜索日志
	searchLogs := app.GetLogs("all", "test.com")
	if len(searchLogs) < 1 {
		t.Error("期望搜索到包含 test.com 的日志")
	}
	fmt.Printf("✓ 关键词搜索成功，找到 %d 条相关日志\n", len(searchLogs))
	
	// 5. 导出日志
	exportContent := app.ExportLogs()
	if exportContent == "" {
		t.Error("导出日志内容为空")
	}
	fmt.Printf("✓ 导出日志成功，内容长度：%d\n", len(exportContent))
	
	// 6. 清空日志
	if err := app.ClearLogs(); err != nil {
		t.Errorf("清空日志失败：%v", err)
		return
	}
	fmt.Println("✓ 清空日志成功")
	
	// 7. 验证日志已清空
	logsAfterClear := app.GetLogs("all", "")
	if len(logsAfterClear) != 0 {
		t.Errorf("期望清空后日志为 0 条，实际 %d 条", len(logsAfterClear))
	}
	fmt.Println("✓ 验证日志已清空")
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	app := NewApp()
	if err := app.initDB(); err != nil {
		t.Fatalf("数据库初始化失败：%v", err)
	}
	defer app.closeDB()
	
	done := make(chan bool, 10)
	
	// 启动 10 个 goroutine 并发访问
	for i := 0; i < 10; i++ {
		go func(id int) {
			// 读操作
			app.ListDomains()
			app.LoadSettings()
			app.GetLogs("all", "")
			
			// 写操作
			domain := DomainConfig{
				Domain:   fmt.Sprintf("test%d.example.com", id),
				Provider: "Cloudflare",
				Token:    "token",
				Interval: 300,
				Enabled:  true,
			}
			app.AddDomain(domain)
			
			done <- true
		}(i)
	}
	
	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
	
	fmt.Println("✓ 并发访问测试通过")
}
