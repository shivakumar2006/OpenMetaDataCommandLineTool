package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"omctl/config"
	"omctl/internal/client"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search data assets in OpenMetadata",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		c := client.New(cfg)

		query := args[0]

		resp, err := c.Get("search/query", map[string]string{
			"q":     query,
			"index": "table_search_index",
			"from":  "0",
			"size":  "10",
		})
		if err != nil {
			log.Fatal(err)
		}

		// parse hits
		raw, _ := json.Marshal(resp)

		var result struct {
			Hits struct {
				Total struct {
					Value int `json:"value"`
				} `json:"total"`
				Hits []struct {
					Source struct {
						Name               string `json:"name"`
						FullyQualifiedName string `json:"fullyQualifiedName"`
						Description        string `json:"description"`
						Owner              *struct {
							Name string `json:"name"`
						} `json:"owner"`
					} `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}

		json.Unmarshal(raw, &result)

		fmt.Printf("\n🔍 %d results for \"%s\"\n\n", result.Hits.Total.Value, query)

		for _, hit := range result.Hits.Hits {
			s := hit.Source
			owner := "no owner"
			if s.Owner != nil {
				owner = s.Owner.Name
			}
			desc := s.Description
			if desc == "" {
				desc = "no description"
			}
			fmt.Printf("  📋 %s\n", s.FullyQualifiedName)
			fmt.Printf("     Owner : %s\n", owner)
			fmt.Printf("     Desc  : %s\n\n", desc)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
