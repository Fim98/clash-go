package main

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

func (a *app) viewLogs(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Fill().Gap(16).Children(func() {
		// Top Control Bar
		Card(c, func() {
			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				// Level Selector & Search
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Segmented(c, &a.logLevelIdx, "全部", "INFO", "WARN", "ERROR", "DEBUG")
					ui.SearchField(c, &a.logSearch).Placeholder("过滤日志关键词...").Width(220)
				})

				// Right Controls
				ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "自动滚屏").FontSize(12).TextColor(t.TextMuted)
						ui.Switch(c, &a.autoScroll)
					})
					if ui.Button(c, "🗑 清空日志").Clicked() {
						a.logs = nil
					}
				})
			})
		})

		// Filter logs
		levels := []string{"", "INFO", "WARN", "ERROR", "DEBUG"}
		activeLevel := ""
		if a.logLevelIdx > 0 && a.logLevelIdx < len(levels) {
			activeLevel = levels[a.logLevelIdx]
		}
		query := strings.ToLower(strings.TrimSpace(a.logSearch))

		var filtered []LogEntry
		for _, entry := range a.logs {
			if activeLevel != "" && entry.Level != activeLevel {
				continue
			}
			if query != "" && !strings.Contains(strings.ToLower(entry.Message), query) {
				continue
			}
			filtered = append(filtered, entry)
		}

		// Log Stream Box
		Card(c, func() {
			if len(filtered) == 0 {
				ui.Column(c).Center().Padding(40).Children(func() {
					ui.Text(c, "暂无符合条件的日志记录").FontSize(14).TextColor(t.TextMuted)
				})
				return
			}

			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(8).Padding(4).Children(func() {
					for _, entry := range filtered {
						ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
							// Timestamp
							ui.Text(c, entry.Time).FontSize(11).TextColor(t.TextMuted)

							// Level Badge
							badge := ui.Badge(c, entry.Level)
							switch entry.Level {
							case "INFO":
								badge.Background(ui.RGBA(59, 130, 246, 0.15)).TextColor(t.Accent)
							case "WARN":
								badge.Background(ui.RGBA(234, 179, 8, 0.15)).TextColor(t.Warning)
							case "ERROR":
								badge.Background(ui.RGBA(239, 68, 68, 0.15)).TextColor(t.Danger)
							default:
								badge.Background(ui.RGBA(128, 128, 128, 0.12)).TextColor(t.TextMuted)
							}

							// Message Content
							ui.Text(c, entry.Message).FontSize(12).SingleLine()
						})
					}
				})
			})
		})
	})
}
