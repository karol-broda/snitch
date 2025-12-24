# snitch

a friendlier `ss` / `netstat` for humans. inspect network connections with a clean tui or styled tables.

![snitch demo](demo/demo.gif)

## install

### go

```bash
go install github.com/karol-broda/snitch@latest
```

### nixos / nix

```bash
# try it
nix run github:karol-broda/snitch

# install to profile
nix profile install github:karol-broda/snitch

# or add to flake inputs
{
  inputs.snitch.url = "github:karol-broda/snitch";
}
# then use: inputs.snitch.packages.${system}.default
```

### arch linux (aur)

```bash
# with yay
yay -S snitch-bin

# with paru
paru -S snitch-bin
```

### shell script

```bash
curl -sSL https://raw.githubusercontent.com/karol-broda/snitch/master/install.sh | sh
```

installs to `~/.local/bin` if available, otherwise `/usr/local/bin`. override with:

```bash
curl -sSL https://raw.githubusercontent.com/karol-broda/snitch/master/install.sh | INSTALL_DIR=~/bin sh
```

> **macos:** the install script automatically removes the quarantine attribute (`com.apple.quarantine`) from the binary to allow it to run without gatekeeper warnings. to disable this, set `KEEP_QUARANTINE=1`.

### binary

download from [releases](https://github.com/karol-broda/snitch/releases):

- **linux:** `snitch_<version>_linux_<arch>.tar.gz` or `.deb`/`.rpm`/`.apk`
- **macos:** `snitch_<version>_darwin_<arch>.tar.gz`

```bash
tar xzf snitch_*.tar.gz
sudo mv snitch /usr/local/bin/
```

> **macos:** if blocked with "cannot be opened because the developer cannot be verified", run:
>
> ```bash
> xattr -d com.apple.quarantine /usr/local/bin/snitch
> ```

## quick start

```bash
snitch              # launch interactive tui
snitch -l           # tui showing only listening sockets
snitch ls           # print styled table and exit
snitch ls -l        # listening sockets only
snitch ls -t -e     # tcp established connections
snitch ls -p        # plain output (parsable)
```

## commands

### `snitch` / `snitch top`

interactive tui with live-updating connection list.

```bash
snitch                  # all connections
snitch -l               # listening only
snitch -t               # tcp only
snitch -e               # established only
snitch -i 2s            # 2 second refresh interval
```

**keybindings:**

```
j/k, ↑/↓      navigate
g/G           top/bottom
t/u           toggle tcp/udp
l/e/o         toggle listen/established/other
s/S           cycle sort / reverse
w             watch/monitor process (highlight)
W             clear all watched
K             kill process (with confirmation)
/             search
enter         connection details
?             help
q             quit
```

### `snitch ls`

one-shot table output. uses a pager automatically if output exceeds terminal height.

```bash
snitch ls               # styled table (default)
snitch ls -l            # listening only
snitch ls -t -l         # tcp listeners
snitch ls -e            # established only
snitch ls -p            # plain/parsable output
snitch ls -o json       # json output
snitch ls -o csv        # csv output
snitch ls -n            # numeric (no dns resolution)
snitch ls --no-headers  # omit headers
```

### `snitch json`

json output for scripting.

```bash
snitch json
snitch json -l
```

### `snitch watch`

stream json frames at an interval.

```bash
snitch watch -i 1s | jq '.count'
snitch watch -l -i 500ms
```

### `snitch upgrade`

check for updates and upgrade in-place.

```bash
snitch upgrade              # check for updates
snitch upgrade --yes        # upgrade automatically
snitch upgrade -v 0.1.7     # install specific version
```

## filters

shortcut flags work on all commands:

```
-t, --tcp           tcp only
-u, --udp           udp only
-l, --listen        listening sockets
-e, --established   established connections
-4, --ipv4          ipv4 only
-6, --ipv6          ipv6 only
```

## resolution

dns and service name resolution options:

```
--resolve-addrs     resolve ip addresses to hostnames (default: true)
--resolve-ports     resolve port numbers to service names
--no-cache          disable dns caching (force fresh lookups)
```

dns lookups are performed in parallel and cached for performance. use `--no-cache` to bypass the cache for debugging or when addresses change frequently.

for more specific filtering, use `key=value` syntax with `ls`:

```bash
snitch ls proto=tcp state=listen
snitch ls pid=1234
snitch ls proc=nginx
snitch ls lport=443
snitch ls contains=google
```

## output

styled table (default):

```
  ╭─────────────────┬───────┬───────┬─────────────┬─────────────────┬────────╮
  │ PROCESS         │ PID   │ PROTO │ STATE       │ LADDR           │ LPORT  │
  ├─────────────────┼───────┼───────┼─────────────┼─────────────────┼────────┤
  │ nginx           │ 1234  │ tcp   │ LISTEN      │ *               │ 80     │
  │ postgres        │ 5678  │ tcp   │ LISTEN      │ 127.0.0.1       │ 5432   │
  ╰─────────────────┴───────┴───────┴─────────────┴─────────────────┴────────╯
  2 connections
```

plain output (`-p`):

```
PROCESS    PID    PROTO   STATE    LADDR       LPORT
nginx      1234   tcp     LISTEN   *           80
postgres   5678   tcp     LISTEN   127.0.0.1   5432
```

## configuration

optional config file at `~/.config/snitch/snitch.toml`:

```toml
[defaults]
numeric = false      # disable name resolution
dns_cache = true     # cache dns lookups (set to false to disable)
theme = "auto"       # color theme: auto, dark, light, mono
```

### environment variables

```bash
SNITCH_THEME=dark          # set default theme
SNITCH_RESOLVE=0           # disable dns resolution
SNITCH_DNS_CACHE=0         # disable dns caching
SNITCH_NO_COLOR=1          # disable color output
SNITCH_CONFIG=/path/to     # custom config file path
```

## requirements

- linux or macos
- linux: reads from `/proc/net/*`, root or `CAP_NET_ADMIN` for full process info
- macos: uses system APIs, may require sudo for full process info
