package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/lib"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/push"
	"github.com/Clever/microplane/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync workflow status with remote repo",
	Run: func(cmd *cobra.Command, args []string) {
		// find files and folders to explain the status of each repo
		initPath := outputPath("", "init")

		if _, err := os.Stat(initPath); os.IsNotExist(err) {
			log.Fatalf("must run init first: %s\n", err.Error())
		}

		var initOutput initialize.Output
		if err := loadJSON(initPath, &initOutput); err != nil {
			log.Fatalf("error loading init.json: %s\n", err.Error())
		}

		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}

		err = parallelize(repos, syncOneRepo)
		if err != nil {
			// TODO: dig into errors and display them with more detail
			log.Fatal(err)
		}
	},
}

func syncOneRepo(r lib.Repo, ctx context.Context) error {
	log.Printf("syncing: %s/%s", r.Owner, r.Name)
	repoName := r.Name

	var pushOutput struct {
		push.Output
		Error string
	}

	if !(loadJSON(outputPath(repoName, "push"), &pushOutput) == nil && pushOutput.Success) {
		return nil
	}
	output, err := syncPush(r, ctx, pushOutput.Output)
	if err != nil {
		return err
	}

	var mergeOutput struct {
		merge.Output
		Error string
	}

	if loadJSON(outputPath(repoName, "merge"), &mergeOutput) == nil && mergeOutput.Success {
		return nil
	}

	if err := syncMerge(r, ctx, output); err != nil {
		return err
	}

	log.Printf("synced: %s/%s", r.Owner, r.Name)
	return nil
}

func syncPush(r lib.Repo, ctx context.Context, pushOutput push.Output) (sync.Output, error) {

	var output sync.Output
	var err error
	if r.IsGitlab() {
		output, err = sync.GitlabSyncPush()
	} else if r.IsGithub() {
		output, err = sync.GithubSyncPush(ctx, r, pushOutput, repoLimiter)
	}
	if err != nil {
		return sync.Output{}, err
	}
	pushOutput.CommitSHA = output.CommitSHA
	pushOutput.PullRequestCombinedStatus = output.PullRequestCombinedStatus

	writeJSON(pushOutput, outputPath(r.Name, "push"))
	return output, nil
}
func syncMerge(r lib.Repo, ctx context.Context, output sync.Output) error {
	if !output.Merged {
		return nil
	}

	mergeOutputPath := outputPath(r.Name, "merge")
	mergeWorkDir := filepath.Dir(mergeOutputPath)
	if err := os.MkdirAll(mergeWorkDir, 0755); err != nil {
		return err
	}

	writeJSON(merge.Output{
		Success:        true,
		MergeCommitSHA: output.MergeCommitSHA,
	}, mergeOutputPath)
	return nil
}
