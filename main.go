package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	sn "github.com/ekzyis/snappy"
)

type Command struct {
	Reply sn.Item
	Cmd   string
}

func main() {
	loadEnv()

	challenge := flag.String("challenge", "", "challenge filesystem root, mounted read-only in the sandbox")
	rootItemId := flag.Int("item", 0, "Stacker News item whose replies the bot answers")
	snBaseUrl := flag.String("sn-base-url", "https://stacker.news", "Stacker News API base URL")
	pollInterval := flag.Duration("poll-interval", 5*time.Second, "how often to poll for new replies")
	sandboxTimeout := flag.Duration("sandbox-timeout", 5*time.Second, "max wall-clock time per sandboxed command")
	sandboxMaxOutput := flag.Int("sandbox-max-output", 2000, "max bytes of command output before truncation")
	flag.Usage = usage
	flag.Parse()

	if *rootItemId <= 0 {
		flag.Usage()
		log.Fatalf("-item <itemId> is required: the Stacker News item whose replies the bot answers")
	}
	if *challenge == "" {
		flag.Usage()
		log.Fatalf("-challenge <dir> is required: the read-only challenge filesystem root")
	}
	// The auth key is a secret, so it stays in the environment (or .env)
	// rather than a CLI flag that would show up in the process list.
	nsec := os.Getenv("SN_NSEC")
	if nsec == "" {
		flag.Usage()
		log.Fatalf("SN_NSEC is required: the Stacker News auth key (set it in the environment or .env)")
	}
	tools := os.Getenv("SANDBOX_TOOLS")
	if tools == "" {
		flag.Usage()
		log.Fatalf("SANDBOX_TOOLS is required: a directory whose bin/ holds the commands available in the sandbox")
	}

	// Fail fast if the runtime environment is missing tools the bot and its
	// challenges rely on.
	if missing := missingCommands("bwrap", "git", "grep", "base64", "pdfinfo"); len(missing) > 0 {
		log.Fatalf("required commands not found on PATH: %s", strings.Join(missing, ", "))
	}

	sandbox, err := NewSandbox(*challenge, tools, *sandboxTimeout, *sandboxMaxOutput)
	if err != nil {
		log.Fatalf("sandbox init failed: %v", err)
	}
	log.Printf(
		"sandbox ready: bwrap=%s root=%s tools=%s timeout=%s max_output=%d",
		sandbox.Bwrap, sandbox.Root, sandbox.Tools, sandbox.Timeout, sandbox.MaxOutputBytes)

	c := sn.NewClient(
		sn.WithBaseUrl(*snBaseUrl),
		sn.WithNsec(nsec),
	)

	me, err := c.Me()
	if err != nil {
		log.Fatalf("could not authenticate with Stacker News: %v", err)
	}
	log.Printf("logged in as @%s", me.Name)
	log.Printf("watching replies to item #%d", *rootItemId)

	for cmd := range streamCommands(c, *pollInterval, *rootItemId) {
		handleCommand(c, sandbox, cmd)
	}
}

// banner is `figlet -f smslant ctfbot`.
const banner = `      __  _____        __
 ____/ /_/ _/ /  ___  / /_
/ __/ __/ _/ _ \/ _ \/ __/
\__/\__/_//_.__/\___/\__/ `

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintln(out, banner)
	fmt.Fprintln(out, "\nA Stacker News CTF bot.")
	fmt.Fprintf(out, "\nUsage:\n  SN_NSEC=<nsec> %s -item <id> -challenge <dir>\n", os.Args[0])
	fmt.Fprintln(out, "\nOptions:")
	flag.PrintDefaults()
	fmt.Fprintln(out, "\nEnvironment:")
	fmt.Fprintf(out, "  %-14s %s\n", "SN_NSEC", "Stacker News auth key / nostr nsec (required; may be set in .env)")
	fmt.Fprintf(out, "  %-14s %s\n", "SANDBOX_TOOLS", "dir whose bin/ holds the commands available in the sandbox (required)")
}

// missingCommands returns which of the given commands are not found on PATH.
func missingCommands(cmds ...string) []string {
	var missing []string
	for _, c := range cmds {
		if _, err := exec.LookPath(c); err != nil {
			missing = append(missing, c)
		}
	}
	return missing
}

// loadEnv loads KEY=VALUE lines from a .env file (if present) into the process
// environment. It is how the SN_NSEC secret is provided without putting it on
// the command line.
func loadEnv() {
	f, err := os.Open(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatalf("error opening .env: %v", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Fatalf(".env: invalid line: %s", line)
		}
		os.Setenv(parts[0], parts[1])
	}
	if err := s.Err(); err != nil {
		log.Println("error scanning .env:", err)
	}
}

func streamCommands(c *sn.Client, interval time.Duration, rootItemId int) <-chan Command {
	out := make(chan Command)

	go func() {
		defer close(out)

		seen := map[int]bool{}

		if replies, err := c.Replies(); err != nil {
			log.Printf("initial notification fetch failed: %v", err)
		} else {
			for _, n := range replies {
				seen[n.Item.Id] = true
			}
			log.Printf("baselined %d existing replies as handled", len(replies))
		}

		for {
			waitUntilNext(interval)

			replies, err := c.Replies()
			if err != nil {
				log.Printf("poll error: %v", err)
				continue
			}

			for i := len(replies) - 1; i >= 0; i-- {
				r := replies[i].Item
				if seen[r.Id] {
					continue
				}
				seen[r.Id] = true

				if r.ParentId != rootItemId {
					continue
				}

				cmd, ok := parseCommand(r.Text)
				if !ok {
					continue
				}
				out <- Command{Reply: r, Cmd: cmd}
			}
		}
	}()

	return out
}

func parseCommand(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "$") {
		return "", false
	}
	return strings.TrimSpace(trimmed[1:]), true
}

func handleCommand(c *sn.Client, sandbox *Sandbox, cmd Command) {
	log.Printf("running command from reply #%d by @%s: %q", cmd.Reply.Id, cmd.Reply.User.Name, oneLine(cmd.Cmd))

	output := sandbox.Run(cmd.Cmd)

	cId, err := c.CreateComment(cmd.Reply.Id, formatOutput(output))
	if err != nil {
		log.Printf("failed to post output for reply #%d: %v", cmd.Reply.Id, err)
		return
	}
	log.Printf("posted output comment #%d for reply #%d", cId, cmd.Reply.Id)
}

func formatOutput(output string) string {
	// Use a fence longer than any run of backticks in the output so the
	// content can't break out of the code block (CommonMark closing-fence rule),
	// but never shorter than the standard 3.
	fenceLen := maxBacktickRun(output) + 1
	if fenceLen < 3 {
		fenceLen = 3
	}
	fence := strings.Repeat("`", fenceLen)

	var b strings.Builder
	b.WriteString(fence)
	b.WriteString("\n")
	b.WriteString(strings.TrimRight(output, "\n"))
	b.WriteString("\n")
	b.WriteString(fence)
	return b.String()
}

// maxBacktickRun returns the length of the longest run of consecutive
// backticks in s.
func maxBacktickRun(s string) int {
	longest, cur := 0, 0
	for _, r := range s {
		if r == '`' {
			cur++
			if cur > longest {
				longest = cur
			}
		} else {
			cur = 0
		}
	}
	return longest
}

func oneLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i] + " ..."
	}
	return s
}

func waitUntilNext(d time.Duration) {
	now := time.Now()
	dur := now.Truncate(d).Add(d).Sub(now)
	time.Sleep(dur)
}

