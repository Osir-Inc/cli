# Interactive Shell

The OSIR CLI includes a Junos/Arista-style interactive shell with context-aware tab completion, `?` help, and persistent command history.

## Launching the Shell

```bash
osir shell
```

```
OSIR Interactive Shell v1.0.0
Type 'help' for commands, Tab or '?' for completions, 'exit' to quit.

osir>
```

The prompt shows `osir>` on the left and your authentication status on the right (username if authenticated, "not authenticated" otherwise).

## Tab Completion

Press **Tab** at any point to see available completions:

```
osir> d<Tab>
dns     domain

osir> domain <Tab>
auto-renew   check   info   list   lock   nameservers   privacy
register     renew   suggest   unlock   validate

osir> domain register --<Tab>
--auto-renew   --nameservers   --privacy   --years
```

Completions are context-aware -- they know which commands, subcommands, and flags are valid at the current cursor position.

## The ? Key

Press **?** at any point to see available completions with descriptions. This works the same way as on Junos and Arista network switches:

```
osir> ?
account     Account management
audit       Audit log management
auth        Authentication management
billing     Billing and payment commands
vps         VPS hosting management
...

osir> domain ?
auto-renew  Enable or disable auto-renewal for a domain
check       Check domain availability
info        Get detailed domain information
list        List all domains in your account
...
```

The `?` character is intercepted by the shell and is not inserted into your command. If you need a literal `?` in an argument, use the non-interactive CLI mode instead.

## Command History

Commands are saved to `~/.osir/shell_history` and persist across sessions. Use the standard readline keys:

| Key | Action |
|-----|--------|
| Up arrow | Previous command |
| Down arrow | Next command |
| Ctrl+R | Reverse search history |

## Shell-Only Commands

These commands are only available inside the interactive shell:

| Command | Description |
|---------|-------------|
| `exit` or `quit` | Exit the shell |
| `clear` | Clear the screen |
| `help` | Show available commands |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Tab | Show/cycle completions |
| ? | Show completions with descriptions |
| Ctrl+C | Cancel current input line |
| Ctrl+D | Exit the shell |
| Ctrl+L | Clear screen |
| Ctrl+A | Move cursor to start of line |
| Ctrl+E | Move cursor to end of line |
| Ctrl+W | Delete word before cursor |
| Ctrl+R | Reverse history search |

## Session Persistence

The interactive shell creates the `App` (config, auth session, API client) once at startup and reuses it for every command. This means:

- You log in once with `auth login --device` and stay authenticated for the entire session
- Token refresh happens automatically in the background
- No per-command overhead for loading credentials

## Example Session

```
$ osir shell
OSIR Interactive Shell v1.0.0
Type 'help' for commands, Tab or '?' for completions, 'exit' to quit.

osir> auth login --device
To sign in, open the following URL in a browser:

  https://auth.osir.com/realms/osir/device?user_code=WXYZ-1234

Waiting for authentication...
[OK] Logged in as john.doe

osir> domain check coolstartup.io
Domain: coolstartup.io
Available: true

osir> domain register coolstartup.io --years 2 --privacy --auto-renew
[OK] Domain registered: coolstartup.io

osir> dns create coolstartup.io A coolstartup.io 192.0.2.1
[OK] DNS record created (ID: coolstartup_io__A_123456)

osir> dns create coolstartup.io A www.coolstartup.io 192.0.2.1
[OK] DNS record created (ID: coolstartup_io__A_123457)

osir> dns list coolstartup.io
ID         TYPE   NAME   CONTENT      TTL    PRIORITY
rec-001    A      @      192.0.2.1    3600
rec-002    A      www    192.0.2.1    3600

osir> billing balance
Balance: 87.50 USD

osir> exit
Goodbye!
```

## Shell vs Non-Interactive Mode

| Feature | `osir shell` | `osir <command>` |
|---------|-------------|------------------|
| Tab completion | Built-in | Requires `osir completion` setup |
| `?` help | Built-in | Not available |
| Command history | Built-in, persistent | Shell-dependent |
| Auth session | Created once, reused | Created per invocation |
| Scriptable | No (interactive) | Yes (pipe, cron, CI/CD) |
| Exit code | Always 0 on exit | Non-zero on errors |

Use the interactive shell for exploration and manual work. Use non-interactive mode for scripts and automation.
