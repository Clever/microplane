package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push short description",
	Long: `push
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("push")
	},
}
