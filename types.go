package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"clash-go/internal/core"
)

// ProxyNode represents a single proxy node.
type ProxyNode struct {
	Name     string
	Type     string
	Latency  int // in ms; -1 means timeout/untested
	Selected bool
	IP       string
}

// ProxyGroup represents a policy group (e.g., SELECT, URL-Test, Fallback).
type ProxyGroup struct {
	Name     string
	Type     string
	Selected string
	Nodes    []ProxyNode
}

// ProfileItem represents a Clash configuration subscription.
type ProfileItem struct {
	ID        string
	Name      string
	Type      string // "remote" or "local"
	URL       string
	UpdatedAt string
	TotalGB   float64
	UsedGB    float64
	Active    bool
	NodeCount int
}

// ActiveConn represents an active network connection tracked by the core.
type ActiveConn struct {
	ID            string
	Host          string
	Process       string
	Rule          string
	Proxy         string
	UpSpeed       string
	DownSpeed     string
	TotalTransfer string
	Duration      string
}

// RoutingRule represents a single Clash routing rule.
type RoutingRule struct {
	Index   int
	Type    string
	Payload string
	Proxy   string
}

// LogEntry represents a core log record.
type LogEntry struct {
	Time    string
	Level   string // "INFO", "WARN", "ERROR", "DEBUG"
	Message string
}

// app is the central state model for the client UI.
type app struct {
	mgr *core.Manager
	win *mygo.Window

	status core.Status
	busy   bool
	err    string

	// Navigation & Shell
	currentTab   string // "dashboard", "proxies", "profiles", "connections", "rules", "logs", "settings"
	outboundMode int    // 0: 规则 (Rule), 1: 全局 (Global), 2: 直连 (Direct)

	// Quick Toggles
	systemProxyOn bool
	tunModeOn     bool
	allowLAN      bool
	autoStart     bool

	// Dashboard Traffic History
	trafficHistoryUp   []float32
	trafficHistoryDown []float32

	// Proxies View State
	proxyGroupTab  int
	proxySearch    string
	proxySortBy    int // 0: 默认, 1: 延迟, 2: 名称
	proxyGroups    []ProxyGroup
	proxyGridState ui.GridState

	// Profiles View State
	profiles            []ProfileItem
	addProfileModalOpen bool
	newProfileName      string
	newProfileURL       string

	// Connections View State
	connSearch    string
	connListState ui.ListState
	connections   []ActiveConn

	// Rules View State
	ruleSearch    string
	ruleListState ui.ListState
	rules         []RoutingRule

	// Logs View State
	logLevelIdx int // 0: 全部, 1: INFO, 2: WARN, 3: ERROR, 4: DEBUG
	logSearch   string
	autoScroll  bool
	logs        []LogEntry

	// Settings View State
	mixedPort      string
	controllerPort string
	secret         string
	themeModeIdx   int // 0: 跟随系统, 1: 浅色模式, 2: 深色模式
	tunStackIdx    int // 0: gVisor, 1: System, 2: Mixed
	bypassDomains  string
}

