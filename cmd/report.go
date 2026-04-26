package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"omctl/config"
	"omctl/internal/client"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var outputFile string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a beautiful HTML governance report",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		c := client.New(cfg)

		fmt.Println("\n📊 Scanning OpenMetadata... please wait")

		type TableInfo struct {
			FQN       string
			HasOwner  bool
			HasDesc   bool
			HasTags   bool
			OwnerName string
			TagCount  int
		}

		var tables []TableInfo
		after := ""

		for {
			params := map[string]string{
				"limit":  "100",
				"fields": "owners,description,tags",
			}
			if after != "" {
				params["after"] = after
			}

			resp, err := c.Get("tables", params)
			if err != nil {
				log.Fatal(err)
			}

			raw, _ := json.Marshal(resp)

			var page struct {
				Data []struct {
					FullyQualifiedName string `json:"fullyQualifiedName"`
					Owners             []struct {
						Name string `json:"name"`
					} `json:"owners"`
					Description string `json:"description"`
					Tags        []struct {
						TagFQN string `json:"tagFQN"`
					} `json:"tags"`
				} `json:"data"`
				Paging struct {
					After string `json:"after"`
				} `json:"paging"`
			}

			json.Unmarshal(raw, &page)

			for _, t := range page.Data {
				info := TableInfo{
					FQN:      t.FullyQualifiedName,
					HasOwner: len(t.Owners) > 0,
					HasDesc:  t.Description != "",
					HasTags:  len(t.Tags) > 0,
					TagCount: len(t.Tags),
				}
				if len(t.Owners) > 0 {
					info.OwnerName = t.Owners[0].Name
				}
				tables = append(tables, info)
			}

			if page.Paging.After == "" {
				break
			}
			after = page.Paging.After
		}

		total := len(tables)
		hasOwner, hasDesc, hasTags := 0, 0, 0
		for _, t := range tables {
			if t.HasOwner {
				hasOwner++
			}
			if t.HasDesc {
				hasDesc++
			}
			if t.HasTags {
				hasTags++
			}
		}

		ownerPct := percent(hasOwner, total)
		descPct := percent(hasDesc, total)
		tagPct := percent(hasTags, total)
		score := (ownerPct + descPct + tagPct) / 3

		scoreColor := "#ef4444"
		scoreLabel := "Critical"
		scoreEmoji := "🔴"
		if score >= 70 {
			scoreColor = "#f59e0b"
			scoreLabel = "Moderate"
			scoreEmoji = "🟡"
		}
		if score >= 90 {
			scoreColor = "#22c55e"
			scoreLabel = "Healthy"
			scoreEmoji = "🟢"
		}

		tableRows := ""
		for _, t := range tables {
			ownerBadge := `<span class="badge badge-red">⚠ Missing</span>`
			if t.HasOwner {
				ownerBadge = `<span class="badge badge-green">` + t.OwnerName + `</span>`
			}
			descBadge := `<span class="badge badge-red">⚠ Missing</span>`
			if t.HasDesc {
				descBadge = `<span class="badge badge-green">✓ Yes</span>`
			}
			tagsBadge := `<span class="badge badge-red">⚠ Missing</span>`
			if t.HasTags {
				tagsBadge = fmt.Sprintf(`<span class="badge badge-green">%d tags</span>`, t.TagCount)
			}
			dot := "red"
			if t.HasOwner && t.HasDesc && t.HasTags {
				dot = "green"
			} else if t.HasOwner || t.HasDesc || t.HasTags {
				dot = "yellow"
			}
			ownerData := "missing"
			if t.HasOwner {
				ownerData = "ok"
			}
			descData := "missing"
			if t.HasDesc {
				descData = "ok"
			}
			tagsData := "missing"
			if t.HasTags {
				tagsData = "ok"
			}
			tableRows += fmt.Sprintf(`
			<tr data-owner="%s" data-desc="%s" data-tags="%s">
				<td><span class="dot %s"></span><code>%s</code></td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>`, ownerData, descData, tagsData, dot, t.FQN, ownerBadge, descBadge, tagsBadge)
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>omctl — Governance Report</title>
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap');
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#07070f;--surface:#0e0e1a;--surface2:#15151f;--border:#1e1e2e;
  --text:#e2e8f0;--muted:#4a5568;--accent:#6366f1;--accent2:#8b5cf6;
  --green:#22c55e;--yellow:#f59e0b;--red:#ef4444;
}
body{font-family:'Inter',sans-serif;background:var(--bg);color:var(--text);min-height:100vh;padding:40px 24px}
.wrap{max-width:1200px;margin:0 auto}

