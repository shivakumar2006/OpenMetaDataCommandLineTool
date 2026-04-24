package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"omctl/config"
	"omctl/internal/client"

	"github.com/spf13/cobra"
)

var lineageCmd = &cobra.Command{
	Use:   "lineage [table-name]",
	Short: "Show upstream and downstream lineage of a table",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		c := client.New(cfg)

		query := args[0]

		// Step 1: table ka FQN aur ID dhundo
		searchResp, err := c.Get("search/query", map[string]string{
			"q":     query,
			"index": "table_search_index",
			"from":  "0",
			"size":  "1",
		})
		if err != nil {
			log.Fatal(err)
		}

		raw, _ := json.Marshal(searchResp)
		var searchResult struct {
			Hits struct {
				Hits []struct {
					Source struct {
						ID                 string `json:"id"`
						FullyQualifiedName string `json:"fullyQualifiedName"`
					} `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}
		json.Unmarshal(raw, &searchResult)

		if len(searchResult.Hits.Hits) == 0 {
			fmt.Printf("\n❌ No table found for \"%s\"\n", query)
			return
		}

		tableID := searchResult.Hits.Hits[0].Source.ID
		tableFQN := searchResult.Hits.Hits[0].Source.FullyQualifiedName

		// Step 2: lineage fetch karo
		lineageResp, err := c.Get(fmt.Sprintf("lineage/table/%s", tableID), map[string]string{
			"upstreamDepth":   "2",
			"downstreamDepth": "2",
		})
		if err != nil {
			log.Fatal(err)
		}

		lineageRaw, _ := json.Marshal(lineageResp)
		var lineage struct {
			Entity struct {
				Name string `json:"name"`
			} `json:"entity"`
			Nodes []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				DisplayName string `json:"displayName"`
				Type        string `json:"type"`
			} `json:"nodes"`
			UpstreamEdges []struct {
				FromEntity string `json:"fromEntity"`
				ToEntity   string `json:"toEntity"`
			} `json:"upstreamEdges"`
			DownstreamEdges []struct {
				FromEntity string `json:"fromEntity"`
				ToEntity   string `json:"toEntity"`
			} `json:"downstreamEdges"`
		}
		json.Unmarshal(lineageRaw, &lineage)

		// ID se name map banao
		idToName := map[string]string{}
		idToName[tableID] = tableFQN
		for _, n := range lineage.Nodes {
			name := n.DisplayName
			if name == "" {
				name = n.Name
			}
			idToName[n.ID] = name
		}

		fmt.Printf("\n📋 %s — Lineage\n", tableFQN)
		fmt.Println("─────────────────────────────────────────")

		// Upstream
		fmt.Println("\n  ⬆️  UPSTREAM (data comes from)")
		if len(lineage.UpstreamEdges) == 0 {
			fmt.Println("      └── no upstream found")
		}
		for _, e := range lineage.UpstreamEdges {
			from := idToName[e.FromEntity]
			if from == "" {
				from = e.FromEntity
			}
			fmt.Printf("      └── %s\n", from)
		}

		// Downstream
		fmt.Println("\n  ⬇️  DOWNSTREAM (data goes to)")
		if len(lineage.DownstreamEdges) == 0 {
			fmt.Println("      └── no downstream found")
		}
		for _, e := range lineage.DownstreamEdges {
			to := idToName[e.ToEntity]
			if to == "" {
				to = e.ToEntity
			}
			fmt.Printf("      └── %s\n", to)
		}

		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(lineageCmd)
}
