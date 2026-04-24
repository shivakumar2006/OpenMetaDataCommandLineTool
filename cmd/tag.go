package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"omctl/config"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag [table-fqn] [tagFQN]",
	Short: "Add a tag to a table in OpenMetadata",
	Long:  `Add a classification tag to any table. Example: omctl tag sample_redshift.staging_db.dim_customer PII.Sensitive`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		tableFQN := args[0]
		tagFQN := args[1]

		fmt.Printf("\n🏷️  Tagging \"%s\" with \"%s\"...\n", tableFQN, tagFQN)

		// Step 1: table ID fetch karo FQN se
		url := fmt.Sprintf("%s/api/v1/tables/name/%s?fields=id,tags", cfg.Host, tableFQN)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
		req.Header.Set("Content-Type", "application/json")

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		var table struct {
			ID   string `json:"id"`
			Tags []struct {
				TagFQN    string `json:"tagFQN"`
				Source    string `json:"source"`
				LabelType string `json:"labelType"`
			} `json:"tags"`
		}
		json.NewDecoder(resp.Body).Decode(&table)

		if table.ID == "" {
			fmt.Printf("❌ Table not found: %s\n", tableFQN)
			return
		}

		// Step 2: existing tags mein naya tag add karo
		// duplicate check
		for _, t := range table.Tags {
			if t.TagFQN == tagFQN {
				fmt.Printf("⚠️  Tag \"%s\" already exists on this table\n\n", tagFQN)
				return
			}
		}

		newTag := map[string]string{
			"tagFQN":    tagFQN,
			"source":    "Classification",
			"labelType": "Manual",
		}

		// PATCH body banao
		patch := []map[string]any{
			{
				"op":    "add",
				"path":  "/tags/-",
				"value": newTag,
			},
		}

		patchBytes, _ := json.Marshal(patch)

		// Step 3: PATCH request bhejo
		patchURL := fmt.Sprintf("%s/api/v1/tables/%s", cfg.Host, table.ID)
		patchReq, _ := http.NewRequest("PATCH", patchURL, bytes.NewBuffer(patchBytes))
		patchReq.Header.Set("Authorization", "Bearer "+cfg.Token)
		patchReq.Header.Set("Content-Type", "application/json-patch+json")

		patchResp, err := httpClient.Do(patchReq)
		if err != nil {
			log.Fatal(err)
		}
		defer patchResp.Body.Close()

		if patchResp.StatusCode == 200 {
			fmt.Printf("✅ Successfully tagged!\n")
			fmt.Printf("   Table : %s\n", tableFQN)
			fmt.Printf("   Tag   : %s\n\n", tagFQN)
		} else {
			var errResp map[string]any
			json.NewDecoder(patchResp.Body).Decode(&errResp)
			fmt.Printf("❌ Failed: %v\n", errResp)
		}
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
