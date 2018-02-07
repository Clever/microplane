package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/plan"
	"github.com/spf13/cobra"
)

var planFlagBranch string
var planFlagMessage string

// TODO: Pass these *not* via globals
// these variables are set when the cmd starts running
var (
	branchName    string
	commitMessage string
	changeCmd     string
	changeCmdArgs []string
	isSingleRepo  bool
)

var planCmd = &cobra.Command{
	Use:   "plan [cmd] [args...]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Plan changes by running a command against cloned repos",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		changeCmd = args[0]
		if len(args) > 1 {
			changeCmdArgs = args[1:]
		}

		branchName, err = cmd.Flags().GetString("branch")
		if err != nil {
			log.Fatal(err)
		}
		if branchName == "" {
			log.Fatal("--branch is required")
		}

		commitMessage, err = cmd.Flags().GetString("message")
		if err != nil {
			log.Fatal(err)
		}
		if commitMessage == "" {
			log.Fatal("--message is required")
		}

		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}
		isSingleRepo = len(repos) == 1

		err = parallelize(repos, planOneRepo)
		if err != nil {
			log.Fatal("# errors = ", strings.Count(err.Error(), " | ")+1)
		}
	},
}

func planOneRepo(r initialize.Repo, ctx context.Context) error {
	log.Printf("planning: %s/%s", r.Owner, r.Name)

	// Get previous step's output
	var cloneOutput clone.Output
	if loadJSON(outputPath(r.Name, "clone"), &cloneOutput) != nil || !cloneOutput.Success {
		log.Printf("skipping %s/%s, must successfully clone first", r.Owner, r.Name)
		return nil
	}

	// Prepare workdir for current step's output
	planOutputPath := outputPath(r.Name, "plan")
	planWorkDir := filepath.Dir(planOutputPath)
	if err := os.MkdirAll(planWorkDir, 0755); err != nil {
		return err
	}

	// Execute
	input := plan.Input{
		RepoName:      r.Name,
		RepoDir:       cloneOutput.ClonedIntoDir,
		WorkDir:       planWorkDir,
		Command:       plan.Command{Path: changeCmd, Args: changeCmdArgs},
		CommitMessage: commitMessage,
		BranchName:    branchName,
	}
	output, err := plan.Plan(ctx, input)
	if err != nil {
		o := struct {
			plan.Output
			Error string
		}{output, err.Error()}
		writeJSON(o, planOutputPath)
		return err
	}
	writeJSON(output, planOutputPath)
	if isSingleRepo {
		fmt.Println(output.GitDiff)
	}
	return nil
}
