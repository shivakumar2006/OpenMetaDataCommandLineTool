# omctl — OpenMetadata for Your Terminal

> **"47% of knowledge workers can't find the data they need."**  
> `omctl` fixes that — without leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.21-blue?logo=go)
![OpenMetadata](https://img.shields.io/badge/OpenMetadata-1.12-orange)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Track](https://img.shields.io/badge/Track-T--04%20Developer%20Tooling-purple)

---

## The Problem

Data engineers context-switch **dozens of times a day** — jumping between terminal and OpenMetadata UI just to answer simple questions:

- *Who owns this table?*
- *Where does this data come from?*
- *Which tables have no owner assigned?*
- *How healthy is our metadata overall?*

Every UI switch breaks flow. `omctl` eliminates that friction.

---

## Demo

```bash
# Find any data asset instantly
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
  📝 Desc   : ตารางข้อมูลลูกค้า
  🏢 Domain : ⚠️  no domain
  🏷️  Tags   : PII.Sensitive, DataTier.Gold

# Trace data lineage
$ omctl lineage "dim_customer"

📋 sample_redshift.staging_db.integration.dim_customer — Lineage
─────────────────────────────────────────
  ⬆️  UPSTREAM (data comes from)
      └── stg_customers
      └── raw_customers

  ⬇️  DOWNSTREAM (data goes to)
      └── customer_360

# Scan 1276 tables — governance health in seconds
$ omctl health

📊 OpenMetadata Health Report
─────────────────────────────────────────
  Total Tables    : 1276

  👤 Has Owner    : 844/1276 (66%)
  📝 Has Desc     : 920/1276 (72%)
  🏷️  Has Tags     : 731/1276 (57%)

  Overall Score   : 65/100 🔴

# Tag a table directly from terminal
$ omctl tag "sample_redshift.staging_db.integration.dim_customer" "PII.Sensitive"

✅ Successfully tagged!
   Table : sample_redshift.staging_db.integration.dim_customer
   Tag   : PII.Sensitive

# List tables with governance issues
$ omctl list --no-owner
$ omctl list --no-tags
$ omctl list --no-owner --no-desc
```

---

## Commands

| Command | Description |
|---------|-------------|
| `omctl search <query>` | Search any data asset across the catalog |
| `omctl whoowns <table>` | Get owner, description, domain and tags |
| `omctl lineage <table>` | Show upstream and downstream lineage tree |
| `omctl health` | Scan all tables and show governance score |
| `omctl tag <table> <tag>` | Add a classification tag to a table |
| `omctl list` | List all tables with governance status |
| `omctl list --no-owner` | Filter tables missing owner |
| `omctl list --no-desc` | Filter tables missing description |
| `omctl list --no-tags` | Filter tables missing tags |

---

## Installation

### Prerequisites
- Go 1.21+
- OpenMetadata instance (cloud or sandbox)
- Personal Access Token from OpenMetadata

### Install

```bash
git clone https://github.com/shivakumar2006/omctl
cd omctl
go install .
```

### Configure

```bash
export OM_HOST=https://your-instance.open-metadata.org
export OM_TOKEN=your_personal_access_token
```

Or add to `~/.zshrc` / `~/.bashrc` for permanent setup.

### Quick Start

```bash
# Test connection
omctl health

# Search your catalog
omctl search "orders"

# Check ownership
omctl whoowns "orders"
```

---

## Why omctl?

OpenMetadata has a powerful UI — but developers live in the terminal.

| Without omctl | With omctl |
|---------------|------------|
| Open browser | Stay in terminal |
| Navigate to OpenMetadata | `omctl search "table"` |
| Click through 3 pages | `omctl whoowns "table"` |
| View lineage graph | `omctl lineage "table"` |
| Export report manually | `omctl health` |
| Tag via UI form | `omctl tag "table" "PII.Sensitive"` |

`omctl` is to OpenMetadata what `kubectl` is to Kubernetes — a **first-class CLI** that makes metadata management effortless for developers.

---

## Architecture

```
omctl/
├── cmd/           # CLI commands (cobra)
│   ├── search.go
│   ├── whoowns.go
│   ├── lineage.go
│   ├── health.go
│   ├── tag.go
│   └── list.go
├── internal/
│   ├── client/    # OpenMetadata REST API client
│   └── display/   # Terminal output formatting
└── config/        # Environment config loader
```

**Stack:** Go · Cobra CLI · OpenMetadata REST API

No external dependencies beyond standard library and Cobra. Single binary, zero config files required.

---

## OpenMetadata Integration

`omctl` uses OpenMetadata's REST API deeply:

| API Endpoint | Used By |
|-------------|---------|
| `GET /api/v1/search/query` | `search`, `whoowns`, `lineage` |
| `GET /api/v1/tables` | `health`, `list` |
| `GET /api/v1/lineage/table/{id}` | `lineage` |
| `GET /api/v1/tables/name/{fqn}` | `tag` |
| `PATCH /api/v1/tables/{id}` | `tag` (write-back) |

---

## Real Impact

Tested against OpenMetadata sandbox with **1,276 real tables**:

- Scanned all 1,276 tables in under 10 seconds
- Identified 432 tables missing owners (34%)
- Identified 356 tables missing descriptions (28%)
- Identified 545 tables missing tags (43%)
- Overall governance score: **65/100**

This is the kind of insight that previously required navigating multiple UI pages — now available in a single command.

---

## Built For

**Back to the Metadata Hackathon** — Paradox #T-04: Developer Tooling & CI/CD

> *"Build CLI tools, GitHub Actions, CI/CD integrations, IDE plugins, or developer-facing utilities that make working with metadata effortless."*

---

## Author

**Shiva** — [@shivakumar2006](https://github.com/shivakumar2006)

BCA Final Year · Hackathon builder · Open source contributor