func newApp(mgr *core.Manager) *app {
	return &app{
		mgr:            mgr,
		currentTab:     "dashboard",
		outboundMode:   0, // Rule
		systemProxyOn:  true,
		tunModeOn:      false,
		allowLAN:       false,
		autoStart:      true,
		mixedPort:      "7890",
		controllerPort: "9090",
		secret:         "clash-secret",
		themeModeIdx:   0,
		tunStackIdx:    0,
		bypassDomains:  "localhost, 127.0.0.1, 192.168.0.0/16, 10.0.0.0/8",
		autoScroll:     true,

		trafficHistoryUp:   make([]float32, historyLen),
		trafficHistoryDown: make([]float32, historyLen),

		proxyGroups: []ProxyGroup{
			{
				Name:     "🚀 节点选择",
				Type:     "Select",
				Selected: "🇭🇰 HK-BGP-01",
				Nodes: []ProxyNode{
					{Name: "🇭🇰 HK-BGP-01", Type: "Shadowsocks", Latency: 42, Selected: true, IP: "103.21.244.12"},
					{Name: "🇭🇰 HK-IEPL-02", Type: "Trojan", Latency: 56, Selected: false, IP: "103.21.244.15"},
					{Name: "🇯🇵 JP-Tokyo-01", Type: "Vmess", Latency: 78, Selected: false, IP: "168.12.33.91"},
					{Name: "🇯🇵 JP-Osaka-02", Type: "Vless", Latency: 84, Selected: false, IP: "168.12.33.95"},
					{Name: "🇸🇬 SG-Direct-01", Type: "Hysteria2", Latency: 65, Selected: false, IP: "139.180.140.2"},
					{Name: "🇺🇸 US-Silicon-01", Type: "Trojan", Latency: 165, Selected: false, IP: "142.250.190.46"},
					{Name: "🇺🇸 US-Oregon-02", Type: "Shadowsocks", Latency: 182, Selected: false, IP: "142.250.190.48"},
					{Name: "🇩🇪 DE-Frankfurt-01", Type: "Wireguard", Latency: 220, Selected: false, IP: "194.12.44.10"},
					{Name: "🇬🇧 UK-London-01", Type: "Vmess", Latency: -1, Selected: false, IP: "185.12.30.5"},
				},
			},
			{
				Name:     "🤖 AI 专线",
				Type:     "URLTest",
				Selected: "🇺🇸 US-Silicon-01",
				Nodes: []ProxyNode{
					{Name: "🇺🇸 US-Silicon-01", Type: "Trojan", Latency: 165, Selected: true, IP: "142.250.190.46"},
					{Name: "🇯🇵 JP-Tokyo-01", Type: "Vmess", Latency: 78, Selected: false, IP: "168.12.33.91"},
					{Name: "🇸🇬 SG-Direct-01", Type: "Hysteria2", Latency: 65, Selected: false, IP: "139.180.140.2"},
				},
			},
			{
				Name:     "🐟 漏网之鱼",
				Type:     "Select",
				Selected: "DIRECT",
				Nodes: []ProxyNode{
					{Name: "DIRECT", Type: "Direct", Latency: 8, Selected: true, IP: "127.0.0.1"},
					{Name: "🚀 节点选择", Type: "Group", Latency: 42, Selected: false, IP: "-"},
				},
			},
			{
				Name:     "🛑 广告拦截",
				Type:     "Select",
				Selected: "REJECT",
				Nodes: []ProxyNode{
					{Name: "REJECT", Type: "Reject", Latency: 0, Selected: true, IP: "0.0.0.0"},
					{Name: "DIRECT", Type: "Direct", Latency: 8, Selected: false, IP: "127.0.0.1"},
				},
			},
		},

		profiles: []ProfileItem{
			{
				ID:        "prof-1",
				Name:      "主力高速专线 (订阅)",
				Type:      "remote",
				URL:       "https://subscribe.example.com/api/v1/client/subscribe?token=xyz",
				UpdatedAt: "10 分钟前",
				TotalGB:   500.0,
				UsedGB:    142.5,
				Active:    true,
				NodeCount: 48,
			},
			{
				ID:        "prof-2",
				Name:      "个人自建备用节点",
				Type:      "local",
				URL:       "config/custom-vps.yaml",
				UpdatedAt: "昨天 18:30",
				TotalGB:   200.0,
				UsedGB:    18.2,
				Active:    false,
				NodeCount: 6,
			},
			{
				ID:        "prof-3",
				Name:      "公司内网分流配置",
				Type:      "local",
				URL:       "config/corporate.yaml",
				UpdatedAt: "3 天前",
				TotalGB:   100.0,
				UsedGB:    8.9,
				Active:    false,
				NodeCount: 12,
			},
		},

		connections: []ActiveConn{
			{ID: "c-1", Host: "api.github.com:443", Process: "git-credential-osx", Rule: "Match(Proxy)", Proxy: "🇭🇰 HK-BGP-01", UpSpeed: "1.2 KB/s", DownSpeed: "18.4 KB/s", TotalTransfer: "245 KB", Duration: "45s"},
			{ID: "c-2", Host: "www.google.com:443", Process: "Google Chrome", Rule: "Rule(GFWList)", Proxy: "🇭🇰 HK-BGP-01", UpSpeed: "8.5 KB/s", DownSpeed: "340 KB/s", TotalTransfer: "2.1 MB", Duration: "3m 12s"},
			{ID: "c-3", Host: "chat.openai.com:443", Process: "Google Chrome", Rule: "Domain(OpenAI)", Proxy: "🇺🇸 US-Silicon-01", UpSpeed: "2.1 KB/s", DownSpeed: "12.8 KB/s", TotalTransfer: "180 KB", Duration: "1m 05s"},
			{ID: "c-4", Host: "gateway.icloud.com:443", Process: "cloudd", Rule: "DomainSuffix(apple.com)", Proxy: "DIRECT", UpSpeed: "0 B/s", DownSpeed: "0 B/s", TotalTransfer: "14 KB", Duration: "12m 40s"},
			{ID: "c-5", Host: "api.telegram.org:443", Process: "Telegram", Rule: "IP-CIDR(Telegram)", Proxy: "🇭🇰 HK-BGP-01", UpSpeed: "14.2 KB/s", DownSpeed: "82.5 KB/s", TotalTransfer: "5.8 MB", Duration: "25m 18s"},
			{ID: "c-6", Host: "registry.npmjs.org:443", Process: "node", Rule: "Match(Proxy)", Proxy: "🇭🇰 HK-BGP-01", UpSpeed: "0 B/s", DownSpeed: "0 B/s", TotalTransfer: "820 KB", Duration: "8m 02s"},
			{ID: "c-7", Host: "v10.events.data.microsoft.com:443", Process: "Code", Rule: "Domain(Microsoft)", Proxy: "DIRECT", UpSpeed: "0.2 KB/s", DownSpeed: "0.1 KB/s", TotalTransfer: "32 KB", Duration: "18m 50s"},
		},

		rules: []RoutingRule{
			{Index: 1, Type: "DOMAIN-SUFFIX", Payload: "google.com", Proxy: "🚀 节点选择"},
			{Index: 2, Type: "DOMAIN-SUFFIX", Payload: "github.com", Proxy: "🚀 节点选择"},
			{Index: 3, Type: "DOMAIN-KEYWORD", Payload: "openai", Proxy: "🤖 AI 专线"},
			{Index: 4, Type: "DOMAIN-KEYWORD", Payload: "telegram", Proxy: "🚀 节点选择"},
			{Index: 5, Type: "IP-CIDR", Payload: "192.168.0.0/16", Proxy: "DIRECT"},
			{Index: 6, Type: "IP-CIDR", Payload: "10.0.0.0/8", Proxy: "DIRECT"},
			{Index: 7, Type: "GEOIP", Payload: "CN", Proxy: "DIRECT"},
			{Index: 8, Type: "DOMAIN-SUFFIX", Payload: "doubleclick.net", Proxy: "🛑 广告拦截"},
			{Index: 9, Type: "MATCH", Payload: "", Proxy: "🐟 漏网之鱼"},
		},

		logs: []LogEntry{
			{Time: "16:38:02", Level: "INFO", Message: "[TCP] 127.0.0.1:54321 --> api.github.com:443 match DomainSuffix(github.com) using 🇭🇰 HK-BGP-01"},
			{Time: "16:38:05", Level: "INFO", Message: "[DNS] resolve host chat.openai.com -> 104.18.2.14 in 24ms"},
			{Time: "16:38:08", Level: "WARN", Message: "[TCP] connection to 185.12.30.5:443 handshake timeout, fallback to HK-BGP-01"},
			{Time: "16:38:12", Level: "INFO", Message: "[Rule] match GeoIP(CN) 114.114.114.114 -> DIRECT"},
			{Time: "16:38:15", Level: "DEBUG", Message: "[TUN] route tun0 packet src=198.18.0.1 dst=1.1.1.1 proto=UDP len=64"},
			{Time: "16:38:20", Level: "INFO", Message: "[Profile] enhanced profile check completed, 48 nodes available"},
		},
	}
}
