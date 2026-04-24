package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"omctl/config"
	"omctl/internal/client"

	"github.com/spf13/cobra"
)

var (
	filterNoOwner bool
	filterNoDesc  bool
	filterNoTags  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list table with optional governance filters",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		c := client.New(cfg)

		after := ""
		count := 0

		fmt.Println()

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

			var Pages struct {
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

			json.Unmarshal(raw, &Pages)

			for _, t := range Pages.Data {
				noOwner := len(t.Owners) == 0
				noDesc := t.Description == ""
				noTags := len(t.Tags) == 0

				if filterNoOwner && !noOwner {
					continue
				}
				if filterNoDesc && !noDesc {
					continue
				}
				if filterNoTags && !noTags {
					continue
				}

				issues := ""
				if noOwner {
					issues += "👤 "
				}
				if noDesc {
					issues += "📝 "
				}
				if noTags {
					issues += "🏷️  "
				}
				if issues == "" {
					issues = "✅"
				}

				fmt.Printf("  %s %s\n", issues, t.FullyQualifiedName)
				count++
			}

			if Pages.Paging.After == "" {
				break
			}
			after = Pages.Paging.After
		}

		fmt.Printf("\n  Total: %d tables\n\n", count)
	},
}

func init() {
	listCmd.Flags().BoolVar(&filterNoOwner, "no-owner", false, "Show only tables missing owner")
	listCmd.Flags().BoolVar(&filterNoDesc, "no-desc", false, "Show only tables missing description")
	listCmd.Flags().BoolVar(&filterNoTags, "no-tags", false, "Show only tables missing tags")
	rootCmd.AddCommand(listCmd)
}
