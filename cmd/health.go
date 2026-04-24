package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"omctl/config"
	"omctl/internal/client"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show governance health of your OpenMetadata instance",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		c := client.New(cfg)

		fmt.Println("\n📊 Scanning OpenMetadata... please wait")

		total := 0
		missingOwner := 0
		missingDesc := 0
		noTags := 0
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
					Owners []struct {
						Name string `json:"name"`
					} `json:"owners"`
					Description string `json:"description"`
					Tags        []struct {
						TagFQN string `json:"tagFQN"`
					} `json:"tags"`
				} `json:"data"`
				Paging struct {
					After string `json:"after"`
					Total int    `json:"total"`
				} `json:"paging"`
			}

			json.Unmarshal(raw, &page)

			for _, t := range page.Data {
				total++
				if len(t.Owners) == 0 {
					missingOwner++
				}
				if t.Description == "" {
					missingDesc++
				}
				if len(t.Tags) == 0 {
					noTags++
				}
			}

			if page.Paging.After == "" {
				break
			}
			after = page.Paging.After
		}

		if total == 0 {
			fmt.Println("❌ No tables found")
			return
		}

		ownerPct := percent(total-missingOwner, total)
		descPct := percent(total-missingDesc, total)
		tagPct := percent(total-noTags, total)
		score := (ownerPct + descPct + tagPct) / 3

		scoreEmoji := "🔴"
		if score >= 70 {
			scoreEmoji = "🟡"
		}
		if score >= 90 {
			scoreEmoji = "🟢"
		}

		fmt.Printf("\n📊 OpenMetadata Health Report\n")
		fmt.Println("─────────────────────────────────────────")
		fmt.Printf("  Total Tables    : %d\n\n", total)
		fmt.Printf("  👤 Has Owner    : %d/%d (%d%%)\n", total-missingOwner, total, ownerPct)
		fmt.Printf("  📝 Has Desc     : %d/%d (%d%%)\n", total-missingDesc, total, descPct)
		fmt.Printf("  🏷️  Has Tags     : %d/%d (%d%%)\n", total-noTags, total, tagPct)
		fmt.Printf("\n  Overall Score   : %d/100 %s\n\n", score, scoreEmoji)
	},
}

func percent(have, total int) int {
	if total == 0 {
		return 0
	}
	return (have * 100) / total
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
