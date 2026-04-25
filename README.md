# omctl — OpenMetadata for Your Terminal

> **"47% of knowledge workers still can't find the data they need."**
> `omctl` fixes that — without leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.21+-blue?logo=go&logoColor=white)
![OpenMetadata](https://img.shields.io/badge/OpenMetadata-1.12-orange?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCI+PC9zdmc+)
![Track](https://img.shields.io/badge/Track-T--04%20Developer%20Tooling-6366f1)
![License](https://img.shields.io/badge/License-Apache%202.0-green)

---

## The Problem

Data engineers context-switch **dozens of times a day** — jumping between terminal and the OpenMetadata UI just to answer simple questions:

- *Who owns this table?*
- *Where does this data come from?*
- *Which tables are missing owners?*
- *How healthy is our metadata overall?*

Every UI switch breaks flow. `omctl` eliminates that friction entirely.

> `omctl` is to OpenMetadata what `kubectl` is to Kubernetes — a first-class CLI that makes metadata management effortless.

---

## Demo

```bash
# Search any data asset instantly
$ omctl search "customer"

🔍 24 results for "customer"

  📋 sample_redshift.staging_db.integration.dim_customer
     Owner : no owner
     Desc  : Raw customer data from MySQL CDC

  📋 acme_nexus_analytics.ANALYTICS.MARTS.dim_customers
     Owner : no owner
     Desc  : Pre-aggregated customer metrics

# Who owns a table?
$ omctl whoowns "dim_customer"

📋 sample_redshift.staging_db.integration.dim_customer
─────────────────────────────────────────
  👤 Owner  : ⚠️  no owner assigned
  📝 Desc   : Raw customer data from MySQL CDC
  🏢 Domain : ⚠️  no domain
  🏷️  Tags   : PII.Sensitive, DataTier.Gold

# Trace full data lineage
$ omctl lineage "dim_customer"

📋 sample_redshift.staging_db.integration.dim_customer — Lineage
─────────────────────────────────────────
  ⬆️  UPSTREAM (data comes from)
      └── stg_customers
      └── raw_customers

  ⬇️  DOWNSTREAM (data goes to)
      └── customer_360

# Scan 1276 tables — governance score in seconds
$ omctl health

📊 OpenMetadata Health Report
─────────────────────────────────────────
  Total Tables    : 1276

  👤 Has Owner    : 844/1276  (66%)
  📝 Has Desc     : 920/1276  (72%)
  🏷️  Has Tags     : 731/1276  (57%)

  Overall Score   : 65/100 🔴

# Tag a table directly from your terminal
$ omctl tag "sample_redshift.staging_db.integration.dim_customer" "PII.Sensitive"

✅ Successfully tagged!
   Table : sample_redshift.staging_db.integration.dim_customer
   Tag   : PII.Sensitive

# List tables with governance issues — filterable
$ omctl list --no-owner
$ omctl list --no-tags
$ omctl list --no-owner --no-desc

# Generate a beautiful HTML report with PDF export
$ omctl report
$ omctl report --output my-report.html

✅ Report generated!
   File   : governance-report.html
   Tables : 1276 scanned
   Score  : 65/100 🔴

   open governance-report.html
```

---

## Commands

| Command | Description |
|---------|-------------|
| `omctl search <query>` | Search any data asset across the catalog |
| `omctl whoowns <table>` | Get owner, description, domain and tags |
| `omctl lineage <table>` | Show upstream + downstream lineage tree |
| `omctl health` | Scan all tables — show governance score |
| `omctl tag <table> <tagFQN>` | Write a tag back to OpenMetadata |
| `omctl list` | List all tables with governance status |
| `omctl list --no-owner` | Only tables missing owner |
| `omctl list --no-desc` | Only tables missing description |
| `omctl list --no-tags` | Only tables missing tags |
| `omctl report` | Generate HTML report with PDF export |
| `omctl report -o file.html` | Custom output path |

---

## Quick Start — 2 Minutes to Running

### Step 1 — Check Go is installed

```bash
go version
# Need: go1.21 or higher
```

Not installed?
```bash
brew install go        # Mac
sudo apt install golang-go  # Ubuntu/Linux
```

### Step 2 — Clone and install

```bash
git clone https://github.com/shivakumar2006/omctl
cd omctl
go install .
```

### Step 3 — Fix PATH if needed

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Make it permanent:
```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

### Step 4 — Get your OpenMetadata token

1. Go to `sandbox.open-metadata.org`
2. Login with Google
3. Click your **profile picture** (top right)
4. Go to **Access Token** tab
5. Click **Generate New Token** → Copy it

### Step 5 — Set environment and run

```bash
export OM_HOST=https://sandbox.open-metadata.org
export OM_TOKEN=your_token_here

omctl health
```

Make it permanent:
```bash
echo 'export OM_HOST=https://sandbox.open-metadata.org' >> ~/.zshrc
echo 'export OM_TOKEN=your_token_here' >> ~/.zshrc
source ~/.zshrc
```

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `omctl: command not found` | `export PATH=$PATH:$(go env GOPATH)/bin && go install .` |
| `OM_TOKEN not set` | `export OM_TOKEN=your_token_here` |
| `API error 401` | Token expired — generate a new one from profile page |
| `API error 404` | Check `OM_HOST` is correct |
| `go: command not found` | `brew install go` (Mac) or `sudo apt install golang-go` (Linux) |

---

## Why omctl?

| Without omctl | With omctl |
|---------------|------------|
| Open browser tab | Stay in terminal |
| Navigate to OpenMetadata UI | `omctl search "table"` |
| Click through 3+ pages | `omctl whoowns "table"` |
| View lineage graph manually | `omctl lineage "table"` |
| Manually export health report | `omctl health` |
| Tag via form in UI | `omctl tag "table" "PII.Sensitive"` |
| Download report manually | `omctl report` → PDF in one click |

---

## Architecture

```
omctl/
├── main.go
├── cmd/
│   ├── root.go       # cobra root
│   ├── search.go     # omctl search
│   ├── whoowns.go    # omctl whoowns
│   ├── lineage.go    # omctl lineage
│   ├── health.go     # omctl health
│   ├── tag.go        # omctl tag
│   ├── list.go       # omctl list
│   └── report.go     # omctl report
├── internal/
│   └── client/       # OpenMetadata REST client
└── config/           # env config loader
```

**Stack:** Go · Cobra · OpenMetadata REST API · Zero external dependencies

Single binary. No Docker. No config files. Just set two env vars and go.

---

## OpenMetadata API Usage

| Endpoint | Used By |
|----------|---------|
| `GET /api/v1/search/query` | `search`, `whoowns`, `lineage` |
| `GET /api/v1/tables` | `health`, `list`, `report` |
| `GET /api/v1/lineage/table/{id}` | `lineage` |
| `GET /api/v1/tables/name/{fqn}` | `tag` |
| `PATCH /api/v1/tables/{id}` | `tag` (write-back) |

---

## Real Numbers

Tested live against OpenMetadata sandbox — **1,276 real tables**:

- Full scan completed in **under 10 seconds**
- **432 tables** missing owners (34%)
- **356 tables** missing descriptions (28%)
- **545 tables** missing tags (43%)
- Governance score: **65/100 🔴**

All fixable directly from the terminal with `omctl tag` and `omctl list`.

---

## Built For

**Back to the Metadata Hackathon** — Paradox **#T-04: Developer Tooling & CI/CD**

> *"Build CLI tools, GitHub Actions, CI/CD integrations, IDE plugins, or developer-facing utilities that make working with metadata effortless."*

---

## Author

**Shiva** — [@shivakumar2006](https://github.com/shivakumar2006)

BCA Final Year · 1700+ GitHub commits · Hackathon builder · Open source contributor
