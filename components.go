package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Card builds a modern card container with a subtle surface background and border.
func Card(c *ui.Context, fn func()) ui.Element {
	t := c.Theme()
	return ui.Column(c).
		Background(t.Surface).
		Border(1, t.Border).
		Radius(8).
		Padding(16).
		Gap(12).
		Children(fn)
}

// SectionHeader displays a title and an optional subtitle description.
func SectionHeader(c *ui.Context, title string, desc string) {
	t := c.Theme()
	ui.Column(c).Gap(4).Children(func() {
		ui.Text(c, title).FontSize(18).Bold()
		if desc != "" {
			ui.Text(c, desc).FontSize(12).TextColor(t.TextMuted)
		}
	})
}

// LatencyBadge builds a colored badge indicating node response latency.
func LatencyBadge(c *ui.Context, ms int) ui.Element {
	t := c.Theme()
	if ms < 0 {
		return ui.Badge(c, "Timeout").
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
