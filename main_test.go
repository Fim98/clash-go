package main

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"clash-go/internal/core"
)

// TestView tests the layout, tab switching, and core controls headlessly with ui.NewTester.
func TestView(t *testing.T) {
	mgr := core.New(t.TempDir())
	a := newApp(mgr)
	tt := ui.NewTester(a.view, 1060, 700)

	// 1. Initial Frame on Dashboard
	if !tt.HasText("核心控制") {
		t.Errorf("texts %q without sidebar navigation", tt.Texts())
	}
	if !tt.HasText("实时流量监控") {
		t.Errorf("texts %q without dashboard elements", tt.Texts())
	}
	if !tt.HasText("启动核心") {
		t.Errorf("texts %q without start core button", tt.Texts())
	}

	// 2. Tab Navigation: Proxies
	a.currentTab = "proxies"
	tt.Frame()
	if !tt.HasText("代理节点 (Proxies)") {
		t.Errorf("texts %q missing proxies title", tt.Texts())
	}
	if !tt.HasText("🇭🇰 HK-BGP-01") {
		t.Errorf("texts %q missing proxy node card", tt.Texts())
	}

	// 3. Tab Navigation: Profiles
	a.currentTab = "profiles"
	tt.Frame()
	if !tt.HasText("订阅配置 (Profiles)") {
		t.Errorf("texts %q missing profiles title", tt.Texts())
	}
	if !tt.HasText("+ 添加订阅") {
		t.Errorf("texts %q missing add profile button", tt.Texts())
	}

	// 4. Tab Navigation: Connections
	a.currentTab = "connections"
	tt.Frame()
	if !tt.HasText("连接监控 (Connections)") {
		t.Errorf("texts %q missing connections title", tt.Texts())
	}
	if !tt.HasText("目标地址 / 端口") {
		t.Errorf("texts %q missing connections table header", tt.Texts())
	}

	// 5. Tab Navigation: Rules
	a.currentTab = "rules"
	tt.Frame()
	if !tt.HasText("分流规则 (Rules)") {
		t.Errorf("texts %q missing rules title", tt.Texts())
	}
	if !tt.HasText("规则集 Providers 概况") {
		t.Errorf("texts %q missing rule providers card", tt.Texts())
	}

	// 6. Tab Navigation: Logs
	a.currentTab = "logs"
	tt.Frame()
	if !tt.HasText("运行日志 (Logs)") {
		t.Errorf("texts %q missing logs title", tt.Texts())
	}

	// 7. Tab Navigation: Settings
	a.currentTab = "settings"
	tt.Frame()
	if !tt.HasText("偏好设置 (Settings)") {
		t.Errorf("texts %q missing settings title", tt.Texts())
	}
	if !tt.HasText("混合代理端口 (Mixed Port)") {
		t.Errorf("texts %q missing mixed port setting", tt.Texts())
	}

	// 8. Core Status Update
	a.status = core.Status{Running: true, Mode: core.ModeRule, Connections: 4, Up: 1024, Down: 2048}
	tt.Frame()
	if !tt.HasText("运行中") {
		t.Errorf("texts %q without running status", tt.Texts())
	}
	if !tt.HasText("停止核心") {
		t.Errorf("texts %q without stop button", tt.Texts())
	}
}
