# Installation Guide

## Pre-built Binaries

Download the latest release from the [Releases page](https://github.com/Osir-Inc/cli/releases).

| Platform | Architecture | Binary |
|----------|-------------|--------|
| Linux | x86_64 | `osir-linux-amd64` |
| Linux | ARM64 (Raspberry Pi, Graviton) | `osir-linux-arm64` |
| macOS | Intel | `osir-darwin-amd64` |
| macOS | Apple Silicon (M1/M2/M3/M4) | `osir-darwin-arm64` |
| Windows | x86_64 | `osir-windows-amd64.exe` |

Each release includes a `checksums.txt` file for integrity verification.

### Linux / macOS

```bash
# Download the latest release (replace with your platform and version)
curl -L -o osir https://github.com/Osir-Inc/cli/releases/download/v1.0.0/osir-linux-amd64
chmod +x osir
sudo mv osir /usr/local/bin/osir

# Verify
osir --version
# osir 1.0.0
```

If you don't have root access, install to your home directory:

```bash
mkdir -p ~/bin
mv osir-linux-amd64 ~/bin/osir
chmod +x ~/bin/osir
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Windows

1. Download `osir-windows-amd64.exe`
2. Rename it to `osir.exe`
3. Move it to a directory in your PATH (e.g., `C:\Users\<you>\bin\`)
4. Or add the directory to your PATH environment variable

From PowerShell:

```powershell
# Add to user PATH (one-time)
$binDir = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Path $binDir -Force
Move-Item osir-windows-amd64.exe "$binDir\osir.exe"
[Environment]::SetEnvironmentVariable("Path", "$binDir;" + [Environment]::GetEnvironmentVariable("Path", "User"), "User")
```

## Building from Source

Requires **Go 1.23** or later.

```bash
git clone https://github.com/Osir-Inc/cli.git osir-cli
cd osir-cli

# Build for current platform
make build
# or: go build -o osir .

# Run tests
make test
# or: go test ./...

# Cross-compile for all 5 platforms
make build-all
# Outputs to dist/:
#   dist/osir-linux-amd64
#   dist/osir-linux-arm64
#   dist/osir-darwin-amd64
#   dist/osir-darwin-arm64
#   dist/osir-windows-amd64.exe
```

### Build for a specific platform

```bash
GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w -X main.version=1.0.0" -o osir-linux-amd64 .
GOOS=linux   GOARCH=arm64 go build -ldflags "-s -w -X main.version=1.0.0" -o osir-linux-arm64 .
GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w -X main.version=1.0.0" -o osir-darwin-amd64 .
GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w -X main.version=1.0.0" -o osir-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=1.0.0" -o osir-windows-amd64.exe .
```

The `-s -w` flags strip debug symbols, reducing binary size by ~30%.

## Deploying to a Remote Server

The CLI is designed for headless servers where only SSH access is available.

```bash
# 1. Upload the binary
scp dist/osir-linux-amd64 user@server:/usr/local/bin/osir

# 2. Make executable
ssh user@server chmod +x /usr/local/bin/osir

# 3. Verify
ssh user@server osir --version

# 4. Authenticate using device flow (no browser needed on the server)
ssh user@server osir auth login --device
# Opens a URL you can visit from any device with a browser
```

### Deploy to multiple servers

```bash
for server in web01 web02 db01 cache01; do
  scp dist/osir-linux-amd64 user@$server:/usr/local/bin/osir
  ssh user@$server chmod +x /usr/local/bin/osir
done
```

## Updating

Replace the binary. Credentials are preserved across updates (stored in `~/.osir/credentials.json`).

```bash
# Remote server
scp dist/osir-linux-amd64 user@server:/usr/local/bin/osir

# Local
sudo mv osir-linux-amd64 /usr/local/bin/osir
```

## Shell Completion Setup

Enable tab completion for your shell. This is a one-time setup.

**Bash** (add to `~/.bashrc`):

```bash
source <(osir completion bash)
```

**Zsh** (add to `~/.zshrc`):

```bash
source <(osir completion zsh)
```

**Fish:**

```bash
osir completion fish | source
```

**PowerShell** (add to `$PROFILE`):

```powershell
osir completion powershell | Out-String | Invoke-Expression
```

## Uninstalling

```bash
# Remove the binary
sudo rm /usr/local/bin/osir

# Remove credentials and history (optional)
rm -rf ~/.osir/
```
