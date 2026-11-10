package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// view builds the complete application shell and current view.
func (a *app) view(c *ui.Context) {
	t := c.Theme()
	canvasBg := t.Background.Mix(t.Surface, 0.45)
	if t.Dark {
		canvasBg = t.Background
	}

	// Root Horizontal Split: AlignItems(ui.Stretch) is critical in MyGo because
	// ui.Row defaults to AlignItems(ui.Center).
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		// ==================== LEFT SIDEBAR ====================
		ui.Column(c).
			Width(210).
			FillHeight().
			Shrink(0).
			Background(t.Surface).
			Children(func() {
				// 1. Top Traffic Light Inset & Drag Region (macOS window controls sit at top-left)
				tb := c.TitleBar()
				topSpace := float32(38)
				if tb.Height > topSpace {
					topSpace = tb.Height
				}
				ui.Row(c).Height(topSpace).Shrink(0).DragWindow()

				// 2. Native macOS AppKit-style Sidebar with Sections
				ui.Sidebar(c, &a.currentTab, func() {
					ui.SidebarSection(c, "核心控制", nil, func() {
						ui.SidebarItem(c, "dashboard", iconDashboard, "仪表盘")
						ui.SidebarItem(c, "proxies", iconProxies, "代理节点")
						ui.SidebarItem(c, "profiles", iconProfiles, "订阅配置")
					})
					ui.SidebarSection(c, "网络与审计", nil, func() {
						connItem := ui.SidebarItem(c, "connections", iconConnections, "连接监控")
						if a.status.Connections > 0 {
							connItem.Children(func() {
								ui.Badge(c, fmt.Sprintf("%d", a.status.Connections))
							})
						}
						ui.SidebarItem(c, "rules", iconRules, "分流规则")
						ui.SidebarItem(c, "logs", iconLogs, "运行日志")
					})
					ui.SidebarSection(c, "系统", nil, func() {
						ui.SidebarItem(c, "settings", iconSettings, "偏好设置")
					})
				}).Grow(1).FillWidth()

				// 3. Sidebar Bottom: Core Status & Active Profile Footer
				ui.Divider(c)
				ui.Column(c).
					Gap(4).
					Padding(10, 16).
					Shrink(0).
					Children(func() {
						ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
							ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
								ui.Text(c, "●").TextColor(a.stateColor(t)).FontSize(10)
								ui.Text(c, a.stateLabel()).FontSize(11).Bold()
							})
							ui.Textf(c, ":%d", mixedPort).FontSize(11).TextColor(t.TextMuted)
						})
						ui.Text(c, "当前订阅: 主力高速专线").FontSize(10).TextColor(t.TextMuted).SingleLine()
					})
			})

		// Crisp 1px vertical separator (ui.Divider inspects parent Row automatically)
		ui.Divider(c)

		// ==================== RIGHT CONTENT AREA ====================
		ui.Column(c).Grow(1).MinWidth(0).FillHeight().Background(canvasBg).Children(func() {
			// 1. Top Header Toolbar (52px height aligned with macOS unified toolbar)
			ui.Row(c).
				Height(52).
				Shrink(0).
				Padding(0, 20).
				Background(t.Background).
				Justify(ui.SpaceBetween).
				AlignItems(ui.Center).
				DragWindow().
				Children(func() {
					// Left: Page Title & Error Banner if any
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
					ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
						ui.Text(c, title).FontSize(15).Bold()
						if a.err != "" {
							ui.Badge(c, a.err).Background(t.Danger.Alpha(0.14)).TextColor(t.Danger)
						}
					})

					// Right Header Actions: Mode Segmented, Start/Stop Button, Reload Button
					ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
						if ui.Segmented(c, &a.outboundMode, "规则", "全局", "直连").Changed() {
							a.setMode(modes[a.outboundMode].mode)
						}

						if a.status.Running {
							if ui.Button(c, "停止核心").Clicked() {
								a.stop()
							}
						} else {
							if ui.PrimaryButton(c, "启动核心").Disabled(a.busy).Clicked() {
								a.start()
							}
						}

						if ui.Button(c, "重载").Clicked() {
							if a.status.Running {
								a.start()
							}
						}
					})
				})

			// Header bottom border
			ui.Divider(c)

			// 2. Main Viewport
			// Views with internal virtualized tables/logs fill the viewport directly;
			// card-based pages wrap in a vertical Scroll container.
			switch a.currentTab {
			case "connections":
				ui.Column(c).Grow(1).Padding(20).Children(func() {
					a.viewConnections(c)
				})
			case "rules":
				ui.Column(c).Grow(1).Padding(20).Children(func() {
					a.viewRules(c)
				})
			case "logs":
				ui.Column(c).Grow(1).Padding(20).Children(func() {
					a.viewLogs(c)
				})
			default:
				ui.Scroll(c).Grow(1).Padding(20).Children(func() {
					switch a.currentTab {
					case "proxies":
						a.viewProxies(c)
					case "profiles":
						a.viewProfiles(c)
					case "settings":
						a.viewSettings(c)
					default:
						a.viewDashboard(c)
					}
				})
			}

			// 3. Bottom Minimal Status Bar
			ui.Divider(c)
			ui.Row(c).
				Height(28).
				Shrink(0).
				Padding(0, 16).
				Background(t.Background).
				Justify(ui.SpaceBetween).
				AlignItems(ui.Center).
				Children(func() {
					// Left: Status dot and uptime
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "●").TextColor(a.stateColor(t)).FontSize(10)
						statusText := "核心未运行"
						if a.status.Running {
							statusText = fmt.Sprintf("Mihomo %s · 运行 %s", a.status.Version, uptime(a.status.Since))
						}
						ui.Text(c, statusText).FontSize(11).TextColor(t.TextMuted)
					})

					// Right: Live Traffic, Memory, Connections from a.status
					ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
						upRate := rate(a.status.Up)
						downRate := rate(a.status.Down)
						if !a.status.Running {
							upRate = "0 B/s"
							downRate = "0 B/s"
						}
						ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
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
					})
				})
		})
	})
}
