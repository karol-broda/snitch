# snitch

a friendlier `ss` / `netstat` for humans. inspect network connections with a clean tui or styled tables.

![snitch demo](demo/demo.gif)

## install

### homebrew

```bash
brew install snitch
```

> thanks to [@bevanjkay](https://github.com/bevanjkay) for adding snitch to homebrew-core

### go

```bash
go install github.com/karol-broda/snitch@latest
```

### nixpkgs

```bash
nix-env -iA nixpkgs.snitch
```

> thanks to [@DieracDelta](https://github.com/DieracDelta) for adding snitch to nixpkgs

### nixos / nix (flake)

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

### home-manager (flake)

add snitch to your flake inputs and import the home-manager module:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    snitch.url = "github:karol-broda/snitch";
  };

  outputs = { nixpkgs, home-manager, snitch, ... }: {
    homeConfigurations."user" = home-manager.lib.homeManagerConfiguration {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      modules = [
        snitch.homeManagerModules.default
        {
          programs.snitch = {
            enable = true;
            # optional: use the flake's package instead of nixpkgs
            # package = snitch.packages.x86_64-linux.default;
            settings = {
              defaults = {
                theme = "catppuccin-mocha";
                interval = "2s";
                resolve = true;
              };
            };
          };
        }
      ];
    };
  };
}
```

available themes: `ansi`, `catppuccin-mocha`, `catppuccin-macchiato`, `catppuccin-frappe`, `catppuccin-latte`, `gruvbox-dark`, `gruvbox-light`, `dracula`, `nord`, `tokyo-night`, `tokyo-night-storm`, `tokyo-night-light`, `solarized-dark`, `solarized-light`, `one-dark`, `mono`

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

### docker

pre-built oci images available from github container registry:

```bash
# pull from ghcr.io
docker pull ghcr.io/karol-broda/snitch:latest          # alpine (default)
docker pull ghcr.io/karol-broda/snitch:latest-alpine   # alpine (~17MB)
docker pull ghcr.io/karol-broda/snitch:latest-scratch  # minimal, binary only (~9MB)
docker pull ghcr.io/karol-broda/snitch:latest-debian   # debian trixie
docker pull ghcr.io/karol-broda/snitch:latest-ubuntu   # ubuntu 24.04

# or use a specific version
docker pull ghcr.io/karol-broda/snitch:0.2.0-alpine
```

alternatively, build locally via nix flake:

```bash
nix build github:karol-broda/snitch#snitch-alpine
docker load < result
```

**running the container:**

```bash
# basic usage - sees host sockets but not process names
docker run --rm --net=host snitch:latest ls

# full info - includes PID, process name, user
docker run --rm --net=host --pid=host --cap-add=SYS_PTRACE snitch:latest ls
```

| flag | purpose |
|------|---------|
| `--net=host` | share host network namespace (required to see host connections) |
| `--pid=host` | share host pid namespace (needed for process info) |
| `--cap-add=SYS_PTRACE` | read process details from `/proc/<pid>` |

> **note:** `CAP_NET_ADMIN` and `CAP_NET_RAW` are not required. snitch reads from `/proc/net/*` which doesn't need special network capabilities.

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

[tui]
remember_state = false   # remember view options between sessions
```

### remembering view options

when `remember_state = true`, the tui will save and restore:

- filter toggles (tcp/udp, listen/established/other)
- sort field and direction
- address and port resolution settings

state is saved to `$XDG_STATE_HOME/snitch/tui.json` (defaults to `~/.local/state/snitch/tui.json`).

cli flags always take priority over saved state.

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
