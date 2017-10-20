package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var workDir string

var rootCmd = &cobra.Command{
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
	rootCmd.PersistentFlags().StringP("repo", "r", "", "single repo to operate on")
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(statusCmd)

	workDir, _ = filepath.Abs("./mp")

	// Create workDir, if doesn't yet exist
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		if err := os.Mkdir(workDir, 0755); err != nil {
			log.Fatalf("error creating workDir: %s\n", err.Error())
		}
	}
}

// Execute starts the CLI
func Execute() error {
	return rootCmd.Execute()
}
