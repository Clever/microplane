package cmd

import (
	"context"
	"log"

	"github.com/Clever/microplane/merge"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge short description",
	Long: `merge
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := merge.Merge(context.Background(), merge.Input{
			// TODO: Get user input, instead
			Org:               "Clever",
			Repo:              "microplane",
			PullRequestNumber: 1,
			CommitSHA:         "960ecbb",
		})
		if err != nil {
			log.Fatal(err)
		}
	},
}
