package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/plan"
	"github.com/facebookgo/errgroup"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var planFlagBranch string
var planFlagMessage string

var planCmd = &cobra.Command{
	Use:   "plan [cmd] [args...]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Plan changes by running a command against cloned repos",
	Run: func(cmd *cobra.Command, args []string) {
		changeCmd := args[0]
		var changeCmdArgs []string
		if len(args) > 1 {
			changeCmdArgs = args[1:]
		}

		var initOutput initialize.Output
		if err := loadJSON(outputPath("", "init"), &initOutput); err != nil {
			log.Fatal(err)
		}

		branchName, err := cmd.Flags().GetString("branch")
		if err != nil {
			log.Fatal(err)
		}
		if branchName == "" {
			log.Fatal("--branch is required")
		}

		commitMessage, err := cmd.Flags().GetString("message")
		if err != nil {
			log.Fatal(err)
		}
		if commitMessage == "" {
			log.Fatal("--message is required")
		}

		singleRepo, err := cmd.Flags().GetString("repo")
		if err == nil && singleRepo != "" {
			valid := false
			for _, r := range initOutput.Repos {
				if r.Name == singleRepo {
					valid = true
					break
				}
			}
			if !valid {
				log.Fatalf("%s not a targeted repo name", singleRepo) // TODO: showing valid repo names would be helpful
			}
		}

		ctx := context.Background()
		var eg errgroup.Group
		parallelLimit := semaphore.NewWeighted(10)
		for _, r := range initOutput.Repos {
			if singleRepo != "" && r.Name != singleRepo {
				continue
			}
			var cloneOutput clone.Output
			if loadJSON(outputPath(r.Name, "clone"), &cloneOutput) != nil || !cloneOutput.Success {
				log.Printf("skipping %s, must successfully clone first", r.Name)
				continue
			}
			planOutputPath := outputPath(r.Name, "plan")
			planWorkDir := filepath.Dir(planOutputPath)
			if err := os.MkdirAll(planWorkDir, 0755); err != nil {
				log.Fatal(err)
			}

			eg.Add(1)
			go func(input plan.Input) {
				parallelLimit.Acquire(ctx, 1)
				defer parallelLimit.Release(1)
				defer eg.Done()
				log.Printf("planning: %s", input)
				output, err := plan.Plan(ctx, input)
				// TODO: should we also write the error? only saving output means "status" command only has Success: true/false to work with
				writeJSON(output, planOutputPath)
				if err != nil {
					eg.Error(err)
					return
				}
				if singleRepo != "" {
					fmt.Println(output.GitDiff)
				}
			}(plan.Input{
				RepoName:      r.Name,
				RepoDir:       cloneOutput.ClonedIntoDir,
				WorkDir:       planWorkDir,
				Command:       plan.Command{Path: changeCmd, Args: changeCmdArgs},
				CommitMessage: commitMessage,
				BranchName:    branchName,
			})
		}
		if err := eg.Wait(); err != nil {
			// TODO: dig into errors and display them with more detail
			log.Fatal(err)
		}
	},
}
