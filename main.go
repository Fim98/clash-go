package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"clash-go/internal/core"
)

// mixedPort is the port the built-in config listens on. Profiles will set
// it from the settings later.
const mixedPort = 7890

// defaultConfig is what the core runs while the app has no profiles: a
// mixed proxy that sends every connection out directly, so the window can
// be used before subscriptions exist.
const defaultConfig = `mixed-port: 7890
allow-lan: false
mode: rule
log-level: warning
rules:
  - MATCH,DIRECT
`

// historyLen is how many one-second samples the traffic graph keeps.
const historyLen = 60

// modes are the tunnel modes, in the order the switch shows them.
var modes = []struct{ mode, label string }{
	{core.ModeRule, "规则"},
	{core.ModeGlobal, "全局"},
	{core.ModeDirect, "直连"},
}

func (a *app) stateLabel() string {
	switch {
	case a.busy:
		return "启动中"
	case a.status.Running:
		return "运行中 · " + uptime(a.status.Since)
	default:
		return "未运行"
	}
}

func (a *app) stateColor(t *ui.Theme) ui.Color {
	switch {
	case a.busy:
		return t.Warning
	case a.status.Running:
		return t.Success
	default:
		return t.TextMuted
	}
}

// start brings the core up off the main thread: parsing the config and
// binding the port take a few milliseconds, which the window does not
// wait for.
func (a *app) start() {
	if a.busy {
		return
	}
	a.busy = true
	a.err = ""
	go func() {
		err := a.mgr.Start([]byte(defaultConfig))
		mygo.RunOnMain(func() {
			a.busy = false
			if err != nil {
				a.err = err.Error()
			}
			a.status = a.mgr.Status()
			a.invalidate()
		})
	}()
}

func (a *app) stop() {
	a.mgr.Stop()
	a.err = ""
	a.status = a.mgr.Status()
}

func (a *app) setMode(mode string) {
	if err := a.mgr.SetMode(mode); err != nil {
		if errors.Is(err, core.ErrStopped) {
			a.err = "核心还没有启动"
		} else {
			a.err = err.Error()
		}
		return
	}
	a.err = ""
	a.status = a.mgr.Status()
}

func (a *app) invalidate() {
	if a.win != nil {
		a.win.Invalidate()
	}
}

// poll refreshes the window once a second while the app runs: it takes a
// snapshot of the core, feeds the traffic graph and keeps the tunnel in
// the mode the user picked.
func (a *app) poll() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		status := a.mgr.Status()
		mode := modes[a.outboundMode].mode
		mygo.RunOnMain(func() {
			if status.Running && status.Mode != mode {
				_ = a.mgr.SetMode(mode)
				status = a.mgr.Status()
			}
			a.status = status
			a.pushTraffic(status.Up, status.Down)
			a.invalidate()
		})
	}
}

// pushTraffic appends a sample to the graph's history, dropping the oldest.
func (a *app) pushTraffic(up, down int64) {
	a.trafficHistoryUp = append(a.trafficHistoryUp[1:], float32(up))
	a.trafficHistoryDown = append(a.trafficHistoryDown[1:], float32(down))
}

// chartMax is the top of the traffic graph: the largest sample, with a floor
// so an idle connection still draws a readable axis.
func chartMax(histories ...[]float32) float64 {
	max := float32(1024)
	for _, h := range histories {
		for _, v := range h {
			if v > max {
				max = v
			}
		}
	}
	return float64(max)
}

func uptime(since time.Time) string {
	if since.IsZero() {
		return ""
	}
	d := time.Since(since).Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%d 秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d 分", int(d.Minutes()))
	}
	return fmt.Sprintf("%d 小时 %d 分", int(d.Hours()), int(d.Minutes())%60)
}

func rate(bps int64) string {
	return bytes(bps) + "/s"
}

// bytes formats a size with the unit that keeps it short.
func bytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	for _, suffix := range []string{"KB", "MB", "GB", "TB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f PB", value/unit)
}

func main() {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Fatalf("clash-go: %v", err)
	}

	mgr := core.New(dir)
	a := newApp(mgr)

	mygo.App.WhenReady(func() {
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:         "clash-go",
			Width:         1060,
			Height:        700,
			MinWidth:      880,
			MinHeight:     580,
			StateKey:      "clash-go-v4",
			TitleBarStyle: mygo.TitleBarHiddenInset,
			Content:       ui.View(a.view),
		})
		go a.poll()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatalf("clash-go: %v", err)
	}

	// The listeners and the tunnel go away with the window.
	a.mgr.Stop()
}
