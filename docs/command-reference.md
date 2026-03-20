# Command Reference

Complete reference for all 76 OSIR CLI commands.

---

## Authentication

### auth login

Authenticate to the OSIR platform.

```
osir auth login [flags]
```

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--device` | `-d` | bool | `false` | Use OAuth 2.0 Device Authorization flow |
| `--username` | `-u` | string | | Username for password login |

**Device flow** (recommended for headless servers):

```bash
osir auth login --device
```

**Password login:**

```bash
osir auth login -u john.doe
```

### auth status

Display current authentication status.

```bash
osir auth status
```

### auth logout

Clear stored credentials and log out.

```bash
osir auth logout
```

---

## Domain Management

### domain check

Check if a domain name is available for registration.

```bash
osir domain check <domain>
```

### domain register

Register a new domain name.

```bash
osir domain register <domain> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--years` | int | `1` | Registration period in years |
| `--nameservers` | string | | Nameservers (comma-separated) |
| `--privacy` | bool | `false` | Enable WHOIS privacy protection |
| `--auto-renew` | bool | `false` | Enable auto-renewal |

```bash
osir domain register coolstartup.io --years 2 --privacy --auto-renew
osir domain register example.com --nameservers "ns1.cloudflare.com,ns2.cloudflare.com"
```

### domain info

Get detailed information about a domain.

```bash
osir domain info <domain>
```

### domain list

List all domains in your account.

```bash
osir domain list [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--page` | int | `0` | Page number |
| `--size` | int | `0` | Page size (0 = server default) |
| `--sort-by` | string | | Field to sort by |
| `--sort-dir` | string | | Sort direction (asc, desc) |

### domain renew

Renew a domain registration.

```bash
osir domain renew <domain> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--years` | int | `1` | Number of years to renew |

### domain lock

Lock a domain to prevent unauthorized transfers.

```bash
osir domain lock <domain>
```

### domain unlock

Unlock a domain to allow transfers.

```bash
osir domain unlock <domain>
```

### domain auto-renew

Enable or disable auto-renewal.

```bash
osir domain auto-renew <domain> <true|false>
```

### domain privacy

Enable or disable WHOIS privacy protection.

```bash
osir domain privacy <domain> <true|false>
```

### domain validate

Validate a domain name format (local validation, no API call).

```bash
osir domain validate <domain>
```

### domain suggest

Get alternative domain name suggestions.

```bash
osir domain suggest <keyword> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `10` | Maximum number of suggestions |
| `--tlds` | string | | Comma-separated TLDs to search |

### domain nameservers

Update nameservers for a domain.

```bash
osir domain nameservers <domain> <ns1> [ns2...]
```

```bash
osir domain nameservers example.com ns1.cloudflare.com ns2.cloudflare.com
```

---

## DNS Record Management

### dns list

List all DNS records for a domain.

```bash
osir dns list <domain>
```

### dns get

Get DNS records using a record ID or smart selector.

```bash
osir dns get <domain> <recordId|TYPE> [nameOrContent]
```

**Smart selector examples:**

```bash
osir dns get example.com 12345           # by record ID
osir dns get example.com A               # all A records
osir dns get example.com MX mail         # MX records matching "mail"
osir dns get example.com TXT spf         # TXT records matching "spf"
```

### dns create

Create a new DNS record.

```bash
osir dns create <domain> <type> <name> <content> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--ttl` | int | Time to live in seconds |
| `--priority` | int | Priority (for MX, SRV records) |

```bash
osir dns create example.com A @ 192.0.2.1 --ttl 3600
osir dns create example.com MX @ mail.google.com --priority 5
osir dns create example.com TXT @ "v=spf1 include:_spf.google.com ~all"
```

### dns update

Update an existing DNS record using a record ID or smart selector.

```bash
osir dns update <domain> <recordId|TYPE> [args...] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--content` | string | New record content |
| `--type` | string | New record type |
| `--name` | string | New record name |
| `--ttl` | int | New time to live in seconds |
| `--priority` | int | New priority |

```bash
osir dns update example.com 12345 --content 192.0.2.2 --ttl 7200
osir dns update example.com A @ --content 192.0.2.2
```

### dns delete

Delete a DNS record using a record ID or smart selector.

```bash
osir dns delete <domain> <recordId|TYPE> [nameOrContent] [content] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--force` | bool | Skip confirmation prompt |

```bash
osir dns delete example.com 12345
osir dns delete example.com A @                    # delete A record named @
osir dns delete example.com TXT @ "v=spf1..." --force
```

### dns zone-init

Initialize the DNS zone for a domain.

```bash
osir dns zone-init <domain>
```

### dns zone-exists

Check if a DNS zone exists for a domain.

```bash
osir dns zone-exists <domain>
```

### dns fix-soa

Fix the SOA record for a domain zone.

```bash
osir dns fix-soa <domain>
```

### dns dnssec-status

Check DNSSEC status for a domain.

```bash
osir dns dnssec-status <domain>
```

### dns dnssec-enable

Enable DNSSEC for a domain.

```bash
osir dns dnssec-enable <domain>
```

### dns dnssec-disable

Disable DNSSEC for a domain.

```bash
osir dns dnssec-disable <domain>
```

---

## Billing & Payments

### billing balance

Get your account balance.

```bash
osir billing balance
```

### billing invoices

List invoices.

```bash
osir billing invoices [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | | Filter by status (e.g., pending, paid) |
| `--page` | int | `0` | Page number |
| `--size` | int | `0` | Page size |

### billing invoice

Get details of a specific invoice.

```bash
osir billing invoice <invoiceId>
```

### billing invoice-number

Look up an invoice by its invoice number.

```bash
osir billing invoice-number <invoiceNumber>
```

### billing pay

Pay an invoice.

```bash
osir billing pay <invoiceId> <amount>
```

### billing stats

Get invoice statistics (total paid, pending, overdue).

```bash
osir billing stats
```

### billing checkout

Create a payment checkout session to add funds.

```bash
osir billing checkout <amount> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--currency` | string | `USD` | Currency code |
| `--description` | string | | Payment description |

```bash
osir billing checkout 50.00
osir billing checkout 100.00 --currency EUR --description "Top-up"
```

### billing history

List payment transactions.

```bash
osir billing history [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--page` | int | `0` | Page number |
| `--size` | int | `0` | Page size |

### billing fees

Preview fees for an amount.

```bash
osir billing fees <amount> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--currency` | string | `USD` | Currency code |
| `--processor` | string | | Payment processor |

```bash
osir billing fees 50.00
osir billing fees 100.00 --currency EUR --processor stripe
```

### billing pricing

Get domain pricing, optionally filtered by extension.

```bash
osir billing pricing [extension]
```

```bash
osir billing pricing          # show all pricing
osir billing pricing com      # show pricing for .com
```

### billing session

Get details of a payment session.

```bash
osir billing session <sessionId>
```

### billing listen

Listen for payment webhook events.

```bash
osir billing listen
```

---

## Contact Management

### contact list

List all contacts.

```bash
osir contact list
```

### contact get

Get contact details.

```bash
osir contact get <contactId>
```

### contact create

Create a new contact.

```bash
osir contact create [flags]
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--first-name` | string | Yes | First name |
| `--last-name` | string | Yes | Last name |
| `--email` | string | Yes | Email address |
| `--phone` | string | No | Phone number |
| `--organization` | string | No | Organization name |
| `--street1` | string | No | Street address line 1 |
| `--street2` | string | No | Street address line 2 |
| `--city` | string | No | City |
| `--state` | string | No | State or province |
| `--postal-code` | string | No | Postal code |
| `--country` | string | No | Country code (e.g., US, AU) |

```bash
osir contact create \
  --first-name John --last-name Doe \
  --email john@example.com \
  --phone "+1-555-0123" \
  --organization "Acme Inc" \
  --street1 "123 Main St" \
  --city "San Francisco" --state "CA" \
  --postal-code "94105" --country "US"
```

### contact update

Update an existing contact. Same flags as `contact create` (all optional).

```bash
osir contact update <contactId> [flags]
```

### contact delete

Delete a contact.

```bash
osir contact delete <contactId>
```

### contact for-domain

Get contacts associated with a domain.

```bash
osir contact for-domain <domain>
```

---

## VPS Hosting

### vps packages

List available VPS packages with pricing. No authentication required.

```bash
osir vps packages
```

```
NAME       CPU   RAM     DISK      TRAFFIC   STORAGE   MONTHLY   ANNUAL    LOCATION
ZANA-S     1     1 GB    25 GB     1 TB      20 GB     $2.99     $29.99    Nueremberg
ZANA-M     2     2 GB    50 GB     2 TB      40 GB     $5.99     $59.99    Nueremberg
ZANA-L     4     4 GB    100 GB    4 TB      80 GB     $9.99     $99.99    Nueremberg
```

### vps locations

List available datacenter locations. No authentication required.

```bash
osir vps locations
```

### vps list

List all your VPS instances.

```bash
osir vps list
```

```
ID          HOSTNAME     STATUS    PACKAGE   IPv4            TERM      RENEWAL      LOCATION
a1b2c3d4    web01        running   ZANA-M    203.0.113.10    MONTHLY   2026-04-20   Nueremberg
e5f6g7h8    db01         running   ZANA-L    203.0.113.11    ANNUAL    2027-03-20   Nueremberg
```

### vps active

List active VPS instances only.

```bash
osir vps active
```

### vps count

Count your VPS instances.

```bash
osir vps count [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--active-only` | bool | `false` | Count only active instances |

```bash
osir vps count
osir vps count --active-only
```

### vps info

Get detailed information about a VPS instance. Auto-generates a control panel login URL.

```bash
osir vps info <instanceId>
```

Short instance ID prefixes are accepted and resolve to the full UUID.

```bash
osir vps info a1b2c3d4
```

### vps order

Order a new VPS instance by package name.

```bash
osir vps order --package <NAME> --hostname <hostname> [flags]
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--package` | string | Yes | Package name (e.g., ZANA-S, ZANA-M, ZANA-L) |
| `--hostname` | string | Yes | Hostname for the instance |
| `--payment-term` | string | No | Payment term (MONTHLY, QUARTERLY, SEMI_ANNUAL, ANNUAL, BIENNIAL, TRIENNIAL) |
| `--location` | string | No | Datacenter location name (e.g., Nueremberg) |
| `--root-password` | string | No | Root password for the instance |

```bash
osir vps order --package ZANA-S --hostname web01
osir vps order --package ZANA-L --hostname db01 --payment-term ANNUAL --location Nueremberg
```

### vps delete

Delete a VPS instance.

```bash
osir vps delete <instanceId>
```

```bash
osir vps delete a1b2c3d4
```

### vps change-term

Change the payment term for a VPS instance.

```bash
osir vps change-term <instanceId> --term <TERM>
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--term` | string | Yes | New payment term (MONTHLY, QUARTERLY, SEMI_ANNUAL, ANNUAL, BIENNIAL, TRIENNIAL) |

```bash
osir vps change-term a1b2c3d4 --term ANNUAL
```

### vps login

Generate a control panel SSO login URL for a VPS instance.

```bash
osir vps login <instanceId>
```

```bash
osir vps login a1b2c3d4
```

---

## Audit Logs

### audit recent

Get recent account activity.

```bash
osir audit recent [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | | Maximum number of entries to show |

### audit domain

Get the audit trail for a specific domain.

```bash
osir audit domain <domain> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--page` | int | `0` | Page number |
| `--size` | int | `0` | Page size |

### audit failures

Get recent failed operations.

```bash
osir audit failures [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--page` | int | `0` | Page number |
| `--size` | int | `0` | Page size |

---

## Account Management

### account profile

Show your user profile.

```bash
osir account profile
```

### account summary

Show a rich account dashboard with domain count, balance, recent activity, and more.

```bash
osir account summary
```

---

## Product Catalog

### catalog domains

List domain extensions with pricing.

```bash
osir catalog domains
```

### catalog servers

List dedicated server configurations.

```bash
osir catalog servers
```

---

## Domain Name Suggestions

### suggest generate

Generate domain name suggestions using AI.

```bash
osir suggest generate <name> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tlds` | string | | Comma-separated TLDs |
| `--lang` | string | | Language code (e.g., eng) |
| `--numbers` | bool | `false` | Include numbers in suggestions |
| `--max` | int | `20` | Maximum results |

### suggest spin

Generate suggestions by replacing words with synonyms.

```bash
osir suggest spin <name> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--position` | int | `0` | Word position to spin (0-based) |
| `--similarity` | float | `0.7` | Similarity threshold (0.0-1.0) |
| `--tlds` | string | | Comma-separated TLDs |
| `--lang` | string | | Language code |
| `--max` | int | `20` | Maximum results |

### suggest prefix

Generate suggestions by adding prefixes.

```bash
osir suggest prefix <name> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--vocabulary` | string | `@prefixes` | Vocabulary source |
| `--tlds` | string | | Comma-separated TLDs |
| `--lang` | string | | Language code |
| `--max` | int | `20` | Maximum results |

### suggest suffix

Generate suggestions by adding suffixes.

```bash
osir suggest suffix <name> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--vocabulary` | string | `@suffixes` | Vocabulary source |
| `--tlds` | string | | Comma-separated TLDs |
| `--lang` | string | | Language code |
| `--max` | int | `20` | Maximum results |

### suggest bulk

Generate bulk suggestions for multiple keywords.

```bash
osir suggest bulk <keyword1> [keyword2...] [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tlds` | string | | Comma-separated TLDs |
| `--lang` | string | | Language code |
| `--numbers` | bool | `false` | Include numbers in suggestions |
| `--max` | int | `20` | Maximum results |

### suggest keyword

Check keyword availability across TLDs with detailed results.

```bash
osir suggest keyword <keyword> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--registries` | string | Comma-separated registries (e.g., verisign,pir) |
| `--tlds` | string | Comma-separated TLDs |

### suggest keyword-summary

Check keyword availability summary (faster, no per-domain details).

```bash
osir suggest keyword-summary <keyword> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--registries` | string | Comma-separated registries |
| `--tlds` | string | Comma-separated TLDs |

---

## Interactive Shell

### shell

Launch the interactive shell with tab completion and `?` help.

```bash
osir shell
```

See [Interactive Shell Guide](interactive-shell.md) for details.

---

## Shell Completion

### completion

Generate shell completion scripts.

```bash
osir completion <bash|zsh|fish|powershell>
```

See [Installation Guide](installation.md#shell-completion-setup) for setup instructions.

---

## Global Flags

These flags are available on every command:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `text` | Output format: `text` or `json` |
| `--verbose` | `-v` | `false` | Show timing and debug info |
| `--timeout` | | `30s` | HTTP request timeout (e.g. `10s`, `1m`) |
| `--version` | | | Print version and exit |
| `--help` | `-h` | | Show help for any command |

```bash
osir -o json domain check example.com     # JSON output
osir -v domain list                        # show HTTP timing
osir --timeout 60s domain list             # custom timeout
```
