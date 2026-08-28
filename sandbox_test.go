package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testRoot = "challenge/test"

func newTestSandbox(t *testing.T) *Sandbox {
	t.Helper()
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not available (e.g. inside the Nix build sandbox); skipping integration test")
	}
	tools, err := defaultTools()
	if err != nil {
		t.Skipf("no sandbox tools on host: %v", err)
	}
	seedTestRoot(t)
	return &Sandbox{
		SandboxRoot:    testRoot,
		Tools:          tools,
		Timeout:        5 * time.Second,
		MaxOutputBytes: 4000,
	}
}

func seedTestRoot(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(testRoot, 0o755); err != nil {
		t.Fatalf("could not create test root: %v", err)
	}
	files := map[string]string{
		"hello.txt": "hello from the sandbox\n",
		".secret":   "hidden treasure\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(testRoot, name), []byte(content), 0o644); err != nil {
			t.Fatalf("could not seed %s: %v", name, err)
		}
	}
}

func TestSandboxListsChallenge(t *testing.T) {
	s := newTestSandbox(t)
	out, _ := s.Run("ls")
	if !strings.Contains(out, "hello.txt") {
		t.Fatalf("expected hello.txt in `ls`, got:\n%s", out)
	}
}

func TestSandboxHiddenFileChain(t *testing.T) {
	s := newTestSandbox(t)
	if out, _ := s.Run("ls -a"); !strings.Contains(out, ".secret") {
		t.Fatalf("expected .secret in `ls -a`, got:\n%s", out)
	}
	out, _ := s.Run("cat .secret | head -n1")
	if !strings.Contains(out, "hidden treasure") {
		t.Fatalf("expected hidden file contents, got:\n%s", out)
	}
}

func TestSandboxReadOnly(t *testing.T) {
	s := newTestSandbox(t)
	out, _ := s.Run("rm hello.txt 2>&1; echo done; ls hello.txt")
	if !strings.Contains(out, "hello.txt") {
		t.Fatalf("hello.txt should survive a rm attempt, got:\n%s", out)
	}
}

func TestSandboxTmpWritable(t *testing.T) {
	s := newTestSandbox(t)
	out, _ := s.Run("echo hi > /tmp/x && cat /tmp/x")
	if !strings.Contains(out, "hi") {
		t.Fatalf("/tmp should be writable, got:\n%s", out)
	}
}

func TestSandboxNoNetwork(t *testing.T) {
	s := newTestSandbox(t)
	out, _ := s.Run("cat /proc/net/dev")
	if !strings.Contains(out, "lo:") {
		t.Fatalf("expected loopback in /proc/net/dev, got:\n%s", out)
	}
	for _, iface := range []string{"eth", "wlan", "en", "wg", "docker"} {
		if strings.Contains(out, iface) {
			t.Fatalf("unexpected network interface %q inside sandbox:\n%s", iface, out)
		}
	}
}

func TestSandboxTimeout(t *testing.T) {
	s := newTestSandbox(t)
	s.Timeout = 1 * time.Second
	start := time.Now()
	out, _ := s.Run("sleep 10")
	if elapsed := time.Since(start); elapsed > 4*time.Second {
		t.Fatalf("timeout not enforced, took %s", elapsed)
	}
	if !strings.Contains(out, "timed out") {
		t.Fatalf("expected timeout notice, got:\n%s", out)
	}
}

func TestSandboxOutputCap(t *testing.T) {
	s := newTestSandbox(t)
	s.MaxOutputBytes = 100
	out, truncated := s.Run("yes ABCDEFGH | head -n 1000")
	if !truncated {
		t.Fatalf("expected truncation")
	}
	if len(out) > 100 {
		t.Fatalf("output not capped: %d bytes", len(out))
	}
}
