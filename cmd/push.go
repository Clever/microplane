package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/plan"
	"github.com/Clever/microplane/push"
	"github.com/spf13/cobra"
)

var pushFlagAssignee string

var prAssignee string

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push planned changes",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		prAssignee, err = cmd.Flags().GetString("assignee")
		if err != nil {
			log.Fatal(err)
		}
		if prAssignee == "" {
			log.Fatal("--assignee is required")
		}

		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}

		err = parallelize(repos, pushOneRepo)
		if err != nil {
			// TODO: dig into errors and display them with more detail
			log.Fatal(err)
		}

		// TODO: Fix this, doesn't play well with parallelize fn
		// query := fmt.Sprintf("org:%s \"%s\" is:open", org, commitMessage)
		// openPullRequestsURL := fmt.Sprintf("https://github.com/pulls?q=%s", url.QueryEscape(query))
		// log.Printf("Open the following URL to view your PRs: %s", openPullRequestsURL)
	},
}

func pushOneRepo(r initialize.Repo, ctx context.Context) error {
	log.Printf("pushing: %s/%s", r.Owner, r.Name)

	// Exit early if already merged
	var mergeOutput struct {
		merge.Output
		Error string
	}
	if loadJSON(outputPath(r.Name, "merge"), &mergeOutput) == nil && mergeOutput.Success {
		log.Printf("%s/%s - already merged", r.Owner, r.Name)
		return nil
	}

	// Get previous step's output
	var planOutput plan.Output
	if loadJSON(outputPath(r.Name, "plan"), &planOutput) != nil || !planOutput.Success {
		log.Printf("skipping %s/%s, must successfully plan first", r.Owner, r.Name)
		return nil
	}

	// Prepare workdir for current step's output
	pushOutputPath := outputPath(r.Name, "push")
	pushWorkDir := filepath.Dir(pushOutputPath)
	if err := os.MkdirAll(pushWorkDir, 0755); err != nil {
		return err
	}

	// Execute
	input := push.Input{
		RepoName:   r.Name,
		PlanDir:    planOutput.PlanDir,
		WorkDir:    pushWorkDir,
		PRMessage:  planOutput.CommitMessage,
		PRAssignee: prAssignee,
		BranchName: planOutput.BranchName,
		RepoOwner:  r.Owner,
	}
	output, err := push.Push(ctx, input, githubLimiter)
	if err != nil {
		o := struct {
			push.Output
			Error string
		}{output, err.Error()}
		writeJSON(o, pushOutputPath)
		return err
	}
	writeJSON(output, pushOutputPath)
	return nil
}
