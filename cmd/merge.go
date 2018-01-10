package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/push"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge pushed changes",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}

		err = parallelize(repos, mergeOneRepo)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func mergeOneRepo(r initialize.Repo, ctx context.Context) error {
	log.Printf("merging: %s/%s", r.Owner, r.Name)

	// Exit early if already merged
	var mergeOutput struct {
		merge.Output
		Error string
	}
	if loadJSON(outputPath(r.Name, "merge"), &mergeOutput) == nil && mergeOutput.Success {
		log.Printf("already merged: %s/%s", r.Owner, r.Name)
		return nil
	}

	// Get previous step's output
	var pushOutput push.Output
	if loadJSON(outputPath(r.Name, "push"), &pushOutput) != nil || !pushOutput.Success {
		log.Printf("skipping %s/%s, must successfully push first", r.Owner, r.Name)
		return nil
	}
	segments := strings.Split(pushOutput.PullRequestURL, "/")
	prNumber, err := strconv.Atoi(strings.TrimSpace(segments[len(segments)-1]))
	if err != nil {
		return err
	}

	// Prepare workdir for current step's output
	mergeOutputPath := outputPath(r.Name, "merge")
	mergeWorkDir := filepath.Dir(mergeOutputPath)
	if err := os.MkdirAll(mergeWorkDir, 0755); err != nil {
		return err
	}

	// Execute
	input := merge.Input{
		Org:       r.Owner,
		Repo:      r.Name,
		PRNumber:  prNumber,
		CommitSHA: pushOutput.CommitSHA,
	}
	output, err := merge.Merge(ctx, input, githubLimiter)
	if err != nil {
		o := struct {
			merge.Output
			Error string
		}{output, err.Error()}
		writeJSON(o, mergeOutputPath)
		return err
	}
	writeJSON(output, mergeOutputPath)
	return nil
}
