package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "mp init",
	Short: "init short description",
	Long: `init
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init")
	},
}
