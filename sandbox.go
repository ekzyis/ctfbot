package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Sandbox struct {
	Bwrap          string
	SandboxRoot    string
	Tools          string
	Timeout        time.Duration
	MaxOutputBytes int
}

func NewSandbox(root string) (*Sandbox, error) {
	bwrap, err := exec.LookPath("bwrap")
	if err != nil {
		return nil, fmt.Errorf("bwrap not found on PATH: %w", err)
	}

	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("challenge root %q is not a directory: %v", root, err)
	}

	tools := os.Getenv("SANDBOX_TOOLS")
	if tools == "" {
		var err error
		if tools, err = defaultTools(); err != nil {
			return nil, err
		}
	}
	if fi, err := os.Stat(tools + "/bin"); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("SANDBOX_TOOLS %q has no bin/ dir: %v", tools, err)
	}

	return &Sandbox{
		Bwrap:          bwrap,
		SandboxRoot:    root,
		Tools:          tools,
		Timeout:        durEnv("SANDBOX_TIMEOUT", 5*time.Second),
		MaxOutputBytes: intEnv("SANDBOX_MAX_OUTPUT", 4000),
	}, nil
}

func (s *Sandbox) Run(cmd string) (output string, truncated bool) {
	if strings.TrimSpace(cmd) == "" {
		return "(no command)", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	script := "ulimit -t 10 -v 524288 -f 4096 -u 128 2>/dev/null; " + cmd

	args := []string{
		"--ro-bind", "/nix/store", "/nix/store",
		"--ro-bind", s.Tools, "/run",
		"--ro-bind", s.SandboxRoot, "/home/ctfbot",
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
	var buf bytes.Buffer
	lw := &limitedWriter{buf: &buf, limit: s.MaxOutputBytes + 1}
	c.Stdout = lw
	c.Stderr = lw
	c.Stdin = bytes.NewReader(nil)

	err := c.Run()

	out := buf.String()
	if len(out) > s.MaxOutputBytes {
		out = out[:s.MaxOutputBytes]
		truncated = true
	}

	if ctx.Err() == context.DeadlineExceeded {
		out = strings.TrimRight(out, "\n")
		if out != "" {
			out += "\n"
		}
		out += fmt.Sprintf("(timed out after %s)", s.Timeout)
		return out, truncated
	}
	if err != nil && out == "" {
		out = fmt.Sprintf("(command exited with error: %v)", err)
	}
	return out, truncated
}

type limitedWriter struct {
	buf   *bytes.Buffer
	limit int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if remaining := w.limit - w.buf.Len(); remaining > 0 {
		if remaining < len(p) {
			w.buf.Write(p[:remaining])
		} else {
			w.buf.Write(p)
		}
	}
	return len(p), nil
}

func defaultTools() (string, error) {
	dir, err := os.MkdirTemp("", "ppc-tools-")
	if err != nil {
		return "", err
	}
	bin := dir + "/bin"
	if err := os.Mkdir(bin, 0o755); err != nil {
		return "", err
	}
	needed := []string{
		"sh", "bash", "env", "ls", "cat", "echo", "grep", "find", "head",
		"tail", "wc", "sort", "uniq", "cut", "tr", "sed", "awk", "base64",
		"xxd", "file", "strings", "pwd", "whoami", "id", "date", "uname",
		"sleep", "yes", "seq", "printf", "stat", "readlink", "basename",
		"dirname", "md5sum", "sha256sum", "tac", "rev", "nl", "du", "tee",
	}
	found := 0
	for _, name := range needed {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if real, err := filepath.EvalSymlinks(path); err == nil {
			path = real
		}
		if err := os.Symlink(path, bin+"/"+name); err == nil {
			found++
		}
	}
	if found == 0 {
		return "", fmt.Errorf("no sandbox tools found on host PATH")
	}
	return dir, nil
}

var _ io.Writer = (*limitedWriter)(nil)
