# omctl — OpenMetadata for Your Terminal

> **"47% of knowledge workers still can't find the data they need."**  
> `omctl` fixes that — without leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.21+-blue?logo=go&logoColor=white)
![OpenMetadata](https://img.shields.io/badge/OpenMetadata-1.12-orange)
![Track](https://img.shields.io/badge/Track-T--04%20Developer%20Tooling-6366f1)
![License](https://img.shields.io/badge/License-Apache%202.0-green)

---

## The Problem

Data engineers context-switch **dozens of times a day** — jumping between terminal and the OpenMetadata UI just to answer simple questions:

- *Who owns this table?*
- *Where does this data come from?*
- *Which tables are missing PII tags?*
- *How do I bulk-fix 545 untagged tables?*

Every UI switch breaks flow. Manual governance takes hours. `omctl` eliminates that friction entirely.

> `omctl` is to OpenMetadata what `kubectl` is to Kubernetes.

---

## Demo

```bash
# 1. Check governance health instantly
$ omctl health

📊 OpenMetadata Health Report
─────────────────────────────────────────
  Total Tables    : 1276

  👤 Has Owner    : 844/1276  (66%)
  📝 Has Desc     : 920/1276  (72%)
  🏷️  Has Tags     : 731/1276  (57%)

  Overall Score   : 65/100 🔴

# 2. Search any data asset
$ omctl search "customer"

🔍 24 results for "customer"
  📋 sample_redshift.staging_db.integration.dim_customer
     Owner : no owner
     Desc  : Raw customer data from MySQL CDC

# 3. Who owns a table?
$ omctl whoowns "dim_customer"

📋 sample_redshift.staging_db.integration.dim_customer
─────────────────────────────────────────
  👤 Owner  : ⚠️  no owner assigned
  📝 Desc   : Raw customer data from MySQL CDC
  🏢 Domain : ⚠️  no domain
  🏷️  Tags   : PII.Sensitive, DataTier.Gold

# 4. Trace full data lineage
$ omctl lineage "dim_customer"

📋 sample_redshift.staging_db.integration.dim_customer — Lineage
─────────────────────────────────────────
  ⬆️  UPSTREAM (data comes from)
      └── stg_customers
      └── raw_customers

  ⬇️  DOWNSTREAM (data goes to)
      └── customer_360

# 5. Tag a single table
$ omctl tag "sample_redshift.staging_db.integration.dim_customer" "PII.Sensitive"

✅ Successfully tagged!
   Table : sample_redshift.staging_db.integration.dim_customer
   Tag   : PII.Sensitive

# 6. Bulk auto-fix — preview first
$ omctl fix --no-tags --dry-run

🔍 DRY RUN — no changes will be applied

📊 Issues found:
   🏷️  Missing tags : 545 tables

  📋 ACME_MYSQL.default.FINANCIAL_STAGING.ACCOUNTS
     → tags→PII.Name
  📋 postgres_aws_harsh.TESTDB.sales.staffs
     → tags→PII.Email
  ...545 tables total

✅ Dry run complete. Remove --dry-run to apply.

# 7. Actually fix all 545 tables at once
$ omctl fix --no-tags

🔧 Scanning tables for issues...
📊 Issues found:
   🏷️  Missing tags : 545 tables

  🏷️  Tagged : ACME_MYSQL.default.FINANCIAL_STAGING.ACCOUNTS → PII.Name
  🏷️  Tagged : postgres_aws_harsh.TESTDB.sales.staffs → PII.Email
  🏷️  Tagged : MySQL2.mysql.mysql.password_history → PII.Sensitive
  ... 545 tables tagged

─────────────────────────────────────────────
✅ Fixed : 545 tables

# 8. List tables with governance issues
$ omctl list --no-owner
$ omctl list --no-tags --no-desc

# 9. Generate beautiful HTML report with PDF export
$ omctl report

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
| `omctl tag <table> <tagFQN>` | Tag a single table from terminal |
| `omctl list` | List all tables with governance status |
| `omctl list --no-owner` | Only tables missing owner |
| `omctl list --no-desc` | Only tables missing description |
| `omctl list --no-tags` | Only tables missing tags |
| `omctl fix --no-tags` | **Bulk auto-classify PII tags** across all tables |
| `omctl fix --no-desc` | Add auto-generated descriptions |
| `omctl fix --all` | Fix all governance issues at once |
| `omctl fix --dry-run` | Preview changes before applying |
| `omctl report` | Generate HTML report with PDF export |
| `omctl report -o file.html` | Custom output path |

---

## Quick Start — 2 Minutes

### Step 1 — Check Go

```bash
go version
# Need: go1.21+
```

Not installed?
```bash
brew install go          # Mac
sudo apt install golang-go  # Ubuntu
```

### Step 2 — Install

```bash
git clone https://github.com/shivakumar2006/OpenMetaDataCommandLineTool
cd omctl
go install .
```

### Step 3 — Fix PATH (if needed)

```bash
export PATH=$PATH:$(go env GOPATH)/bin

# Permanent fix
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

### Step 4 — Get Token

1. Go to `sandbox.open-metadata.org`
2. Login with Google
3. Profile picture → **Access Token** → **Generate New Token**
4. Copy the token

### Step 5 — Configure and Run

```bash
export OM_HOST=https://sandbox.open-metadata.org
export OM_TOKEN=your_token_here

# Permanent
echo 'export OM_HOST=https://sandbox.open-metadata.org' >> ~/.zshrc
echo 'export OM_TOKEN=your_token_here' >> ~/.zshrc
source ~/.zshrc

# Test
omctl health
```

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `omctl: command not found` | `export PATH=$PATH:$(go env GOPATH)/bin && go install .` |
| `OM_TOKEN not set` | `export OM_TOKEN=your_token_here` |
| `API error 401` | Token expired — generate a new one |
| `API error 400` | Check OM_HOST is correct |
| `go: command not found` | `brew install go` |

---

## Why omctl?

| Without omctl | With omctl |
|---------------|------------|
| Open browser | Stay in terminal |
| Search UI manually | `omctl search "table"` |
| Click through pages | `omctl whoowns "table"` |
| View lineage graph | `omctl lineage "table"` |
| Tag tables one by one | `omctl fix --no-tags` ← 545 tables at once |
| Export report manually | `omctl report` → PDF in one click |
| No bulk governance | `omctl fix --all --dry-run` |

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
│   ├── fix.go        # omctl fix (bulk automation)
│   └── report.go     # omctl report (HTML + PDF)
├── internal/
│   └── client/       # OpenMetadata REST client
└── config/           # env config loader
```

**Stack:** Go · Cobra · OpenMetadata REST API  
Single binary. No Docker. No config files. Zero external dependencies.

---

## OpenMetadata API Usage

| Endpoint | Used By |
|----------|---------|
| `GET /api/v1/search/query` | `search`, `whoowns`, `lineage` |
| `GET /api/v1/tables` | `health`, `list`, `fix`, `report` |
| `GET /api/v1/lineage/table/{id}` | `lineage` |
| `GET /api/v1/tables/name/{fqn}` | `tag` |
| `PATCH /api/v1/tables/{id}` | `tag`, `fix` (write-back) |

---

## Real Impact — Live Numbers

Tested against OpenMetadata sandbox — **1,276 real tables**:

- Full scan in **under 10 seconds**
- **545 tables** auto-tagged with PII classification in one command
- **432 tables** identified missing owners
- **356 tables** identified missing descriptions
- Governance score improved from **57% → 100%** on tags in one command

---

## Built For

**Back to the Metadata Hackathon** — Paradox **#T-04: Developer Tooling & CI/CD**

> *"Build CLI tools, GitHub Actions, CI/CD integrations, IDE plugins, or developer-facing utilities that make working with metadata effortless."*

---

## Author

**Shiva** — [@shivakumar2006](https://github.com/shivakumar2006)

BCA Final Year · 1700+ GitHub commits · Hackathon builder · Open source contributor
