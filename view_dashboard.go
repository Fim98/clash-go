package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
)

func (a *app) viewDashboard(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Gap(14).Children(func() {
		// 1. Realtime Traffic Chart Card
		Card(c, func() {
			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					ui.Text(c, "实时流量监控 (Realtime Traffic)").FontSize(14).Bold()
					ui.Text(c, "近 60 秒代理与直连上下行网络速率").FontSize(11).TextColor(t.TextMuted)
				})

				// Realtime Rate Badges
				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					upRate := rate(a.status.Up)
					downRate := rate(a.status.Down)
					if !a.status.Running {
						upRate = "0 B/s"
						downRate = "0 B/s"
					}
					ui.Row(c).Gap(6).AlignItems(ui.Center).Padding(4, 10).Background(t.Warning.Alpha(0.1)).Radius(6).Children(func() {
						ui.Text(c, "↑").TextColor(t.Warning).Bold().FontSize(12)
						ui.Text(c, upRate).FontSize(12).Bold().TextColor(t.Warning)
						ui.Textf(c, "(累计 %s)", bytes(a.status.UpTotal)).FontSize(10).TextColor(t.TextMuted)
					})

					ui.Row(c).Gap(6).AlignItems(ui.Center).Padding(4, 10).Background(t.Accent.Alpha(0.1)).Radius(6).Children(func() {
						ui.Text(c, "↓").TextColor(t.Accent).Bold().FontSize(12)
						ui.Text(c, downRate).FontSize(12).Bold().TextColor(t.Accent)
						ui.Textf(c, "(累计 %s)", bytes(a.status.DownTotal)).FontSize(10).TextColor(t.TextMuted)
					})
				})
			})

			// Chart Plot
			yMax := chartMax(a.trafficHistoryUp, a.trafficHistoryDown)
			yScale := chart.NewLinear(0, yMax).Nice(4)

			chart.Plot(c, chart.Axes{}, func(area ui.Element) {
				area.Draw(func(p *ui.Painter, r ui.Rect) {
					for _, tick := range yScale.Ticks(4) {
						yPos := chart.Y(r, tick.At)
						p.Line(r.X, yPos, r.X+r.W, yPos, 1, t.Border.Alpha(0.5))
					}

					n := len(a.trafficHistoryUp)
					if n > 1 {
						xs := make([]float32, n)
						ysUp := make([]float32, n)
						ysDown := make([]float32, n)
						for i := 0; i < n; i++ {
							xs[i] = float32(i) / float32(n-1)
							ysUp[i] = yScale.At(float64(a.trafficHistoryUp[i]))
							ysDown[i] = yScale.At(float64(a.trafficHistoryDown[i]))
						}

						upPath := chart.Line(r, xs, ysUp, chart.CurveMonotone)
						p.StrokePath(upPath, 2, t.Warning)

						downPath := chart.Line(r, xs, ysDown, chart.CurveMonotone)
						p.StrokePath(downPath, 2, t.Accent)
					}
				})
			}).Height(105)
		})

		// 2. Control & IP Environment Cards Grid (2 Equal-Height Columns)
		ui.Row(c).Gap(14).AlignItems(ui.Stretch).Children(func() {
			// Left Card: Quick Control
			Card(c, func() {
				ui.Text(c, "快捷网络开关").FontSize(14).Bold()

				// System Proxy Switch
				ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
					ui.Column(c).Gap(2).Children(func() {
						ui.Text(c, "系统代理 (System Proxy)").FontSize(13).Bold()
						ui.Textf(c, "自动配置系统网络代理 127.0.0.1:%d", mixedPort).FontSize(11).TextColor(t.TextMuted)
					})
					ui.Switch(c, &a.systemProxyOn)
				})

				ui.Divider(c)

				// TUN Mode Switch
				ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
					ui.Column(c).Gap(2).Children(func() {
						ui.Text(c, "TUN 虚拟网卡模式").FontSize(13).Bold()
						ui.Text(c, "接管全局流量并绕过应用代理配置").FontSize(11).TextColor(t.TextMuted)
					})
					ui.Switch(c, &a.tunModeOn)
				})

				ui.Divider(c)

				// Outbound Mode Selector
				ui.Column(c).Gap(8).Children(func() {
					ui.Text(c, "分流规则模式").FontSize(13).Bold()
					if ui.Segmented(c, &a.outboundMode, "规则模式", "全局模式", "直连模式").Changed() {
						a.setMode(modes[a.outboundMode].mode)
					}
				})
			}).Grow(1)

			// Right Card: External IP & Environment
			Card(c, func() {
				ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "出口 IP 与网络环境").FontSize(14).Bold()
					ui.Badge(c, "M5").Background(t.Text.Alpha(0.08)).TextColor(t.TextMuted)
				})

				StatRow(c, "出口 IP 地址", "—")
				StatRow(c, "地理位置", "—")
				StatRow(c, "运营商 / ASN", "—")
				StatRow(c, "本地测试延迟", "—")
				ui.Spacer(c)
				ui.Text(c, "探测出口 IP 需要外部服务，随系统集成一起接入。").FontSize(11).TextColor(t.TextMuted)
			}).Grow(1)
		})

		// 3. Status & Health Cards (2 Equal-Height Columns)
		ui.Row(c).Gap(14).AlignItems(ui.Stretch).Children(func() {
			// Clash Core Health
			Card(c, func() {
				ui.Text(c, "核心运行状态").FontSize(14).Bold()

				coreVersion := "未启动"
				if a.status.Running {
					coreVersion = "mihomo " + a.status.Version
				}
				StatRow(c, "核心版本", coreVersion)

				uptimeStr := "未运行"
				if a.status.Running {
					uptimeStr = uptime(a.status.Since)
				}
				StatRow(c, "运行时间", uptimeStr)

				memStr := bytes(int64(a.status.Memory))
				if !a.status.Running {
					memStr = "0 B"
				}
				StatRow(c, "内存占用", memStr)
				StatRow(c, "活跃连接", fmt.Sprintf("%d 个连接", a.status.Connections))
				StatRow(c, "混合端口", fmt.Sprintf("%d", mixedPort))
			}).Grow(1)

			// Quick Action Shortcuts
			Card(c, func() {
				ui.Text(c, "常用快捷工具").FontSize(14).Bold()

				ui.Column(c).Gap(10).Grow(1).Justify(ui.Center).Children(func() {
					ui.Row(c).Gap(10).Children(func() {
						if ui.Button(c, "一键全部测速").Grow(1).Clicked() {
							a.currentTab = "proxies"
						}
						if ui.Button(c, "刷新订阅配置").Grow(1).Clicked() {
							a.currentTab = "profiles"
						}
					})

					ui.Row(c).Gap(10).Children(func() {
						if ui.Button(c, "断开全部连接").Grow(1).Clicked() {
							a.connections = nil
						}
						if ui.Button(c, "查看运行日志").Grow(1).Clicked() {
							a.currentTab = "logs"
						}
					})
				})

				ui.Text(c, "提示: 可随时通过顶部工具栏快速切换规则 / 全局 / 直连模式。").FontSize(11).TextColor(t.TextMuted)
			}).Grow(1)
		})
	})
}
