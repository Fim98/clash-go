package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

func (a *app) viewProfiles(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Fill().Gap(16).Children(func() {
		// Top Action Toolbar
		Card(c, func() {
			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					if ui.PrimaryButton(c, "+ 添加订阅").Clicked() {
						a.addProfileModalOpen = true
						a.newProfileName = ""
						a.newProfileURL = ""
					}
					if ui.Button(c, "📥 导入本地文件").Clicked() {
						// Local file import
					}
				})

				if ui.Button(c, "⟳ 一键更新全部").Clicked() {
					// Update all profiles
				}
			})
		})

		// Profile Cards List
		ui.Column(c).Gap(12).Fill().Children(func() {
			for i := range a.profiles {
				p := &a.profiles[i]

				Card(c, func() {
					ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
						// Left: Active radio indicator & Info
						ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
							// Active Radio Switch Button
							radioBtn := ui.Button(c, "○").Width(36)
							if p.Active {
								radioBtn = ui.PrimaryButton(c, "●").Width(36)
							}
							if radioBtn.Clicked() {
								for j := range a.profiles {
									a.profiles[j].Active = false
								}
								p.Active = true
							}

							// Profile Details
							ui.Column(c).Gap(4).Children(func() {
								ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
									ui.Text(c, p.Name).FontSize(15).Bold()
									if p.Active {
										ui.Badge(c, "使用中").Background(ui.RGBA(34, 197, 94, 0.15)).TextColor(t.Success)
									}
									typeLabel := "远程订阅"
									if p.Type == "local" {
										typeLabel = "本地文件"
									}
									ui.Badge(c, typeLabel).Background(ui.RGBA(128, 128, 128, 0.12)).TextColor(t.TextMuted)
								})

								ui.Textf(c, "%s • 包含 %d 个节点 • 更新于 %s", p.URL, p.NodeCount, p.UpdatedAt).
									FontSize(12).
									TextColor(t.TextMuted).
									SingleLine()
							})
						})

						// Right: Action buttons
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							if ui.Button(c, "⟳ 更新").Clicked() {
								// Update profile
							}
							if ui.Button(c, "✎ 编辑").Clicked() {
								// Edit profile
							}
							if ui.Button(c, "🗑").Clicked() {
								// Delete profile
							}
						})
					})

					// Traffic Progress Bar (if remote quota available)
					if p.TotalGB > 0 {
						pct := p.UsedGB / p.TotalGB
						ui.Column(c).Gap(6).Children(func() {
							ui.Progress(c, pct)
							ui.Row(c).Justify(ui.SpaceBetween).Children(func() {
								ui.Textf(c, "已用流量: %.1f GB / %.1f GB (%.1f%%)", p.UsedGB, p.TotalGB, pct*100).
									FontSize(11).
									TextColor(t.TextMuted)
								ui.Text(c, "到期时间: 2027-12-31").
									FontSize(11).
									TextColor(t.TextMuted)
							})
						})
					}
				})
			}
		})

		// Modal for Adding Profile
		ui.Modal(c, &a.addProfileModalOpen, func() {
			ui.Column(c).
				Background(t.Surface).
				Border(1, t.Border).
				Radius(10).
				Padding(24).
				Gap(16).
				Width(440).
				Children(func() {
					ui.Text(c, "添加远程订阅").FontSize(18).Bold()
					ui.Text(c, "请输入服务商提供的 Clash 订阅链接").FontSize(12).TextColor(t.TextMuted)

					ui.Column(c).Gap(6).Children(func() {
						ui.Text(c, "配置名称").FontSize(13).Bold()
						ui.TextInput(c, &a.newProfileName).Placeholder("例如：主力机场订阅")
					})

					ui.Column(c).Gap(6).Children(func() {
						ui.Text(c, "订阅 URL").FontSize(13).Bold()
						ui.TextInput(c, &a.newProfileURL).Placeholder("https://...")
					})

					ui.Row(c).Justify(ui.End).Gap(10).Children(func() {
						if ui.Button(c, "取消").Clicked() {
							a.addProfileModalOpen = false
						}
						if ui.PrimaryButton(c, "保存并下载").Clicked() {
							if a.newProfileName != "" && a.newProfileURL != "" {
								a.profiles = append(a.profiles, ProfileItem{
									ID:        fmt.Sprintf("prof-%d", len(a.profiles)+1),
									Name:      a.newProfileName,
									Type:      "remote",
									URL:       a.newProfileURL,
									UpdatedAt: "刚刚",
									TotalGB:   200.0,
									UsedGB:    0.0,
									Active:    false,
									NodeCount: 32,
								})
								a.addProfileModalOpen = false
							}
						}
					})
				})
		})
	})
}
