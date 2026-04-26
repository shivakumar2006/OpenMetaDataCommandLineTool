package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"omctl/config"
	"omctl/internal/client"
	"strings"

	"github.com/spf13/cobra"
)

var (
	fixNoOwner      bool
	fixNoDesc       bool
	fixNoTags       bool
	fixAll          bool
	fixDryRun       bool
	fixAutoClassify bool
	fixDefaultOwner string
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Auto-fix governance issues across your OpenMetadata tables",
	Long:  `Scan and fix missing owners, descriptions, and PII tags across all tables. Use --dry-run to preview changes first.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		c := client.New(cfg)

		if fixAll {
			fixNoOwner = true
			fixNoTags = true
		}

		if !fixNoOwner && !fixNoTags && !fixNoDesc {
			fmt.Println("\n❌ Specify what to fix:")
			fmt.Println("   --no-owner       Fix missing owners")
			fmt.Println("   --no-tags        Auto-classify PII tags")
			fmt.Println("   --no-desc        Add placeholder descriptions")
			fmt.Println("   --all            Fix everything")
			fmt.Println("   --dry-run        Preview without applying")
			fmt.Println()
			return
		}

		if fixDryRun {
			fmt.Println("\n🔍 DRY RUN — no changes will be applied\n")
		}

		fmt.Println("\n🔧 Scanning tables for issues...")

		type TableRow struct {
			ID                 string
			FullyQualifiedName string
			HasOwner           bool
			HasDesc            bool
			HasTags            bool
			Columns            []struct {
				Name     string `json:"name"`
				DataType string `json:"dataType"`
			}
		}

		var tables []TableRow
		after := ""

		for {
			params := map[string]string{
				"limit":  "100",
				"fields": "owners,description,tags,columns",
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
					ID                 string `json:"id"`
					FullyQualifiedName string `json:"fullyQualifiedName"`
					Owners             []struct {
						Name string `json:"name"`
					} `json:"owners"`
					Description string `json:"description"`
					Tags        []struct {
						TagFQN string `json:"tagFQN"`
					} `json:"tags"`
					Columns []struct {
						Name     string `json:"name"`
						DataType string `json:"dataType"`
					} `json:"columns"`
				} `json:"data"`
				Paging struct {
					After string `json:"after"`
				} `json:"paging"`
			}

			json.Unmarshal(raw, &page)

			for _, t := range page.Data {
				tables = append(tables, TableRow{
					ID:                 t.ID,
					FullyQualifiedName: t.FullyQualifiedName,
					HasOwner:           len(t.Owners) > 0,
					HasDesc:            t.Description != "",
					HasTags:            len(t.Tags) > 0,
					Columns:            t.Columns,
				})
			}

			if page.Paging.After == "" {
				break
			}
			after = page.Paging.After
		}

		// Count issues
		ownerIssues, tagIssues, descIssues := 0, 0, 0
		for _, t := range tables {
			if !t.HasOwner {
				ownerIssues++
			}
			if !t.HasTags {
				tagIssues++
			}
			if !t.HasDesc {
				descIssues++
			}
		}

		fmt.Printf("📊 Issues found:\n")
		if fixNoOwner {
			fmt.Printf("   👤 Missing owner : %d tables\n", ownerIssues)
		}
		if fixNoTags {
			fmt.Printf("   🏷️  Missing tags  : %d tables\n", tagIssues)
		}
		if fixNoDesc {
			fmt.Printf("   📝 Missing desc  : %d tables\n", descIssues)
		}
		fmt.Println()

		if fixDryRun {
			fmt.Println("── DRY RUN PREVIEW ──────────────────────────")
			for _, t := range tables {
				issues := []string{}
				if fixNoOwner && !t.HasOwner {
					issues = append(issues, "owner→bot-service")
				}
				if fixNoTags && !t.HasTags {
					issues = append(issues, "tags→"+detectPIITags(t.Columns))
				}
				if fixNoDesc && !t.HasDesc {
					issues = append(issues, "desc→auto-generated")
				}
				if len(issues) > 0 {
					fmt.Printf("  📋 %s\n     → %s\n\n", t.FullyQualifiedName, strings.Join(issues, " | "))
				}
			}
			fmt.Println("─────────────────────────────────────────────")
			fmt.Println("✅ Dry run complete. Remove --dry-run to apply.")
			fmt.Println()
			return
		}

		// Apply fixes
		fixed := 0
		failed := 0
		httpClient := &http.Client{}

		for _, t := range tables {
			patched := false

			// Fix tags — auto PII classify
			if fixNoTags && !t.HasTags {
				detectedTag := detectPIITags(t.Columns)
				if detectedTag != "PII.None" {
					patch := []map[string]any{{
						"op":   "add",
						"path": "/tags/-",
						"value": map[string]string{
							"tagFQN":    detectedTag,
							"source":    "Classification",
							"labelType": "Automated",
						},
					}}
					if applyPatch(httpClient, cfg.Host, cfg.Token, t.ID, patch) {
						fmt.Printf("  🏷️  Tagged  : %s → %s\n", t.FullyQualifiedName, detectedTag)
						patched = true
					} else {
						failed++
					}
				}
			}

			// Fix description
			if fixNoDesc && !t.HasDesc {
				parts := strings.Split(t.FullyQualifiedName, ".")
				tableName := parts[len(parts)-1]
				autoDesc := fmt.Sprintf("Table: %s. Auto-documented by omctl. Review and update with accurate description.", tableName)
				patch := []map[string]any{{
					"op":    "add",
					"path":  "/description",
					"value": autoDesc,
				}}
				if applyPatch(httpClient, cfg.Host, cfg.Token, t.ID, patch) {
					fmt.Printf("  📝 Desc    : %s → auto-generated\n", t.FullyQualifiedName)
					patched = true
				} else {
					failed++
				}
			}

			if patched {
				fixed++
			}
		}

		fmt.Println()
		fmt.Println("─────────────────────────────────────────────")
		fmt.Printf("✅ Fixed  : %d tables\n", fixed)
		if failed > 0 {
			fmt.Printf("❌ Failed : %d tables\n", failed)
		}
		fmt.Println()
	},
}

// detectPIITags — column names se PII type detect karo
func detectPIITags(columns []struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}) string {
	piiKeywords := map[string]string{
		"email":       "PII.Email",
		"mail":        "PII.Email",
		"phone":       "PII.Phone",
		"mobile":      "PII.Phone",
		"ssn":         "PII.SSN",
		"aadhar":      "PII.SSN",
		"passport":    "PII.Sensitive",
		"password":    "PII.Sensitive",
		"credit_card": "PII.BankingInformation",
		"card_number": "PII.BankingInformation",
		"dob":         "PII.DOB",
		"birth":       "PII.DOB",
		"address":     "PII.Location",
		"location":    "PII.Location",
		"ip_address":  "PII.Location",
		"name":        "PII.Name",
		"first_name":  "PII.Name",
		"last_name":   "PII.Name",
		"salary":      "PII.Sensitive",
		"income":      "PII.Sensitive",
		"gender":      "PII.Sensitive",
		"race":        "PII.Sensitive",
	}

	for _, col := range columns {
		colLower := strings.ToLower(col.Name)
		for keyword, tag := range piiKeywords {
			if strings.Contains(colLower, keyword) {
				return tag
			}
		}
	}
	return "PII.None"
}

// applyPatch — PATCH request bhejo OpenMetadata ko
func applyPatch(httpClient *http.Client, host, token, tableID string, patch []map[string]any) bool {
	patchBytes, _ := json.Marshal(patch)
	url := fmt.Sprintf("%s/api/v1/tables/%s", host, tableID)
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(patchBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json-patch+json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func init() {
	fixCmd.Flags().BoolVar(&fixNoOwner, "no-owner", false, "Fix tables missing owners")
	fixCmd.Flags().BoolVar(&fixNoDesc, "no-desc", false, "Add auto-generated descriptions")
	fixCmd.Flags().BoolVar(&fixNoTags, "no-tags", false, "Auto-classify PII tags from column names")
	fixCmd.Flags().BoolVar(&fixAll, "all", false, "Fix all governance issues")
	fixCmd.Flags().BoolVar(&fixDryRun, "dry-run", false, "Preview changes without applying")
	fixCmd.Flags().StringVar(&fixDefaultOwner, "owner", "bot-service", "Default owner name for --no-owner")
	rootCmd.AddCommand(fixCmd)
}
