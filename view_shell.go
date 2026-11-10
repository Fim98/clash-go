package main

import (
	"github.com/egoist/mygo/ui"
)

// view builds the complete application shell and current view.
func (a *app) view(c *ui.Context) {
	t := c.Theme()

	// Root Horizontal Split: Left Sidebar + Right Content Area
	ui.Row(c).Fill().Children(func() {
		// ==================== LEFT SIDEBAR ====================
		ui.Column(c).
			Width(210).
			Background(t.Surface).
			Border(1, t.Border).
			Justify(ui.SpaceBetween).
			Children(func() {
				// Top Section: App Branding & Navigation Items
				ui.Column(c).Fill().Gap(8).Padding(16, 12).Children(func() {
					// App Brand Header
					ui.Row(c).Gap(10).AlignItems(ui.Center).Padding(4, 8).Children(func() {
						ui.Text(c, "⚡").FontSize(22)
						ui.Column(c).Gap(2).Children(func() {
							ui.Text(c, "clash-go").FontSize(17).Bold()
							ui.Text(c, "GPU 原生客户端").FontSize(10).TextColor(t.TextMuted)
						})
					})

					ui.Divider(c)

					// Navigation Sidebar Tree
					ui.Sidebar(c, &a.currentTab, func() {
						ui.SidebarItem(c, "dashboard", nil, "📊 仪表盘")
						ui.SidebarItem(c, "proxies", nil, "⚡ 代理节点")
						ui.SidebarItem(c, "profiles", nil, "📑 订阅配置")
						ui.SidebarItem(c, "connections", nil, "🌐 连接监控")
						ui.SidebarItem(c, "rules", nil, "🔀 分流规则")
						ui.SidebarItem(c, "logs", nil, "📜 运行日志")
						ui.SidebarItem(c, "settings", nil, "⚙️ 偏好设置")
					}).Grow(1)
				})

				// Bottom Section: Quick Toggles & Client Info
				ui.Column(c).
					Gap(10).
					Padding(14, 16).
					Background(ui.RGBA(128, 128, 128, 0.05)).
					Border(1, t.Border).
					Children(func() {
						// System Proxy Switch
						ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
							ui.Text(c, "系统代理").FontSize(12).Bold()
							ui.Switch(c, &a.systemProxyOn)
						})

						// TUN Mode Switch
						ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
							ui.Text(c, "TUN 模式").FontSize(12).Bold()
							ui.Switch(c, &a.tunModeOn)
						})

						ui.Divider(c)

						ui.Row(c).Justify(ui.SpaceBetween).Children(func() {
							ui.Text(c, "v0.2.0 • Pure Go").FontSize(10).TextColor(t.TextMuted)
							ui.Text(c, "Zero CGO").FontSize(10).TextColor(t.TextMuted)
						})
					})
			})

		// ==================== RIGHT CONTENT AREA ====================
		ui.Column(c).Grow(1).Fill().Children(func() {
			// Top Header Toolbar
			ui.Row(c).
				Height(52).
				Padding(16, 20).
				Background(t.Surface).
				Border(1, t.Border).
				Justify(ui.SpaceBetween).
				AlignItems(ui.Center).
				DragWindow().
				Children(func() {
					// Page Title
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						tabTitles := map[string]string{
							"dashboard":   "仪表盘 (Dashboard)",
							"proxies":     "代理节点 (Proxies)",
							"profiles":    "订阅配置 (Profiles)",
							"connections": "连接监控 (Connections)",
							"rules":       "分流规则 (Rules)",
							"logs":        "运行日志 (Logs)",
							"settings":    "偏好设置 (Settings)",
						}
						title := tabTitles[a.currentTab]
						if title == "" {
							title = "clash-go"
						}
						ui.Text(c, title).FontSize(16).Bold()
					})

					// Header Actions: Outbound Mode & Core Reload
					ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
						ui.Segmented(c, &a.outboundMode, "规则", "全局", "直连")
						if ui.Button(c, "⟳ 重载核心").Clicked() {
							// Trigger core reload
						}
					})
				})

			// Main Scrollable Viewport
			ui.Scroll(c).Grow(1).Padding(20).Children(func() {
				switch a.currentTab {
				case "proxies":
					a.viewProxies(c)
				case "profiles":
					a.viewProfiles(c)
				case "connections":
					a.viewConnections(c)
				case "rules":
					a.viewRules(c)
				case "logs":
					a.viewLogs(c)
				case "settings":
					a.viewSettings(c)
				default:
					a.viewDashboard(c)
				}
			})

			// Bottom Status Bar
			ui.Row(c).
				Height(30).
				Padding(4, 16).
				Background(t.Surface).
				Border(1, t.Border).
				Justify(ui.SpaceBetween).
				AlignItems(ui.Center).
				Children(func() {
					// Left: Core status light & toggle
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "●").TextColor(a.stateColor(t)).FontSize(12)
						ui.Text(c, a.stateLabel()).FontSize(11).TextColor(t.TextMuted)
						if a.status.Running {
							if ui.Button(c, "停止核心").Clicked() {
								a.stop()
							}
						} else {
							if ui.PrimaryButton(c, "启动核心").Disabled(a.busy).Clicked() {
								a.start()
							}
						}
					})

					// Right: Traffic, Memory, Latency
					ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
						upRate := rate(a.status.Up)
						downRate := rate(a.status.Down)
						if !a.status.Running {
							upRate = "0 B/s"
							downRate = "0 B/s"
						}
						ui.Textf(c, "↑ %s   ↓ %s", upRate, downRate).FontSize(11).Bold()

						memStr := bytes(int64(a.status.Memory))
						if !a.status.Running {
							memStr = "0 B"
						}
						ui.Textf(c, "内存: %s", memStr).FontSize(11).TextColor(t.TextMuted)
						ui.Textf(c, "连接: %d", a.status.Connections).FontSize(11).TextColor(t.TextMuted)
						ui.Text(c, "延迟: 42 ms").FontSize(11).TextColor(t.Success)
					})
				})
		})
	})
}
