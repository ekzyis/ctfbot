package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Sandbox struct {
	Bwrap          string
	Root           string
	Tools          string
	Timeout        time.Duration
	MaxOutputBytes int
}

func NewSandbox(root, tools string, timeout time.Duration, maxOutput int) (*Sandbox, error) {
	bwrap, err := exec.LookPath("bwrap")
	if err != nil {
		return nil, fmt.Errorf("bwrap not found on PATH: %w", err)
	}

	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("challenge root %q is not a directory: %v", root, err)
	}

	if fi, err := os.Stat(tools + "/bin"); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("SANDBOX_TOOLS %q has no bin/ dir: %v", tools, err)
	}

	return &Sandbox{
		Bwrap:          bwrap,
		Root:           root,
		Tools:          tools,
		Timeout:        timeout,
		MaxOutputBytes: maxOutput,
	}, nil
}

func (s *Sandbox) Run(cmd string) (output string) {
	if strings.TrimSpace(cmd) == "" {
		return "(no command)"
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	script := "ulimit -t 10 -v 524288 -f 4096 -u 128 2>/dev/null; " + cmd

	args := []string{
		"--ro-bind", "/nix/store", "/nix/store",
		"--ro-bind", s.Tools, "/run",
		"--ro-bind", s.Root, "/home/ctfbot",
		"--symlink", "/run/bin/sh", "/bin/sh",
		"--symlink", "/run/bin/env", "/usr/bin/env",
		"--tmpfs", "/tmp",
		"--proc", "/proc",
		"--dev", "/dev",
		"--chdir", "/home/ctfbot",
		"--clearenv", // very important to not leak SN_NSEC
		"--setenv", "PATH", "/run/bin",
		"--setenv", "HOME", "/home/ctfbot",
		"--setenv", "TERM", "xterm",
		"--setenv", "PS1", "$ ",
		"--unshare-all",
		"--die-with-parent",
		"--new-session",
		"/run/bin/sh", "-c", script,
	}

	bwrap := s.Bwrap
	if bwrap == "" {
		bwrap = "bwrap"
	}
	c := exec.CommandContext(ctx, bwrap, args...)
	lw := &limitedWriter{limit: s.MaxOutputBytes}
	c.Stdout = lw
	c.Stderr = lw
	c.Stdin = bytes.NewReader(nil)

	err := c.Run()

	out := lw.String()

	if ctx.Err() == context.DeadlineExceeded {
		out = strings.TrimRight(out, "\n")
		if out != "" {
			out += "\n"
		}
		out += fmt.Sprintf("(timed out after %s)", s.Timeout)
		return out
	}
	if err != nil && out == "" {
		out = fmt.Sprintf("(command exited with error: %v)", err)
	}
	return out
}

type limitedWriter struct {
	buf   bytes.Buffer
	limit int
	seen  int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if remaining := w.limit - w.buf.Len(); remaining > 0 {
		if remaining < len(p) {
			w.buf.Write(p[:remaining])
		} else {
			w.buf.Write(p)
		}
	}
	w.seen += len(p)
	return len(p), nil
}

func (w *limitedWriter) String() string {
	if left := w.seen - w.buf.Len(); left > 0 {
		return w.buf.String() + fmt.Sprintf("\n[truncated with %d bytes left]", left)
	}
	return w.buf.String()
}

var _ io.Writer = (*limitedWriter)(nil)