/* HEADER */
.header{display:flex;align-items:center;justify-content:space-between;margin-bottom:48px;padding-bottom:24px;border-bottom:1px solid var(--border)}
.brand{display:flex;align-items:center;gap:12px}
.brand-icon{width:42px;height:42px;background:linear-gradient(135deg,var(--accent),var(--accent2));border-radius:12px;display:flex;align-items:center;justify-content:center;font-size:22px;box-shadow:0 0 20px #6366f144}
.brand-name{font-size:22px;font-weight:800;letter-spacing:-0.5px}
.brand-name span{color:var(--accent)}
.meta{text-align:right}
.meta-time{font-family:'JetBrains Mono',monospace;font-size:12px;color:var(--muted)}
.meta-tag{font-size:11px;color:var(--accent);margin-top:2px;text-transform:uppercase;letter-spacing:1px}

/* HERO */
.hero{background:linear-gradient(135deg,#0d0d1f,#111128,#0d1829);border:1px solid var(--border);border-radius:24px;padding:48px;margin-bottom:28px;display:flex;align-items:center;gap:48px;position:relative;overflow:hidden}
.hero::before{content:'';position:absolute;inset:0;background:radial-gradient(ellipse at 20%% 50%%,#6366f10a,transparent 60%%),radial-gradient(ellipse at 80%% 20%%,#8b5cf608,transparent 50%%);pointer-events:none}
.ring{width:160px;height:160px;flex-shrink:0;position:relative}
.ring svg{transform:rotate(-90deg)}
.ring-num{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center}
.ring-score{font-size:44px;font-weight:800;line-height:1}
.ring-label{font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:1px}
.hero-copy h1{font-size:30px;font-weight:800;letter-spacing:-1px;margin-bottom:8px}
.hero-copy p{color:var(--muted);font-size:15px;margin-bottom:20px;line-height:1.6}
.pill{display:inline-flex;align-items:center;gap:8px;padding:6px 18px;border-radius:100px;font-size:13px;font-weight:600}

/* CARDS */
.cards{display:grid;grid-template-columns:repeat(3,1fr);gap:16px;margin-bottom:28px}
.card{background:var(--surface);border:1px solid var(--border);border-radius:18px;padding:24px 24px 20px;cursor:pointer;transition:transform .2s,border-color .2s,box-shadow .2s;position:relative;overflow:hidden}
.card:hover{transform:translateY(-3px);box-shadow:0 8px 30px #0008}
.card.active{border-color:var(--accent);box-shadow:0 0 0 1px var(--accent),0 8px 30px #6366f122}
.card-top{display:flex;align-items:center;justify-content:space-between;margin-bottom:16px}
.card-icon{font-size:18px}
.card-pct{font-size:32px;font-weight:800;line-height:1}
.card-title{font-size:12px;text-transform:uppercase;letter-spacing:1px;color:var(--muted);margin-bottom:4px}
.card-sub{font-size:12px;color:var(--muted)}
.bar{height:4px;background:var(--border);border-radius:100px;margin-top:14px;overflow:hidden}
.bar-fill{height:100%%;border-radius:100px;transition:width 1.2s cubic-bezier(.4,0,.2,1)}
.card.g .card-pct,.card.g .bar-fill{color:var(--green);background:var(--green)}
.card.y .card-pct,.card.y .bar-fill{color:var(--yellow);background:var(--yellow)}
.card.r .card-pct,.card.r .bar-fill{color:var(--red);background:var(--red)}
.card-shine{position:absolute;top:0;left:0;right:0;height:1px}
.card.g .card-shine{background:linear-gradient(90deg,transparent,var(--green),transparent)}
.card.y .card-shine{background:linear-gradient(90deg,transparent,var(--yellow),transparent)}
.card.r .card-shine{background:linear-gradient(90deg,transparent,var(--red),transparent)}

/* TABLE SECTION */
.tbox{background:var(--surface);border:1px solid var(--border);border-radius:18px;overflow:hidden}
.tbox-head{padding:20px 24px;border-bottom:1px solid var(--border);display:flex;align-items:center;justify-content:space-between;flex-wrap:gap;gap:12px}
.tbox-title{font-size:15px;font-weight:700}
.tbox-count{font-size:12px;color:var(--muted);font-family:'JetBrains Mono',monospace}
.actions{display:flex;gap:8px;flex-wrap:wrap;padding:16px 24px;border-bottom:1px solid var(--border)}
.btn{padding:7px 16px;border-radius:8px;font-size:12px;font-weight:600;cursor:pointer;border:none;font-family:'Inter',sans-serif;transition:all .15s}
.btn-pdf{background:linear-gradient(135deg,var(--accent),var(--accent2));color:#fff;box-shadow:0 2px 12px #6366f133}
.btn-pdf:hover{opacity:.85;transform:translateY(-1px)}
.btn-filter{background:var(--surface2);color:var(--muted);border:1px solid var(--border)}
.btn-filter:hover,.btn-filter.on{background:var(--accent);color:#fff;border-color:var(--accent)}
.searchbox{padding:12px 24px;border-bottom:1px solid var(--border)}
.searchbox input{width:100%%;padding:9px 14px;background:var(--surface2);border:1px solid var(--border);border-radius:8px;color:var(--text);font-size:13px;font-family:'JetBrains Mono',monospace;outline:none;transition:border-color .2s}
.searchbox input:focus{border-color:var(--accent)}
table{width:100%%;border-collapse:collapse}
thead th{padding:10px 20px;text-align:left;font-size:11px;text-transform:uppercase;letter-spacing:1px;color:var(--muted);background:var(--surface2);border-bottom:1px solid var(--border)}
tbody tr{border-bottom:1px solid var(--border);transition:background .1s}
tbody tr:last-child{border-bottom:none}
tbody tr:hover{background:var(--surface2)}
td{padding:10px 20px;font-size:12px;vertical-align:middle}
code{font-family:'JetBrains Mono',monospace;font-size:11px;color:#a5b4fc}
.badge{display:inline-block;padding:3px 10px;border-radius:100px;font-size:11px;font-weight:500}
.badge-green{background:#22c55e15;color:var(--green);border:1px solid #22c55e30}
.badge-red{background:#ef444415;color:var(--red);border:1px solid #ef444430}
.dot{display:inline-block;width:7px;height:7px;border-radius:50%%;margin-right:10px;flex-shrink:0}
.dot.green{background:var(--green);box-shadow:0 0 6px var(--green)}
.dot.yellow{background:var(--yellow);box-shadow:0 0 6px var(--yellow)}
.dot.red{background:var(--red);box-shadow:0 0 6px var(--red)}

/* FOOTER */
.footer{text-align:center;margin-top:36px;padding-top:20px;border-top:1px solid var(--border);font-size:12px;color:var(--muted)}
.footer b{color:var(--accent)}

/* PRINT */
@media print{
  body{background:#fff!important;color:#000!important;padding:20px}
  .actions,.searchbox,.header .meta-tag{display:none!important}
  .hero{background:#f8fafc!important;border:1px solid #e2e8f0!important}
  .tbox,.card{border:1px solid #e2e8f0!important;background:#fff!important}
  code{color:#4f46e5!important}
  .badge-green{background:#f0fdf4!important;color:#16a34a!important}
  .badge-red{background:#fef2f2!important;color:#dc2626!important}
  .dot.green{background:#22c55e!important}
  .dot.yellow{background:#f59e0b!important}
  .dot.red{background:#ef4444!important}
  thead th{background:#f8fafc!important;color:#64748b!important}
  tbody tr:hover{background:transparent!important}
}
</style>
</head>
<body>
<div class="wrap">

<!-- HEADER -->
<div class="header">
  <div class="brand">
    <div class="brand-icon">⚡</div>
    <div class="brand-name"><span>om</span>ctl</div>
  </div>
  <div class="meta">
    <div class="meta-time">%s</div>
    <div class="meta-tag">Governance Report</div>
  </div>
</div>

<!-- HERO -->
<div class="hero">
  <div class="ring">
    <svg width="160" height="160" viewBox="0 0 160 160">
      <circle cx="80" cy="80" r="68" fill="none" stroke="#1e1e2e" stroke-width="10"/>
      <circle cx="80" cy="80" r="68" fill="none" stroke="%s" stroke-width="10"
        stroke-dasharray="427" stroke-dashoffset="%d"
        stroke-linecap="round" style="transition:stroke-dashoffset 1.5s ease"/>
    </svg>
    <div class="ring-num">
      <div class="ring-score" style="color:%s" id="scoreNum">0</div>
      <div class="ring-label">/ 100</div>
    </div>
  </div>
  <div class="hero-copy">
    <h1>Governance Health Report</h1>
    <p>Scanned <strong>%d tables</strong> across your OpenMetadata instance.<br>
    Use <code>omctl</code> to fix issues directly from your terminal.</p>
    <div class="pill" style="background:%s22;border:1px solid %s44;color:%s">
      %s &nbsp;%s
    </div>
  </div>
</div>

<!-- CARDS -->
<div class="cards">
  <div class="card g" id="card-owner" onclick="filterBy('owner',this)">
    <div class="card-shine"></div>
    <div class="card-top">
      <span class="card-icon">👤</span>
      <span class="card-pct" data-target="%d">0%%</span>
    </div>
    <div class="card-title">Has Owner</div>
    <div class="card-sub">%d of %d tables</div>
    <div class="bar"><div class="bar-fill" style="width:0" data-w="%d"></div></div>
  </div>
  <div class="card y" id="card-desc" onclick="filterBy('desc',this)">
    <div class="card-shine"></div>
    <div class="card-top">
      <span class="card-icon">📝</span>
      <span class="card-pct" data-target="%d">0%%</span>
    </div>
    <div class="card-title">Has Description</div>
    <div class="card-sub">%d of %d tables</div>
    <div class="bar"><div class="bar-fill" style="width:0" data-w="%d"></div></div>
  </div>
  <div class="card r" id="card-tags" onclick="filterBy('tags',this)">
    <div class="card-shine"></div>
    <div class="card-top">
      <span class="card-icon">🏷️</span>
      <span class="card-pct" data-target="%d">0%%</span>
    </div>
    <div class="card-title">Has Tags</div>
    <div class="card-sub">%d of %d tables</div>
    <div class="bar"><div class="bar-fill" style="width:0" data-w="%d"></div></div>
  </div>
</div>

<!-- TABLE -->
<div class="tbox">
  <div class="tbox-head">
    <span class="tbox-title">All Tables</span>
    <span class="tbox-count" id="visibleCount">%d tables</span>
  </div>
  <div class="actions">
    <button class="btn btn-pdf" onclick="downloadPDF()">⬇ Download PDF</button>
    <button class="btn btn-filter on" id="btn-all" onclick="filterBy('all',this)">All</button>
    <button class="btn btn-filter" id="btn-owner" onclick="filterBy('owner',this)">Missing Owner</button>
    <button class="btn btn-filter" id="btn-desc" onclick="filterBy('desc',this)">Missing Desc</button>
    <button class="btn btn-filter" id="btn-tags" onclick="filterBy('tags',this)">Missing Tags</button>
    <button class="btn btn-filter" id="btn-critical" onclick="filterBy('critical',this)">🔴 Critical</button>
  </div>
  <div class="searchbox">
    <input type="text" id="search" placeholder="Search tables by name..." onkeyup="doSearch()"/>
  </div>
  <table id="tbl">
    <thead>
      <tr>
        <th>Table</th>
        <th>Owner</th>
        <th>Description</th>
        <th>Tags</th>
      </tr>
    </thead>
    <tbody>%s</tbody>
  </table>
</div>

<!-- FOOTER -->
<div class="footer">
  Generated by <b>omctl</b> &mdash; OpenMetadata CLI &mdash;
  Built for <b>Back to the Metadata Hackathon</b> &mdash; T-04 Developer Tooling
</div>

</div>
<script>
const totalRows = %d;
let currentFilter = 'all';
let searchTerm = '';

function updateCount() {
  const visible = document.querySelectorAll('#tbl tbody tr:not([style*="display: none"])').length;
  document.getElementById('visibleCount').textContent = visible + ' tables';
}

function applyFilters() {
  const rows = document.querySelectorAll('#tbl tbody tr');
  rows.forEach(row => {
    const fqn = row.querySelector('code').textContent.toLowerCase();
    const matchSearch = !searchTerm || fqn.includes(searchTerm);
    let matchFilter = true;
    if (currentFilter === 'owner')    matchFilter = row.dataset.owner === 'missing';
    if (currentFilter === 'desc')     matchFilter = row.dataset.desc === 'missing';
    if (currentFilter === 'tags')     matchFilter = row.dataset.tags === 'missing';
    if (currentFilter === 'critical') matchFilter = row.dataset.owner === 'missing' && row.dataset.desc === 'missing' && row.dataset.tags === 'missing';
    row.style.display = (matchSearch && matchFilter) ? '' : 'none';
  });
  updateCount();
}

function filterBy(type, el) {
  currentFilter = type;
  document.querySelectorAll('.btn-filter').forEach(b => b.classList.remove('on'));
  const btnId = 'btn-' + type;
  const btn = document.getElementById(btnId);
  if(btn) btn.classList.add('on');
  document.querySelectorAll('.card').forEach(c => c.classList.remove('active'));
  if(type !== 'all' && type !== 'critical') {
    const card = document.getElementById('card-' + type);
    if(card) card.classList.add('active');
  }
  applyFilters();
}

function doSearch() {
  searchTerm = document.getElementById('search').value.toLowerCase();
  applyFilters();
}

function downloadPDF() {
  const btn = event.target;
  btn.textContent = '⏳ Preparing...';
  btn.disabled = true;
  setTimeout(() => { window.print(); btn.textContent = '⬇ Download PDF'; btn.disabled = false; }, 400);
}

// Animated counters
function animCount(el, target) {
  let n = 0;
  const step = Math.max(1, Math.ceil(target / 50));
  const t = setInterval(() => {
    n = Math.min(n + step, target);
    el.textContent = n + '%%';
    if(n >= target) clearInterval(t);
  }, 25);
}

// Animated score ring
function animScore(target) {
  const el = document.getElementById('scoreNum');
  let n = 0;
  const t = setInterval(() => {
    n = Math.min(n + 2, target);
    el.textContent = n;
    if(n >= target) clearInterval(t);
  }, 20);
}

window.addEventListener('load', () => {
  // Counter animations
  document.querySelectorAll('.card-pct[data-target]').forEach(el => {
    animCount(el, parseInt(el.dataset.target));
  });
  // Bar animations
  document.querySelectorAll('.bar-fill[data-w]').forEach(el => {
    setTimeout(() => el.style.width = el.dataset.w + '%%', 100);
  });
  // Score animation
  animScore(%d);
  updateCount();
});
</script>
</body>
</html>`,
			time.Now().Format("02 Jan 2006 · 15:04 MST"),
			scoreColor,
			int(427-(float64(score)/100.0)*427),
			scoreColor,
			total,
			scoreColor, scoreColor, scoreColor,
			scoreEmoji, scoreLabel,
			ownerPct, hasOwner, total, ownerPct,
			descPct, hasDesc, total, descPct,
			tagPct, hasTags, total, tagPct,
			total,
			tableRows,
			total,
			score,
		)

		if outputFile == "" {
			outputFile = "governance-report.html"
		}
		if err := os.WriteFile(outputFile, []byte(html), 0644); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\n✅ Report generated!\n")
		fmt.Printf("   File   : %s\n", outputFile)
		fmt.Printf("   Tables : %d scanned\n", total)
		fmt.Printf("   Score  : %d/100 %s\n\n", score, scoreEmoji)
		fmt.Printf("   open %s\n\n", outputFile)
	},
}

func init() {
	reportCmd.Flags().StringVarP(&outputFile, "output", "o", "governance-report.html", "Output file path")
	rootCmd.AddCommand(reportCmd)
}
