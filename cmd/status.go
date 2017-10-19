package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status short description",
	Long: `status
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("status")
	},
}
