// Package core runs the mihomo proxy core inside the application.
//
// mihomo keeps its state in package level variables: one process runs one
// core, with one tunnel, one connection table and one set of listeners.
// Everything that talks to it therefore goes through this package, and the
// rest of the app only sees the snapshots it returns.
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/metacubex/mihomo/config"
	"github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/hub/executor"
	"github.com/metacubex/mihomo/tunnel"
	"github.com/metacubex/mihomo/tunnel/statistic"
)

// The modes the tunnel runs in, as config and the user name them.
const (
	ModeRule   = "rule"
	ModeGlobal = "global"
	ModeDirect = "direct"
)

// ErrStopped is returned by calls that need a running core.
var ErrStopped = errors.New("core is not running")

// Manager owns the process' core: it starts it from a runtime config,
// reloads it when the config changes and stops it on exit.
//
// Only one Manager may run in a process, and its methods are safe for
// concurrent use, but Start and Reload block while the config is parsed
// and the listeners are (re)bound, so call them from a goroutine when the
// UI is up.
type Manager struct {
	mu      sync.Mutex
	dir     string
	running bool
	since   time.Time
}

// New returns a Manager that keeps the core's files, its cache and the
// runtime config it runs, in dir.
func New(dir string) *Manager {
	return &Manager{dir: dir}
}

// Dir is the directory the core keeps its files in.
func (m *Manager) Dir() string { return m.dir }

// Start parses yaml as the runtime config, applies it and starts the
// listeners it names. A config that does not parse leaves the core as it
// was, so a bad edit never takes a working tunnel down.
func (m *Manager) Start(yaml []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return errors.New("core is already running")
	}

	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return fmt.Errorf("prepare %s: %w", m.dir, err)
	}
	// mihomo resolves cache.db, the geo files and its config.yaml against
	// the home directory, which defaults to ~/.config/mihomo. Point it at
	// the app's own directory first, before anything reads or writes them.
	constant.SetHomeDir(m.dir)
	constant.SetConfig(filepath.Join(m.dir, "config.yaml"))
	if err := config.Init(m.dir); err != nil {
		return fmt.Errorf("prepare the core's config: %w", err)
	}

	cfg, err := executor.ParseWithBytes(yaml)
	if err != nil {
		return fmt.Errorf("parse the config: %w", err)
	}
	runtimePath := filepath.Join(m.dir, "runtime.yaml")
	if err := os.WriteFile(runtimePath, yaml, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", runtimePath, err)
	}

	executor.ApplyConfig(cfg, true)

	m.running = true
	m.since = time.Now()
	return nil
}

// Reload applies a new runtime config to the running core: the listeners
// are rebound, the tunnel keeps its connections where it can, and the
// update chain and DNS are replaced. The config that runs stays in place
// until yaml parses.
func (m *Manager) Reload(yaml []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return ErrStopped
	}

	cfg, err := executor.ParseWithBytes(yaml)
	if err != nil {
		return fmt.Errorf("parse the config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(m.dir, "runtime.yaml"), yaml, 0o600); err != nil {
		return fmt.Errorf("write the runtime config: %w", err)
	}

	executor.ApplyConfig(cfg, true)
	return nil
}

// Stop closes the listeners, the tunnel and the connection table, and
// returns once the ports are free, so the same Manager can start again.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}
	// executor.Shutdown closes the TUN device and stores the fake-ip pool,
	// but not the inbound listeners: their ports stay bound until the
	// process exits. Applying a config with no inbounds closes them.
	if cfg, err := executor.ParseWithBytes([]byte(noInboundConfig)); err == nil {
		executor.ApplyConfig(cfg, true)
	}
	executor.Shutdown()
	m.running = false
}

// noInboundConfig is the config Stop applies to release the ports.
const noInboundConfig = `port: 0
socks-port: 0
mixed-port: 0
redir-port: 0
tproxy-port: 0
rules:
  - MATCH,DIRECT
`

// Running reports whether the core runs.
func (m *Manager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// Status is a snapshot of what the core is doing, cheap enough to take
// once a second for the window.
type Status struct {
	Running bool
	// Since is when the core started, zero when it does not run.
	Since time.Time
	// Mode is ModeRule, ModeGlobal or ModeDirect.
	Mode string
	// Up and Down are the bytes the core moved in the last second, the
	// rates a traffic graph shows.
	Up, Down int64
	// UpTotal and DownTotal count everything since the core started.
	UpTotal, DownTotal int64
	// Connections is how many connections are open.
	Connections int
	// Memory is what the core uses, in bytes.
	Memory uint64
}

// Status takes a snapshot of the running core.
func (m *Manager) Status() Status {
	m.mu.Lock()
	st := Status{Running: m.running, Since: m.since}
	m.mu.Unlock()
	if !st.Running {
		return st
	}

	st.Mode = modeName()
	st.Up, st.Down = statistic.DefaultManager.Now()
	st.UpTotal, st.DownTotal = statistic.DefaultManager.Total()
	if snap := statistic.DefaultManager.Snapshot(); snap != nil {
		st.Connections = len(snap.Connections)
		st.Memory = snap.Memory
	}
	return st
}

// SetMode switches between ModeRule, ModeGlobal and ModeDirect, as the
// window's mode switch does. The change lasts until the next Start or
// Reload, which take the mode from the config.
func (m *Manager) SetMode(name string) error {
	if !m.Running() {
		return ErrStopped
	}
	mode, ok := tunnel.ModeMapping[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("unknown mode %q", name)
	}
	tunnel.SetMode(mode)
	return nil
}

// modeName names the mode the tunnel runs in.
func modeName() string {
	current := tunnel.Mode()
	for name, mode := range tunnel.ModeMapping {
		if mode == current {
			return name
		}
	}
	return ""
}
