# [@ctfbot](https://stacker.news/ctfbot_)

[Stacker News](https://stacker.news) CTF bot.

Any reply to the bot that starts with `$` is run as a shell command inside a
[bubblewrap](https://github.com/containers/bubblewrap) sandbox, and the bot replies with the command
output.

The bot does not keep any state. On startup, all existing notifications are marked as DONE, then we
poll for new reply notifications and process them.

## Sandbox

Each command runs via:

```
bwrap \
  --ro-bind /nix/store /nix/store \        # tools resolve, read-only
  --ro-bind $TOOLS /run \                  # curated PATH (SANDBOX_TOOLS)
  --ro-bind $CHALLENGE /home/ctfbot \      # the puzzle, READ-ONLY (-challenge)
  --tmpfs /tmp --proc /proc --dev /dev \
  --chdir /home/ctfbot \
  --clearenv --setenv PATH /run/bin \      # no host env leaks (e.g. SN_NSEC)
  --unshare-all --die-with-parent --new-session \
  /run/bin/sh -c "ulimit ...; <text after the $>"
```

## Running

The bot answers replies to a single Stacker News item, over a challenge filesystem:

```sh
nix build              # wraps the bot with bwrap, the tool set, and SANDBOX_TOOLS baked in
SN_NSEC=<nsec> ./result/bin/ctfbot -item <id> -challenge <dir>
```

`nix build` bakes the runtime environment (bwrap + the sandbox tool set) into the wrapper. A plain
`go build` binary instead needs `SANDBOX_TOOLS` set and the required commands on PATH itself:

```sh
go build -o ctfbot .
SN_NSEC=<nsec> SANDBOX_TOOLS=<tools-dir> ./ctfbot -item <id> -challenge <dir>
```

`SN_NSEC` (the auth key) and `SANDBOX_TOOLS` are read from the environment; a `.env` file in the
working directory is loaded for `SN_NSEC`. Everything else is a flag with a default — see
`./ctfbot -h`. At startup the bot also checks that `bwrap`, `git`, `grep`, `base64`, and `pdfinfo`
are on PATH and exits if any is missing.

## Configuration

| flag                  | default               | purpose                                                        |
|-----------------------|-----------------------|----------------------------------------------------------------|
| `-item`               | *(required)*          | Stacker News item whose replies the bot answers                |
| `-challenge`          | *(required)*          | challenge filesystem root, mounted read-only in the sandbox    |
| `-sn-base-url`        | `https://stacker.news`| Stacker News API base URL                                      |
| `-poll-interval`      | `5s`                  | how often to poll for new replies                              |
| `-sandbox-timeout`    | `5s`                  | max wall-clock time per sandboxed command                      |
| `-sandbox-max-output` | `4000`                | max bytes of command output before truncation                  |

| env var         | default      | purpose                                                       |
|-----------------|--------------|---------------------------------------------------------------|
| `SN_NSEC`       | *(required)* | Stacker News auth key / nostr nsec; may be set in `.env`      |
| `SANDBOX_TOOLS` | *(required)* | dir whose `bin/` holds the commands mounted into the sandbox  |

### Sandbox tools

`SANDBOX_TOOLS` must point at a directory whose `bin/` holds the commands available to players; the
bot mounts it read-only as the sandbox PATH and refuses to start if it is unset or has no `bin/`.
There is no built-in fallback — the tool set is defined entirely by `flake.nix`.

`flake.nix` builds that directory (`sandboxTools`, a `buildEnv` over `runtimeTools` — coreutils,
grep, sed, awk, `file`, `vim`, `git`, **poppler-utils (`pdfinfo`, `pdftotext`, …)**, …), sets
`SANDBOX_TOOLS` to it, and puts the same tools on the bot's own PATH so the startup command checks
pass. Edit `runtimeTools` to change what the box "has installed".

## Layout

| file / dir      | role                                                        |
|-----------------|-------------------------------------------------------------|
| `main.go`       | flags/env, notification stream, `$` parsing, run each command, reply |
| `sandbox.go`    | the bubblewrap runner (read-only, no network, capped)       |
| `cmd/leet/`     | small CLI that rewrites text in leetspeak (flag authoring)  |
| `flake.nix`     | dev shell + wrapped package (bwrap + `SANDBOX_TOOLS`)        |
