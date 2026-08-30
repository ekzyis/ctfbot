# [@ctfbot_](https://stacker.news/ctfbot_)

[Stacker News](https://stacker.news) CTF bot.

## Past CTFs

> _Talk to me like you do to your terminal~~~_
>
> _in 15 minutes_

-- https://stacker.news/items/1557680

## Bot

Any reply to the bot that starts with `$` is run as a shell command inside a
[bubblewrap](https://github.com/containers/bubblewrap) read-only sandbox, and the bot replies with
the command output.

The bot does not keep any state. On startup, all existing notifications are marked as DONE, then we
poll for new replies with new commands to run in the sandbox.

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

The tools available in the sandbox are defined by `sandboxTools` in the nix flake.

## Running

The bot answers replies to a single item, over a challenge filesystem:

```sh
nix build
# SN_NSEC is added outside of build
echo "SN_NSEC=<nsec>" > .env
# run bot wrapped by nix with bwrap in PATH and SANDBOX_TOOLS in env
./result/bin/ctfbot -item <id> -challenge <dir>
```

It is NOT recommended to run the bot without nix. If you decide to do so, you will need to provide
SANDBOX_TOOLS and `bwrap` on PATH yourself.

The bot checks on startup if the commands necessary to solve the challenges are available.

## Configuration

| flag                  | default               | purpose                                                        |
|-----------------------|-----------------------|----------------------------------------------------------------|
| `-item`               | *(required)*          | Stacker News item whose replies the bot answers                |
| `-challenge`          | *(required)*          | challenge filesystem root, mounted read-only in the sandbox    |
| `-sn-base-url`        | `https://stacker.news`| Stacker News API base URL                                      |
| `-poll-interval`      | `5s`                  | how often to poll for new replies                              |
| `-sandbox-timeout`    | `5s`                  | max wall-clock time per sandboxed command                      |
| `-sandbox-max-output` | `2000`                | max bytes of command output before truncation                  |

| env var         | default      | purpose                                                       |
|-----------------|--------------|---------------------------------------------------------------|
| `SN_NSEC`       | *(required)* | Stacker News auth key / nostr nsec; may be set in `.env`      |
| `SANDBOX_TOOLS` | *(required)* | dir whose `bin/` holds the commands mounted into the sandbox  |

## Layout

| file / dir      | role                                                           |
|-----------------|----------------------------------------------------------------|
| `main.go`       | bot loop: stream commands to sandbox and reply with the output |
| `sandbox.go`    | the bubblewrap runner (read-only, no network, capped)          |
| `cmd/leet/`     | small CLI useful to create leetspeak flags                     |
| `flake.nix`     | dev shell + wrapped package (bwrap + `SANDBOX_TOOLS`)          |
