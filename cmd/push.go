package cmd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/plan"
	"github.com/Clever/microplane/push"
	"github.com/facebookgo/errgroup"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

func pushOutputPath(repo string) string {
	return path.Join(workDir, repo, "push", "push.json")
}

var pushFlagAssignee string

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push planned changes",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var initOutput initialize.Output
		if err := loadJSON(initOutputPath(), &initOutput); err != nil {
			log.Fatal(err)
		}

		prAssignee, err := cmd.Flags().GetString("assignee")
		if err != nil {
			log.Fatal(err)
		}
		if prAssignee == "" {
			log.Fatal("--assignee is required")
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
		var commitMessage string
		var org string
		for _, r := range initOutput.Repos {
			if singleRepo != "" && r.Name != singleRepo {
				continue
			}
			var planOutput plan.Output
			if loadJSON(planOutputPath(r.Name), &planOutput) != nil || !planOutput.Success {
				log.Printf("skipping %s, must successfully plan first", r.Name)
				continue
			}
			commitMessage = planOutput.CommitMessage
			org = r.Owner
			outputPath := pushOutputPath(r.Name)
			pushWorkDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(pushWorkDir, 0755); err != nil {
				log.Fatal(err)
			}

			eg.Add(1)
			go func(input push.Input) {
				parallelLimit.Acquire(ctx, 1)
				defer parallelLimit.Release(1)
				defer eg.Done()
				log.Printf("pushing: %s", input)
				output, err := push.Push(ctx, input)
				// TODO: should we also write the error? only saving output means "status" command only has Success: true/false to work with
				writeJSON(output, outputPath)
				if err != nil {
					eg.Error(err)
					return
				}
			}(push.Input{
				RepoName:   r.Name,
				PlanDir:    planOutput.PlanDir,
				WorkDir:    pushWorkDir,
				PRMessage:  planOutput.CommitMessage,
				PRAssignee: prAssignee,
				BranchName: planOutput.BranchName,
				RepoOwner:  r.Owner,
			})
		}
		if err := eg.Wait(); err != nil {
			// TODO: dig into errors and display them with more detail
			log.Fatal(err)
		}
		query := fmt.Sprintf("org:%s \"%s\" is:open", org, commitMessage)
		openPullRequestsURL := fmt.Sprintf("https://github.com/pulls?q=%s", url.QueryEscape(query))
		log.Printf("Open the following URL to view your PRs: %s", openPullRequestsURL)
	},
}
