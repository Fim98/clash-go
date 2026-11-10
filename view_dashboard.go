package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
)

func (a *app) viewDashboard(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Fill().Gap(16).Children(func() {
		// Page Header
		SectionHeader(c, "仪表盘 (Dashboard)", "实时网络流量与核心状态监控")

		// 1. Realtime Traffic Chart Card
		Card(c, func() {
			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "实时网速监控").FontSize(15).Bold()
					ui.Text(c, "(近 60 秒)").FontSize(12).TextColor(t.TextMuted)
				})

				// Realtime Indicators
				ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Badge(c, "↑ 上行").Background(ui.RGBA(234, 179, 8, 0.15)).TextColor(t.Warning)
						ui.Text(c, rate(a.status.Up)).FontSize(13).Bold()
					})
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Badge(c, "↓ 下行").Background(ui.RGBA(59, 130, 246, 0.15)).TextColor(t.Accent)
						ui.Text(c, rate(a.status.Down)).FontSize(13).Bold()
					})
				})
			})

			// Chart Plot
			yScale := chart.NewLinear(0, chartMax(a.trafficHistoryUp, a.trafficHistoryDown)).Nice(4)
			chart.Plot(c, chart.Axes{Left: yScale.Ticks(4)}, func(area ui.Element) {
				area.Draw(func(p *ui.Painter, r ui.Rect) {
					// Draw background grid lines
					for _, tick := range yScale.Ticks(4) {
						yPos := chart.Y(r, tick.At)
						p.Line(r.X, yPos, r.X+r.W, yPos, 1, ui.RGBA(128, 128, 128, 0.15))
					}

					// Build normalized points for Up traffic
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

						// Draw Up curve (Orange / Warning)
						upPath := chart.Line(r, xs, ysUp, chart.CurveMonotone)
						p.StrokePath(upPath, 2.5, t.Warning)

						// Draw Down curve (Blue / Accent)
						downPath := chart.Line(r, xs, ysDown, chart.CurveMonotone)
						p.StrokePath(downPath, 2.5, t.Accent)
					}
				})
			}).Height(160).Gap(8)
		})

		// 2. Control Cards Grid (2 Columns)
		ui.Row(c).Gap(16).Fill().Children(func() {
			// Left Card: Quick Control
			ui.Column(c).Grow(1).Children(func() {
				Card(c, func() {
					ui.Text(c, "快捷网络开关").FontSize(15).Bold()

					// System Proxy Switch
					ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
						ui.Column(c).Gap(2).Children(func() {
							ui.Text(c, "系统代理 (System Proxy)").FontSize(13).Bold()
							ui.Text(c, "自动配置系统网络代理端口 127.0.0.1:7890").FontSize(11).TextColor(t.TextMuted)
						})
						ui.Switch(c, &a.systemProxyOn)
					})

					ui.Divider(c)

					// TUN Mode Switch
					ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
						ui.Column(c).Gap(2).Children(func() {
							ui.Text(c, "TUN 虚拟网卡模式").FontSize(13).Bold()
							ui.Text(c, "接管全局流量并绕过应用代理设置").FontSize(11).TextColor(t.TextMuted)
						})
						ui.Switch(c, &a.tunModeOn)
					})

					ui.Divider(c)

					// Outbound Mode
					ui.Column(c).Gap(6).Children(func() {
						ui.Text(c, "分流模式").FontSize(13).Bold()
						ui.Segmented(c, &a.outboundMode, "规则模式", "全局模式", "直连模式")
					})
				})
			})

			// Right Card: External IP & Environment
			ui.Column(c).Grow(1).Children(func() {
				Card(c, func() {
					ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "出口 IP 与网络环境").FontSize(15).Bold()
						ui.Badge(c, "M5").Background(ui.RGBA(128, 128, 128, 0.15)).TextColor(t.TextMuted)
					})

					StatRow(c, "出口 IP 地址", "—")
					StatRow(c, "地理位置", "—")
					StatRow(c, "运营商与 ASN", "—")
					StatRow(c, "本地直连延迟", "—")
					ui.Text(c, "探测出口 IP 需要一个外部服务，随系统集成一起接入。").FontSize(11).TextColor(t.TextMuted)
				})
			})
		})

		// 3. Status & Health Cards (2 Columns)
		ui.Row(c).Gap(16).Fill().Children(func() {
			// Clash Core Health
			ui.Column(c).Grow(1).Children(func() {
				Card(c, func() {
					ui.Text(c, "核心运行状态").FontSize(15).Bold()

					coreVersion := "未启动"
					if a.status.Running {
						coreVersion = "mihomo (Meta) 内嵌"
					}
					StatRow(c, "核心版本", coreVersion)
					StatRow(c, "运行时间", uptime(a.status.Since))
					StatRow(c, "内存占用", bytes(int64(a.status.Memory)))
					StatRow(c, "活跃连接", fmt.Sprintf("%d 个连接", a.status.Connections))
					StatRow(c, "混合端口", fmt.Sprintf("%d", mixedPort))
				})
			})

			// Quick Actions
			ui.Column(c).Grow(1).Children(func() {
				Card(c, func() {
					ui.Text(c, "快捷操作").FontSize(15).Bold()

					ui.Row(c).Gap(10).Children(func() {
						if ui.Button(c, "⚡ 一键测速").Grow(1).Clicked() {
							// Trigger latency check
						}
						if ui.Button(c, "⟳ 更新订阅").Grow(1).Clicked() {
							// Trigger profile update
						}
					})

					ui.Row(c).Gap(10).Children(func() {
						if ui.Button(c, "🗑 断开连接").Grow(1).Clicked() {
							// Close all connections
						}
						if ui.Button(c, "📁 配置目录").Grow(1).Clicked() {
							// Open config directory
						}
					})
				})
			})
		})
	})
}
