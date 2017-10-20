package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/plan"
	"github.com/Clever/microplane/push"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status short description",
	Long: `status
                long
                description`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		// find files and folders to explain the status of each repo
		initPath := initOutputPath(target)

		if _, err := os.Stat(initPath); os.IsNotExist(err) {
			log.Fatalf("target not found: %s\n", err.Error())
		}

		var initOutput initialize.Output
		if err := loadJSON(initPath, &initOutput); err != nil {
			log.Fatalf("error loading init.json: %s\n", err.Error())
		}

		// TODO: pretty print status
		fmt.Printf("%40s    %20s    %20s\n", "repo", "status", "details")
		fmt.Printf("%40s    %20s    %20s\n", "====", "======", "=======")
		for _, r := range initOutput.Repos {
			status := getRepoStatus(target, r.Name)
			fmt.Printf("%40s    %20s    %20s\n", r.Name, status, "...")
		}
	},
}

func getRepoStatus(target, repo string) string {
	status := "initialized"
	var cloneOutput clone.Output
	if !(loadJSON(cloneOutputPath(target, repo), &cloneOutput) == nil && cloneOutput.Success) {
		return status
	}
	status = "cloned"

	var planOutput plan.Output
	if !(loadJSON(planOutputPath(target, repo), &planOutput) == nil && planOutput.Success) {
		return status
	}
	status = "planned"

	var pushOutput push.Output
	if !(loadJSON(pushOutputPath(target, repo), &pushOutput) == nil && pushOutput.Success) {
		return status
	}
	status = "pushed"

	var mergeOutput merge.Output
	if !(loadJSON(mergeOutputPath(target, repo), &mergeOutput) == nil && mergeOutput.Success) {
		return status
	}
	status = "merged"

	return status
}
