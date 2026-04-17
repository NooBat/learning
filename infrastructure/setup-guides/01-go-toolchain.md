# 01 — Install the Go toolchain

**Goal:** you can run `go version` and see Go 1.22 or newer.
**Who runs this:** Daniel. Claude does not execute install commands.

## Prerequisites

- macOS (you're on Darwin; this guide is Mac-specific)
- Homebrew installed — check with `brew --version`. If you don't have it: https://brew.sh

## Step 1: Install Go via Homebrew

```bash
brew install go
```

**Why Homebrew (not the official installer):** Homebrew manages upgrades cleanly (`brew upgrade go`) and doesn't require manual PATH fiddling on macOS. The official installer from go.dev works too, but leaves you managing PATH yourself.

## Step 2: Verify installation

```bash
go version
```

Expected: `go version go1.23.x darwin/arm64` (or similar; anything 1.22+ is fine).

If `go: command not found`:
- Make sure Homebrew's bin is on PATH. Add to `~/.zshrc`:
  ```bash
  export PATH="/opt/homebrew/bin:$PATH"
  ```
- Apply: `source ~/.zshrc` or open a new terminal.

## Step 3: Confirm your Go environment

```bash
go env GOROOT GOPATH GOMODCACHE
```

Expected output (approximately):
- `GOROOT` — stdlib location (e.g., `/opt/homebrew/Cellar/go/.../libexec`)
- `GOPATH` — your workspace (default `~/go`)
- `GOMODCACHE` — module cache (default `~/go/pkg/mod`)

You don't need to set any of these for L01. Modern Go (modules) means your project can live anywhere.

## Step 4: Install `gopls` (language server) — recommended

Gives you autocomplete, jump-to-def, linting in VS Code / Neovim / etc.

```bash
go install golang.org/x/tools/gopls@latest
```

If you see `gopls: command not found` after install, add `~/go/bin` to PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

## Step 5: `golangci-lint` — defer to L03

Skip for now. We'll install it at L03 when stricter linting becomes useful.

## Verification

Test with a throwaway Go program:

```bash
mkdir /tmp/gotest && cd /tmp/gotest
go mod init example.com/hello
cat > main.go <<'EOF'
package main

import "fmt"

func main() {
    fmt.Println("hello from go")
}
EOF
go run .
```

Expected output:
```
hello from go
```

Clean up:
```bash
cd ~ && rm -rf /tmp/gotest
```

If that printed "hello from go", Go is installed correctly.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `go: command not found` | PATH missing Homebrew bin | Add `/opt/homebrew/bin` to PATH |
| Go version < 1.22 | Old install | `brew upgrade go` |
| `gopls` not found after install | `~/go/bin` not on PATH | Add `$HOME/go/bin` to PATH |
| Module download hangs / proxy error | Network / firewall | `go env -w GOPROXY=https://proxy.golang.org,direct` (usually not needed) |

## What Claude does NOT do here

- Does not run `brew install` for you.
- Does not modify your `~/.zshrc`.
- Does not run the verification step — you run it and show Claude the output if something goes wrong.

## Next

Once `go version` prints a 1.22+ version, continue to `02-postgres-local.md`.
