package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Card builds an elevated modern card container with subtle shadow and border.
func Card(c *ui.Context, fn func()) ui.Element {
	t := c.Theme()
	shadowColor := ui.RGBA(0, 0, 0, 0.04)
	if t.Dark {
		shadowColor = ui.RGBA(0, 0, 0, 0.25)
	}

	return ui.Column(c).
		Background(t.Surface).
		Border(1, t.Border).
		Radius(10).
		Shadow(0, 2, 6, 0, shadowColor).
		Padding(18).
		Gap(14).
		Children(fn)
}

// LatencyBadge builds a refined colored pill badge indicating node response latency.
func LatencyBadge(c *ui.Context, ms int) ui.Element {
	t := c.Theme()
	if ms < 0 {
		return ui.Badge(c, "超时").
			Background(ui.RGBA(239, 68, 68, 0.15)).
			TextColor(t.Danger)
	}

	label := fmt.Sprintf("%d ms", ms)
	if ms < 100 {
		return ui.Badge(c, label).
			Background(ui.RGBA(34, 197, 94, 0.15)).
			TextColor(t.Success)
	}
	if ms < 300 {
		return ui.Badge(c, label).
			Background(ui.RGBA(234, 179, 8, 0.15)).
			TextColor(t.Warning)
	}
	return ui.Badge(c, label).
		Background(ui.RGBA(239, 68, 68, 0.15)).
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
