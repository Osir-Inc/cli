# OSIR CLI

A command-line tool for the [OSIR](https://osir.com) domain registrar platform. Single static binary, zero dependencies, cross-platform.

Manage domains, DNS records, billing, contacts, and more -- from your terminal.

```
$ osir shell
OSIR Interactive Shell v1.0.1
Type 'help' for commands, Tab or '?' for completions, 'exit' to quit.

osir> domain check coolstartup.io
Domain:    coolstartup.io
[OK] Available

osir> domain register coolstartup.io --years 2 --privacy --auto-renew
[OK] Domain registered: coolstartup.io

osir> dns create coolstartup.io A coolstartup.io 192.0.2.1
[OK] DNS record created (ID: coolstartup_io__A_123456)
```

## Features

- **76 commands** across 11 command groups -- domains, DNS, VPS hosting, billing, contacts, audit, accounts, catalog, suggestions, and more
- **Interactive shell** (`osir shell`) -- Junos/Arista-style REPL with Tab completion, `?` help, and persistent command history
- **Single binary** -- no runtime, no dependencies, just copy and run
- **Cross-platform** -- Linux (amd64/arm64), macOS (Intel/Apple Silicon), Windows
- **Two auth methods** -- OAuth 2.0 device flow (headless-safe) and username/password
- **JSON output** -- every command supports `-o json` for scripting and automation
- **Resilient** -- automatic retry with exponential backoff on 5xx errors, configurable timeout
- **Proxy-aware** -- respects HTTP_PROXY/HTTPS_PROXY environment variables
- **Shell completion** -- bash, zsh, fish, PowerShell

## Quick Start

### 1. Install

**Download a pre-built binary** from the [latest release](https://github.com/Osir-Inc/cli/releases):

| Platform | Binary |
|----------|--------|
| Linux x86_64 | `osir-linux-amd64` |
| Linux ARM64 | `osir-linux-arm64` |
| macOS Intel | `osir-darwin-amd64` |
| macOS Apple Silicon | `osir-darwin-arm64` |
| Windows x86_64 | `osir-windows-amd64.exe` |

```bash
# Download, install, verify
curl -L -o osir https://github.com/Osir-Inc/cli/releases/download/v1.0.1/osir-linux-amd64
chmod +x osir
sudo mv osir /usr/local/bin/osir
osir --version
```

**Or build from source** (requires Go 1.23+):

```bash
git clone https://github.com/Osir-Inc/cli.git osir-cli
cd osir-cli
make build        # builds ./osir for current platform
make build-all    # cross-compiles to dist/ for all 5 platforms
```

### 2. Authenticate

```bash
# Browser-based (works on headless servers too)
osir auth login --device

# Or username/password
osir auth login -u your-username
```

### 3. Use it

```bash
osir domain check example.com          # check availability
osir domain list                       # list your domains
osir dns list example.com              # view DNS records
osir vps list                          # list your VPS instances
osir billing balance                   # check account balance
```

Or launch the **interactive shell** for a guided experience:

```bash
osir shell
```

## Documentation

| Document | Description |
|----------|-------------|
| [Installation Guide](docs/installation.md) | All installation methods, deploying to servers, updating |
| [Getting Started](docs/getting-started.md) | First-time setup walkthrough with examples |
| [Interactive Shell](docs/interactive-shell.md) | Using the Junos-style interactive shell mode |
| [Command Reference](docs/command-reference.md) | Complete reference for all 76 commands |
| [Configuration](docs/configuration.md) | Environment variables, credentials, multi-environment setup |
| [Scripting & Automation](docs/scripting.md) | JSON output, batch operations, cron jobs, CI/CD |

## Command Groups

| Group | Commands | Description |
|-------|----------|-------------|
| `auth` | 3 | Login, logout, check status |
| `domain` | 12 | Check, register, renew, lock, privacy, nameservers |
| `dns` | 11 | List, get, create, update, delete, zone-init, zone-exists, fix-soa, dnssec-status, dnssec-enable, dnssec-disable |
| `billing` | 12 | Balance, invoices, payments, pricing |
| `contact` | 6 | Create, update, delete registrant contacts |
| `vps` | 10 | Browse packages, order, manage instances |
| `audit` | 3 | Recent activity, domain audit, failures |
| `account` | 2 | Profile and account summary |
| `catalog` | 2 | Browse TLDs, servers |
| `suggest` | 7 | AI suggestions, word spinning, prefix/suffix |
| `shell` | 1 | Launch interactive shell |
| `completion` | 1 | Generate shell completion scripts |

## Project Structure

```
com.osir.cli/
├── cmd/                     # Cobra commands (11 command groups)
│   ├── root.go              # Command tree factory, App DI
│   ├── shell.go             # Interactive shell (reeflective/console)
│   ├── auth.go              # Authentication commands
│   ├── domain.go            # Domain management (12 subcommands)
│   ├── dns.go               # DNS record management (11 subcommands)
│   ├── billing.go           # Billing and payments
│   ├── contact.go           # Contact management
│   ├── audit.go             # Audit logs
│   ├── account.go           # Account management
│   ├── catalog.go           # Product catalog
│   ├── suggest.go           # Domain name suggestions
│   ├── vps.go               # VPS hosting management (10 subcommands)
│   └── completion.go        # Shell completion scripts
├── internal/
│   ├── api/                 # Backend interface + HTTP client
│   │   └── models/          # Request/response structs
│   ├── auth/                # OAuth session, credentials, device flow, SessionManager interface
│   ├── config/              # Environment-based configuration
│   └── output/              # Text/JSON formatter
├── main.go                  # Entry point
├── go.mod                   # Go module (github.com/osir/cli)
├── Makefile                 # build, build-all, test, clean
└── docs/                    # Documentation
```

## Development

```bash
go build -o osir .     # build
go test ./...          # run tests (43 tests)
go vet ./...           # static analysis
make build-all         # cross-compile for 5 platforms
```

## License

Proprietary. Copyright OSIR Pty Ltd.
