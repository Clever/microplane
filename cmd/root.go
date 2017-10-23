package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var workDir string

var rootCmd = &cobra.Command{
	Use:   "mp",
	Short: "Microplane makes git changes across many repos",
}

func init() {
	rootCmd.PersistentFlags().StringP("repo", "r", "", "single repo to operate on")
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mergeCmd)

	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVarP(&planFlagBranch, "branch", "b", "", "Git branch to commit to")
	planCmd.Flags().StringVarP(&planFlagMessage, "message", "m", "", "Commit message")

	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushFlagAssignee, "assignee", "a", "", "Github user to assign the PR to")

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
