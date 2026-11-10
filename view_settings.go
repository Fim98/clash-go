package main

import (
	"github.com/egoist/mygo/ui"
)

func (a *app) viewSettings(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Fill().Gap(16).Children(func() {
		// Page Header
		SectionHeader(c, "偏好设置 (Settings)", "Clash 核心网络、系统代理旁路、TUN 网卡栈与外观交互")

		// 1. Clash Core Network Settings
		Card(c, func() {
			ui.Text(c, "Clash 核心与网络监听").FontSize(15).Bold()

			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					ui.Text(c, "允许来自局域网的连接 (Allow LAN)").FontSize(13).Bold()
					ui.Text(c, "同一 Wi-Fi 或局域网下的手机与其他设备可连入代理").FontSize(11).TextColor(t.TextMuted)
				})
				ui.Switch(c, &a.allowLAN)
			})

			ui.Divider(c)

			ui.Row(c).Gap(16).Children(func() {
				ui.Column(c).Grow(1).Gap(6).Children(func() {
					ui.Text(c, "混合代理端口 (Mixed Port)").FontSize(13).Bold()
					ui.TextInput(c, &a.mixedPort).Placeholder("7890")
				})
				ui.Column(c).Grow(1).Gap(6).Children(func() {
					ui.Text(c, "外部控制器端口 (Controller Port)").FontSize(13).Bold()
					ui.TextInput(c, &a.controllerPort).Placeholder("9090")
				})
			})

			ui.Column(c).Gap(6).Children(func() {
				ui.Text(c, "REST API 访问密钥 (Secret)").FontSize(13).Bold()
				ui.TextInput(c, &a.secret).Placeholder("留空表示免密通信")
			})
		})

		// 2. System Proxy & TUN Mode Settings
		Card(c, func() {
			ui.Text(c, "系统代理与 TUN 虚拟网卡").FontSize(15).Bold()

			ui.Column(c).Gap(6).Children(func() {
				ui.Text(c, "系统代理绕过名单 (Bypass Domain List)").FontSize(13).Bold()
				ui.TextInput(c, &a.bypassDomains).Placeholder("localhost, 127.0.0.1, 192.168.0.0/16")
			})

			ui.Divider(c)

			ui.Column(c).Gap(6).Children(func() {
				ui.Text(c, "TUN 虚拟网卡网络协议栈").FontSize(13).Bold()
				ui.Segmented(c, &a.tunStackIdx, "gVisor (轻量推荐)", "System (系统原生)", "Mixed (混合)")
			})
		})

		// 3. Appearance & App Behavior Settings
		Card(c, func() {
			ui.Text(c, "界面外观与常规行为").FontSize(15).Bold()

			ui.Column(c).Gap(6).Children(func() {
				ui.Text(c, "主题模式选择").FontSize(13).Bold()
				ui.Segmented(c, &a.themeModeIdx, "跟随系统", "明亮浅色", "优雅深色")
			})

			ui.Divider(c)

			ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					ui.Text(c, "登录系统时开机自启动").FontSize(13).Bold()
					ui.Text(c, "系统启动时自动在后台启动 Clash 核心与系统托盘").FontSize(11).TextColor(t.TextMuted)
				})
				ui.Switch(c, &a.autoStart)
			})
		})

		// Save Button
		ui.Row(c).Justify(ui.End).Children(func() {
			if ui.PrimaryButton(c, "保存并应用配置").Clicked() {
				// Save preferences
			}
		})
	})
}
