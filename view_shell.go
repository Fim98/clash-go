package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// view builds the complete application shell and current view.
func (a *app) view(c *ui.Context) {
	t := c.Theme()
	tb := c.TitleBar()

	// Top inset for macOS traffic lights when TitleBarHiddenInset is used
	topInset := tb.Height
	if topInset < 28 {
		topInset = 28
	}

	// Root Horizontal Split: Full-height Left Sidebar + Right Content Area
	ui.Row(c).Fill().Children(func() {
		// ==================== LEFT SIDEBAR ====================
		ui.Column(c).
			Width(210).
			Background(t.Surface).
			Border(1, t.Border).
			Justify(ui.SpaceBetween).
			Children(func() {
				// Top Section: Traffic light spacing + Branding + Navigation
				ui.Column(c).Fill().Gap(10).Padding(topInset+8, 14, 14, 14).Children(func() {
					// App Brand Header
					ui.Row(c).Gap(10).AlignItems(ui.Center).Padding(0, 4).Children(func() {
						ui.Column(c).Gap(2).Children(func() {
							ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
								ui.Text(c, "clash-go").FontSize(18).Bold()
								ui.Badge(c, "Core").Background(ui.RGBA(59, 130, 246, 0.15)).TextColor(t.Accent)
							})
							ui.Text(c, "GPU 原生轻量客户端").FontSize(11).TextColor(t.TextMuted)
						})
					})

					ui.Divider(c)

					// Navigation Sidebar Tree
					ui.Sidebar(c, &a.currentTab, func() {
						ui.SidebarItem(c, "dashboard", nil, "仪表盘  Dashboard")
						ui.SidebarItem(c, "proxies", nil, "代理节点  Proxies")
						ui.SidebarItem(c, "profiles", nil, "订阅配置  Profiles")
						ui.SidebarItem(c, "connections", nil, "连接监控  Connections")
						ui.SidebarItem(c, "rules", nil, "分流规则  Rules")
						ui.SidebarItem(c, "logs", nil, "运行日志  Logs")
						ui.SidebarItem(c, "settings", nil, "偏好设置  Settings")
					}).Grow(1)
				})

				// Bottom Section: Active Profile & Core Status Indicator
				ui.Column(c).
					Gap(8).
					Padding(14).
					Background(ui.RGBA(128, 128, 128, 0.05)).
					Border(1, t.Border).
					Children(func() {
						// Active Profile Info
						ui.Column(c).Gap(2).Children(func() {
							ui.Text(c, "当前订阅").FontSize(10).TextColor(t.TextMuted)
							ui.Text(c, "主力高速专线").FontSize(12).Bold().SingleLine()
						})

						ui.Divider(c)

						// Core Status Pill
						ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
							ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
								ui.Text(c, "●").TextColor(a.stateColor(t)).FontSize(11)
								ui.Text(c, a.stateLabel()).FontSize(11).Bold()
							})
							ui.Text(c, "端口 7890").FontSize(10).TextColor(t.TextMuted)
						})
					})
			})

		// ==================== RIGHT CONTENT AREA ====================
		ui.Column(c).Grow(1).Fill().Children(func() {
			// Top Header Bar
			ui.Row(c).
				Height(topInset+24).
				Padding(topInset-10, 24, 8, 24).
				Background(t.Surface).
				Border(1, t.Border).
				Justify(ui.SpaceBetween).
				AlignItems(ui.Center).
				DragWindow().
				Children(func() {
					// Page Title
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

					// Right Header Controls: Mode Selector & Core Toggle Button
					ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
						ui.Segmented(c, &a.outboundMode, "规则", "全局", "直连")

						if a.status.Running {
							if ui.Button(c, "⏸ 停止核心").Clicked() {
								a.stop()
							}
						} else {
							if ui.PrimaryButton(c, "▶ 启动核心").Disabled(a.busy).Clicked() {
								a.start()
							}
						}

						if ui.Button(c, "⟳ 重载").Clicked() {
							// Trigger reload
						}
					})
				})

			// Main Scrollable Viewport (NO duplicate titles!)
			ui.Scroll(c).Grow(1).Padding(24).Children(func() {
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

			// Bottom Minimal Status Bar
			ui.Row(c).
				Height(28).
				Padding(2, 20).
				Background(t.Surface).
				Border(1, t.Border).
				Justify(ui.SpaceBetween).
				AlignItems(ui.Center).
				Children(func() {
					// Left: Status dot and uptime
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "●").TextColor(a.stateColor(t)).FontSize(11)
						statusText := "核心未运行"
						if a.status.Running {
							statusText = fmt.Sprintf("Mihomo 核心运行中 • %s", uptime(a.status.Since))
						}
						ui.Text(c, statusText).FontSize(11).TextColor(t.TextMuted)
					})

					// Right: Traffic, Memory, Connections
					ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
						upRate := rate(a.status.Up)
						downRate := rate(a.status.Down)
						if !a.status.Running {
							upRate = "0 B/s"
							downRate = "0 B/s"
						}
						ui.Row(c).Gap(6).Children(func() {
							ui.Text(c, "↑").TextColor(t.Warning).FontSize(11).Bold()
							ui.Text(c, upRate).FontSize(11).Bold()
							ui.Text(c, "↓").TextColor(t.Accent).FontSize(11).Bold()
							ui.Text(c, downRate).FontSize(11).Bold()
						})

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
