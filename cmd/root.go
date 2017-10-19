package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "microplane",
	Short: "microplane",
	Long: `Microplane
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Microplane RootCmd")
	},
}

func init() {
	RootCmd.PersistentFlags().StringP("repo", "r", "", "single repo to operate on")
	RootCmd.AddCommand(cloneCmd, initCmd)
}

func Execute() error {
	return RootCmd.Execute()
}
