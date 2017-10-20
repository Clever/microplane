package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs [path]",
	Short: "Generates markdown docs for each command",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "docs/"
		if len(args) == 1 {
			path = os.Args[0]
		}

		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatal(err)
		}

		err := doc.GenMarkdownTree(rootCmd, path)
		if err != nil {
			log.Fatal(err)
		}
	},
}
