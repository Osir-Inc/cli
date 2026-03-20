# Getting Started

This guide walks you through your first session with the OSIR CLI -- from authentication to registering a domain and setting up DNS.

## Step 1: Authenticate

The CLI supports two authentication methods.

### Device Flow (recommended)

Best for headless servers and environments where you can't type a password. Run:

```bash
osir auth login --device
```

You'll see:

```
To sign in, open the following URL in a browser:

  https://auth.osir.com/realms/osir/device?user_code=ABCD-EFGH

Waiting for authentication...
```

Open that URL on **any device** -- your laptop, phone, or a colleague's computer -- and log in. The CLI detects the authentication automatically.

```
[OK] Logged in as john.doe
```

The device flow supports MFA, SSO, and any identity provider configured in your Keycloak realm.

### Username and Password

```bash
osir auth login -u john.doe
Password: ********
[OK] Logged in as john.doe
```

If you omit `-u`, you'll be prompted for the username interactively.

### Check Your Status

```bash
osir auth status
```

```
Authenticated
Username:         john.doe
Method:           device
Token expires in: 4m 45s
```

Your credentials are saved at `~/.osir/credentials.json` and automatically refreshed. You won't need to log in again until your refresh token expires.

## Step 2: Find a Domain

### Check availability

```bash
osir domain check coolstartup.com
```

### Get suggestions if it's taken

```bash
# Quick suggestions
osir domain suggest coolstartup --limit 10 --tlds "com,net,io"

# AI-generated suggestions
osir suggest generate coolstartup --tlds "com,net,io" --max 20

# Check a keyword across many TLDs
osir suggest keyword coolstartup --tlds "com,net,org,io,co,app,dev"
```

### Check pricing

```bash
osir billing pricing com
osir billing pricing io --operation register
```

## Step 3: Register a Domain

```bash
osir domain register coolstartup.io --years 2 --privacy --auto-renew
```

With custom nameservers:

```bash
osir domain register coolstartup.io \
  --years 1 \
  --nameservers "ns1.cloudflare.com,ns2.cloudflare.com" \
  --privacy \
  --auto-renew
```

## Step 4: Set Up DNS

### View current records

```bash
osir dns list coolstartup.io
```

```
ID         TYPE   NAME   CONTENT           TTL    PRIORITY
rec-001    A      @      192.0.2.1         3600
rec-002    A      www    192.0.2.1         3600
```

### Create records

Point your domain to a server:

```bash
osir dns create coolstartup.io A coolstartup.io 192.0.2.1
osir dns create coolstartup.io A www.coolstartup.io 192.0.2.1
```

Set up email (MX records):

```bash
osir dns create coolstartup.io MX coolstartup.io aspmx.l.google.com
osir dns create coolstartup.io MX coolstartup.io alt1.aspmx.l.google.com
```

Add a CNAME alias:

```bash
osir dns create coolstartup.io CNAME blog.coolstartup.io mysite.github.io
```

Add email security (SPF):

```bash
osir dns create coolstartup.io TXT coolstartup.io "v=spf1 include:_spf.google.com ~all"
```

### Update or delete records

Use smart selectors to target records by type and name instead of opaque IDs:

```bash
osir dns update coolstartup.io rec-001 --content 198.51.100.1
osir dns delete coolstartup.io rec-003
```

## Step 5: Explore More

### Your account

```bash
osir account profile          # your profile info
osir account summary          # domains, balance
osir billing balance          # account balance
```

### Your domains

```bash
osir domain list              # list all domains
osir domain info example.com  # detailed domain info
osir domain lock example.com  # prevent unauthorized transfers
```

### VPS hosting

```bash
osir vps packages             # browse available packages
osir vps order --package ZANA-S --hostname web01   # order a VPS
osir vps list                 # list your instances
osir vps info a1b2            # instance details (short ID prefix works)
osir vps login a1b2           # control panel SSO login URL
```

### Audit trail

```bash
osir audit domain example.com # what happened to this domain?
osir audit recent             # recent account activity
```

## Interactive Shell

For a guided, network-switch-style experience, launch the interactive shell:

```bash
osir shell
```

See the [Interactive Shell Guide](interactive-shell.md) for details.

## Getting Help

Every command has built-in help:

```bash
osir --help                    # top-level help
osir domain --help             # domain command group
osir domain register --help    # specific command with all flags
osir dns create --help         # see positional args and flags
```
