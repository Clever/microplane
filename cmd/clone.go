package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone short description",
	Long: `clone
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("clone")
	},
}
