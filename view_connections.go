package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

func (a *app) viewConnections(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Grow(1).Gap(16).Children(func() {
		// Top Filter & Action Bar
		Card(c, func() {
			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.SearchField(c, &a.connSearch).Placeholder("搜索域名 / IP / 进程...").Width(260)
					ui.Badge(c, fmt.Sprintf("%d 个活跃连接", len(a.connections))).
						Background(t.Accent.Alpha(0.14)).
						TextColor(t.Accent)
				})

				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					if ui.Button(c, "断开全部连接").Clicked() {
						a.connections = nil
					}
				})
			})
		})

		// Filter connections by query
		query := strings.ToLower(strings.TrimSpace(a.connSearch))
		var filtered []ActiveConn
		for _, conn := range a.connections {
			if query == "" ||
				strings.Contains(strings.ToLower(conn.Host), query) ||
				strings.Contains(strings.ToLower(conn.Process), query) ||
				strings.Contains(strings.ToLower(conn.Proxy), query) {
				filtered = append(filtered, conn)
			}
		}

		// Connections Table
		columns := []ui.TableColumn{
			{Title: "目标地址 / 端口", ID: "host", Width: 185},
			{Title: "应用程序", ID: "process", Width: 115},
			{Title: "命中规则", ID: "rule", Width: 140},
			{Title: "出站节点", ID: "proxy", Width: 125},
			{Title: "实时速率", ID: "speed", Width: 145},
			{Title: "累计流量", ID: "total", Width: 80},
		}

		Card(c, func() {
			if len(filtered) == 0 {
				ui.Column(c).Grow(1).Center().Padding(40).Gap(8).Children(func() {
					ui.Text(c, "暂无活跃网络连接").FontSize(14).TextColor(t.TextMuted)
				})
				return
			}

			ui.Table(c, &a.connListState, columns, len(filtered), func(row int, col int) {
				conn := filtered[row]
				switch col {
				case 0:
					ui.Text(c, conn.Host).SingleLine().Bold()
				case 1:
					ui.Badge(c, conn.Process).
						Background(t.Text.Alpha(0.08)).
						TextColor(t.TextMuted)
				case 2:
					ui.Text(c, conn.Rule).FontSize(12).SingleLine()
				case 3:
					ui.Text(c, conn.Proxy).FontSize(12).TextColor(t.Accent).Bold().SingleLine()
				case 4:
					ui.Textf(c, "↑%s ↓%s", conn.UpSpeed, conn.DownSpeed).FontSize(12).SingleLine()
				case 5:
					ui.Text(c, conn.TotalTransfer).FontSize(12).TextColor(t.TextMuted).SingleLine()
				}
			}).Grow(1)
		}).Grow(1).Padding(0).Clip()
	})
}
