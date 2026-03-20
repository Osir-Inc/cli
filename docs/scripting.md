# Scripting & Automation

Every OSIR CLI command supports `-o json` for machine-readable output. Combined with tools like `jq`, you can build powerful automation workflows.

## JSON Output

Add `-o json` to any command:

```bash
osir -o json domain check coolstartup.com
```

```json
{
  "domain": "coolstartup.com",
  "available": true
}
```

## Common Patterns with jq

### Extract a single value

```bash
osir -o json domain check coolstartup.com | jq '.available'
# true

osir -o json billing balance | jq '.balance'
# 87.50
```

### List domain names

```bash
osir -o json domain list | jq -r '.domains[].domain'
# coolstartup.io
# example.com
# mysite.net
```

### List VPS hostnames

```bash
osir -o json vps list | jq -r '.[].hostname'
# web01
# db01
```

### Filter results

```bash
# Domains expiring before a date
osir -o json domain list | jq '.domains[] | select(.expirationDate < "2026-06-01")'

# Only available suggestions
osir -o json suggest generate startup | jq '[.results[] | select(.availability == "available")]'

# Unpaid invoices
osir -o json billing invoices --status pending | jq '.invoices'
```

## Batch Operations

### Check multiple domains

```bash
for domain in coolstartup.com coolstartup.io coolstartup.net coolstartup.dev; do
  available=$(osir -o json domain check "$domain" | jq -r '.available')
  echo "$domain: $available"
done
```

### Back up DNS records

```bash
osir -o json dns list coolstartup.io > dns-backup-$(date +%Y%m%d).json
```

### Restore DNS from backup

```bash
cat dns-backup.json | jq -c '.[]' | while read -r record; do
  type=$(echo "$record" | jq -r '.type')
  name=$(echo "$record" | jq -r '.name')
  content=$(echo "$record" | jq -r '.content')
  ttl=$(echo "$record" | jq -r '.ttl')
  priority=$(echo "$record" | jq -r '.priority')

  osir dns create coolstartup.io \
    --type "$type" --name "$name" --content "$content" \
    --ttl "$ttl" --priority "$priority"
done
```

### Deploy to multiple servers

```bash
for server in web01 web02 db01 cache01; do
  scp dist/osir-linux-amd64 user@$server:/usr/local/bin/osir
  ssh user@$server chmod +x /usr/local/bin/osir
done
```

## Cron Jobs

### Daily balance check

```bash
# /etc/cron.d/osir-balance
0 9 * * * john /usr/local/bin/osir -o json billing balance >> /var/log/osir-balance.log 2>&1
```

### Weekly expiring domain alert

```bash
# /etc/cron.d/osir-expiring
0 8 * * 1 john /usr/local/bin/osir -o json domain list | jq -r '.domains[] | select(.expirationDate < "2026-04-01") | .domain' | mail -s "Expiring domains" admin@company.com
```

## CI/CD Integration

### GitHub Actions example

```yaml
jobs:
  deploy-dns:
    runs-on: ubuntu-latest
    steps:
      - name: Install OSIR CLI
        run: |
          curl -L -o osir https://releases.osir.com/osir-linux-amd64
          chmod +x osir
          sudo mv osir /usr/local/bin/

      - name: Authenticate
        env:
          OSIR_BACKEND_URL: ${{ secrets.OSIR_BACKEND_URL }}
          KEYCLOAK_URL: ${{ secrets.KEYCLOAK_URL }}
        run: |
          osir auth login -u ${{ secrets.OSIR_USER }}
          # Password provided via stdin or env

      - name: Update DNS
        run: |
          osir dns create myapp.com A myapp.com ${{ env.SERVER_IP }} --ttl 300
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error (API failure, invalid args, auth required) |

Use exit codes in scripts:

```bash
if osir domain check coolstartup.com -o json | jq -e '.available' > /dev/null 2>&1; then
  echo "Domain is available -- registering..."
  osir domain register coolstartup.com --years 1 --privacy
else
  echo "Domain is taken"
fi
```

## Tips

- Always use `-o json` in scripts for stable, parseable output
- Use `jq -r` for raw strings (no quotes) and `jq -e` for exit-code filtering
- Pipe output to `tee` to log and process simultaneously: `osir -o json domain list | tee domains.json | jq '.domains | length'`
- Set `OSIR_BACKEND_URL` in your CI environment to target staging/production
