package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"omctl/config"
	"omctl/internal/client"

	"github.com/spf13/cobra"
)

var whoownsCmd = &cobra.Command{
	Use:   "whoowns [table-name]",
	Short: "Find owner of a data asset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		c := client.New(cfg)

		query := args[0]

		resp, err := c.Get("search/query", map[string]string{
			"q":     query,
			"index": "table_search_index",
			"from":  "0",
			"size":  "1",
		})
		if err != nil {
			log.Fatal(err)
		}

		raw, _ := json.Marshal(resp)

		var result struct {
			Hits struct {
				Hits []struct {
					Source struct {
						FullyQualifiedName string `json:"fullyQualifiedName"`
						Description        string `json:"description"`
						Owner              *struct {
							Name        string `json:"name"`
							DisplayName string `json:"displayName"`
							Type        string `json:"type"`
						} `json:"owner"`
						Tags []struct {
							TagFQN string `json:"tagFQN"`
						} `json:"tags"`
						Domain *struct {
							DisplayName string `json:"displayName"`
						} `json:"domain"`
					} `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}

		json.Unmarshal(raw, &result)

		if len(result.Hits.Hits) == 0 {
			fmt.Printf("\n❌ No asset found for \"%s\"\n", query)
			return
		}

		s := result.Hits.Hits[0].Source

		fmt.Printf("\n📋 %s\n", s.FullyQualifiedName)
		fmt.Println("─────────────────────────────────────────")

		// Owner
		if s.Owner != nil {
			fmt.Printf("  👤 Owner  : %s (%s)\n", s.Owner.DisplayName, s.Owner.Type)
		} else {
			fmt.Printf("  👤 Owner  : ⚠️  no owner assigned\n")
		}

		// Description
		if s.Description != "" {
			fmt.Printf("  📝 Desc   : %s\n", stripHTML(s.Description))
		} else {
			fmt.Printf("  📝 Desc   : ⚠️  no description\n")
		}

		// Domain
		if s.Domain != nil {
			fmt.Printf("  🏢 Domain : %s\n", s.Domain.DisplayName)
		} else {
			fmt.Printf("  🏢 Domain : ⚠️  no domain\n")
		}

		// Tags
		if len(s.Tags) > 0 {
			fmt.Printf("  🏷️  Tags   : ")
			for i, t := range s.Tags {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(t.TagFQN)
			}
			fmt.Println()
		} else {
			fmt.Printf("  🏷️  Tags   : ⚠️  no tags\n")
		}

		fmt.Println()
	},
}

// strip basic HTML tags from description
func stripHTML(s string) string {
	result := ""
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result += string(r)
		}
	}
	return result
}

func init() {
	rootCmd.AddCommand(whoownsCmd)
}
