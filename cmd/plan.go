package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "plan short description",
	Long: `plan
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("plan")
	},
}
