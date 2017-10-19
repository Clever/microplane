package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge short description",
	Long: `merge
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("merge")
	},
}
