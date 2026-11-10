package main

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

func (a *app) viewProxies(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Gap(16).Children(func() {
		// Top Control Bar
		Card(c, func() {
			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.SearchField(c, &a.proxySearch).Placeholder("搜索节点关键词 / 地区...").Width(260)

				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					if ui.Segmented(c, &a.outboundMode, "规则", "全局", "直连").Changed() {
						a.setMode(modes[a.outboundMode].mode)
					}
					if ui.PrimaryButton(c, "全部测速").Clicked() {
						// Speed test trigger
					}
				})
			})
		})

		// Proxy Groups Tab Bar
		groupLabels := make([]string, len(a.proxyGroups))
		for i, g := range a.proxyGroups {
			groupLabels[i] = g.Name
		}
		if a.proxyGroupTab < 0 || a.proxyGroupTab >= len(a.proxyGroups) {
			a.proxyGroupTab = 0
		}

		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Tabs(c, &a.proxyGroupTab, groupLabels...)
		})

		// Active Group Info & Current Selection
		curGroup := &a.proxyGroups[a.proxyGroupTab]
		ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, curGroup.Name).FontSize(14).Bold()
				ui.Badge(c, curGroup.Type).Background(t.Text.Alpha(0.08)).TextColor(t.TextMuted)
			})
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "当前生效:").FontSize(12).TextColor(t.TextMuted)
				ui.Text(c, curGroup.Selected).FontSize(13).Bold().TextColor(t.Accent)
			})
		})

		// Node Cards Grid (3 equal columns stretching across full width)
		ui.Grid(c).Columns(3).Gap(12).Children(func() {
			filter := strings.ToLower(strings.TrimSpace(a.proxySearch))
			for i := range curGroup.Nodes {
				node := &curGroup.Nodes[i]
				if filter != "" && !strings.Contains(strings.ToLower(node.Name), filter) {
					continue
				}

				isSelected := (curGroup.Selected == node.Name)

				cardBg := t.Background
				if t.Dark {
					cardBg = t.Background.Mix(t.Surface, 0.5)
				}
				cardCol := ui.Column(c).
					Key(node.Name).
					Background(cardBg).
					Radius(8).
					Padding(12, 14).
					Gap(8).
					Cursor(ui.CursorPointer)

				if isSelected {
					cardCol.Border(1.5, t.Accent).Background(t.Accent.Alpha(0.08))
				} else {
					cardCol.Border(1, t.Border.Alpha(0.75))
				}

				if cardCol.Children(func() {
					ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
						ui.Text(c, node.Name).FontSize(13).Bold().SingleLine()
						if isSelected {
							ui.Text(c, "✓").TextColor(t.Accent).Bold()
						}
					})

					ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
						ui.Badge(c, node.Type).
							Background(t.Text.Alpha(0.08)).
							TextColor(t.TextMuted)
						LatencyBadge(c, node.Latency)
					})
				}).Clicked() {
					curGroup.Selected = node.Name
				}
			}
		})
	})
}
