package core

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	mlog "github.com/metacubex/mihomo/log"
)

func TestMain(m *testing.M) {
	// The core logs every connection at info level; tests only care about
	// failures.
	mlog.SetLevel(mlog.ERROR)
	os.Exit(m.Run())
}

// config is a runtime config with everything the tests need: a mixed-port
// listener, a rule that sends every connection out directly and no DNS
// server of its own.
func runtimeConfig(port int) []byte {
	return []byte(fmt.Sprintf(`mixed-port: %d
mode: rule
log-level: warning
rules:
  - MATCH,DIRECT
`, port))
}

// freePort returns a port that is free right now.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// proxyClient returns a client that sends every request through the mixed
// port.
func proxyClient(port int) *http.Client {
	u := &url.URL{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", port)}
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(u)},
		Timeout:   10 * time.Second,
	}
}

func TestStartStatusStop(t *testing.T) {
	port := freePort(t)
	m := New(t.TempDir())

	if m.Running() {
		t.Fatal("a new manager runs")
	}
	if err := m.Start(runtimeConfig(port)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Stop)

	if !m.Running() {
		t.Fatal("the manager does not run after Start")
	}
	if st := m.Status(); st.Mode != ModeRule {
		t.Errorf("mode %q, want %q", st.Mode, ModeRule)
	}
	if st := m.Status(); st.Since.IsZero() {
		t.Error("Since is zero")
	}

	// The listener binds and the tunnel forwards a request.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	body := func() string {
		resp, err := proxyClient(port).Get(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}()
	if body != "hello" {
		t.Errorf("body %q, want hello", body)
	}

	// The listeners close and the port is free again.
	m.Stop()
	if m.Running() {
		t.Error("the manager still runs after Stop")
	}
	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second); err == nil {
		conn.Close()
		t.Error("the port is still listening after Stop")
	}
}

func TestSetMode(t *testing.T) {
	m := New(t.TempDir())
	if err := m.SetMode(ModeGlobal); err != ErrStopped {
		t.Errorf("SetMode before Start: %v, want ErrStopped", err)
	}

	if err := m.Start(runtimeConfig(freePort(t))); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Stop)

	if err := m.SetMode(ModeGlobal); err != nil {
		t.Fatal(err)
	}
	if st := m.Status(); st.Mode != ModeGlobal {
		t.Errorf("mode %q, want %q", st.Mode, ModeGlobal)
	}
	if err := m.SetMode("sideways"); err == nil {
		t.Error("SetMode accepted an unknown mode")
	}
}

// A config that does not parse leaves the running core, and its port,
// alone.
func TestReloadKeepsTheRunningConfigOnError(t *testing.T) {
	port := freePort(t)
	m := New(t.TempDir())
	if err := m.Start(runtimeConfig(port)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Stop)

	if err := m.Reload([]byte("mixed-port: [oops")); err == nil {
		t.Fatal("Reload accepted a broken config")
	}
	if !m.Running() {
		t.Fatal("the manager stopped after a broken reload")
	}
	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second); err != nil {
		t.Errorf("the old port stopped listening: %v", err)
	} else {
		conn.Close()
	}

	// A good config moves the listener.
	newPort := freePort(t)
	if err := m.Reload(runtimeConfig(newPort)); err != nil {
		t.Fatal(err)
	}
	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", newPort), time.Second); err != nil {
		t.Errorf("the new port does not listen: %v", err)
	} else {
		conn.Close()
	}
}

// The core keeps its files in the manager's directory, not in the user's
// mihomo config.
func TestManagerOwnsItsDirectory(t *testing.T) {
	dir := t.TempDir()
	m := New(dir)
	if err := m.Start(runtimeConfig(freePort(t))); err != nil {
		t.Fatal(err)
	}
	defer m.Stop()

	for _, name := range []string{"config.yaml", "runtime.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestVersion(t *testing.T) {
	if v := Version(); v == "" || v == "unknown" {
		t.Fatalf("Version() = %q, want the mihomo module version", v)
	}
}
