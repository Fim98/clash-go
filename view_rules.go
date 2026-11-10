package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

func (a *app) viewRules(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Fill().Gap(16).Children(func() {
		// Page Header
		SectionHeader(c, "分流规则 (Rules)", "流量路由决策表与规则集 (Rule Providers) 监控")

		// Top Search Bar
		Card(c, func() {
			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.SearchField(c, &a.ruleSearch).Placeholder("搜索域名 / 规则类型 / 策略...").Width(280)
					ui.Badge(c, fmt.Sprintf("总计 %d 条规则", len(a.rules))).
						Background(ui.RGBA(128, 128, 128, 0.15)).
						TextColor(t.TextMuted)
				})

				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					if ui.Button(c, "⟳ 更新规则集").Clicked() {
						// Update rule providers
					}
				})
			})
		})

		// Rule Providers Overview Card
		Card(c, func() {
			ui.Text(c, "规则集 Providers 概况").FontSize(14).Bold()

			ui.Row(c).Gap(16).Children(func() {
				// Provider 1
				ui.Column(c).Grow(1).Padding(10).Background(ui.RGBA(128, 128, 128, 0.06)).Radius(6).Gap(4).Children(func() {
					ui.Row(c).Justify(ui.SpaceBetween).Children(func() {
						ui.Text(c, "🛡️ Reject-Rules").FontSize(13).Bold()
						ui.Badge(c, "1,200 条").Background(ui.RGBA(239, 68, 68, 0.15)).TextColor(t.Danger)
					})
					ui.Text(c, "行为: domain • 远端同步 • 1小时前更新").FontSize(11).TextColor(t.TextMuted)
				})

				// Provider 2
				ui.Column(c).Grow(1).Padding(10).Background(ui.RGBA(128, 128, 128, 0.06)).Radius(6).Gap(4).Children(func() {
					ui.Row(c).Justify(ui.SpaceBetween).Children(func() {
						ui.Text(c, "🇨🇳 Direct-GeoIP").FontSize(13).Bold()
						ui.Badge(c, "4,500 条").Background(ui.RGBA(34, 197, 94, 0.15)).TextColor(t.Success)
					})
					ui.Text(c, "行为: ipcidr • 远端同步 • 3小时前更新").FontSize(11).TextColor(t.TextMuted)
				})
			})
		})

		// Filter rules by query
		query := strings.ToLower(strings.TrimSpace(a.ruleSearch))
		var filtered []RoutingRule
		for _, rule := range a.rules {
			if query == "" ||
				strings.Contains(strings.ToLower(rule.Type), query) ||
				strings.Contains(strings.ToLower(rule.Payload), query) ||
				strings.Contains(strings.ToLower(rule.Proxy), query) {
				filtered = append(filtered, rule)
			}
		}

		// Rules Table
		columns := []ui.TableColumn{
			{Title: "#", ID: "index", Width: 50},
			{Title: "规则类型", ID: "type", Width: 160},
			{Title: "匹配内容 (Payload)", ID: "payload", Width: 280},
			{Title: "目标策略组 / 动作", ID: "proxy", Width: 180},
		}

		Card(c, func() {
			if len(filtered) == 0 {
				ui.Column(c).Center().Padding(40).Children(func() {
					ui.Text(c, "没有匹配的分流规则").FontSize(14).TextColor(t.TextMuted)
				})
				return
			}

			ui.Table(c, &a.ruleListState, columns, len(filtered), func(row int, col int) {
				rule := filtered[row]
				switch col {
				case 0:
					ui.Textf(c, "%d", rule.Index).FontSize(12).TextColor(t.TextMuted)
				case 1:
					ui.Badge(c, rule.Type).
						Background(ui.RGBA(59, 130, 246, 0.12)).
						TextColor(t.Accent)
				case 2:
					payload := rule.Payload
					if payload == "" {
						payload = "(全局兜底匹配)"
						ui.Text(c, payload).FontSize(12).TextColor(t.TextMuted).SingleLine()
					} else {
						ui.Text(c, payload).FontSize(13).Bold().SingleLine()
					}
				case 3:
					ui.Text(c, rule.Proxy).FontSize(13).TextColor(t.Accent).Bold().SingleLine()
				}
			}).Grow(1)
		})
	})
}
