package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Card builds an elevated native-style card container whose background contrasts
// cleanly with the window canvas and with controls (Button, Segmented) inside it.
func Card(c *ui.Context, fn func()) ui.Element {
	t := c.Theme()
	cardBg := t.Background
	borderColor := t.Border.Alpha(0.75)
	shadowColor := ui.RGBA(0, 0, 0, 0.03)
	if t.Dark {
		cardBg = t.Background.Mix(t.Surface, 0.5)
		borderColor = t.Border
		shadowColor = ui.RGBA(0, 0, 0, 0.22)
	}

	return ui.Column(c).
		Background(cardBg).
		Border(1, borderColor).
		Radius(10).
		Shadow(0, 1, 4, 0, shadowColor).
		Padding(14, 18).
		Gap(10).
		Children(fn)
}

// LatencyBadge builds a refined colored pill badge indicating node response latency.
func LatencyBadge(c *ui.Context, ms int) ui.Element {
	t := c.Theme()
	if ms < 0 {
		return ui.Badge(c, "超时").
			Background(t.Danger.Alpha(0.14)).
			TextColor(t.Danger)
	}

	label := fmt.Sprintf("%d ms", ms)
	if ms < 100 {
		return ui.Badge(c, label).
			Background(t.Success.Alpha(0.14)).
			TextColor(t.Success)
	}
	if ms < 300 {
		return ui.Badge(c, label).
			Background(t.Warning.Alpha(0.14)).
			TextColor(t.Warning)
	}
	return ui.Badge(c, label).
		Background(t.Danger.Alpha(0.14)).
		TextColor(t.Danger)
}

// StatRow renders a key-value row with label and value nicely aligned.
func StatRow(c *ui.Context, label string, value string) {
	t := c.Theme()
	ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(13).TextColor(t.TextMuted)
		ui.Text(c, value).FontSize(13).Bold()
	})
}
