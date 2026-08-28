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
  --ro-bind /nix/store /nix/store \      # tools resolve, read-only
  --ro-bind $SANDBOX_TOOLS /run \        # curated PATH (bash, coreutils, ...)
  --ro-bind $CHALLENGE /challenge \      # the puzzle, READ-ONLY (-challenge)
  --tmpfs /tmp --proc /proc --dev /dev \
  --chdir /challenge \
  --unshare-all --die-with-parent --new-session \
  /run/bin/sh -c "ulimit ...; <text after the $>"
```

## Running

The bot answers replies to a single Stacker News item, over a challenge filesystem:

```sh
go build -o ctfbot
./ctfbot -item <id> -challenge <dir>
```

## Layout

| file          | role                                                        |
|---------------|-------------------------------------------------------------|
| `main.go`     | notification stream, `$` parsing, run each command, reply   |
| `sandbox.go`  | the bubblewrap runner (read-only, no network, capped)       |
| `flake.nix`   | dev shell + wrapped package (bwrap + `SANDBOX_TOOLS`)       |
