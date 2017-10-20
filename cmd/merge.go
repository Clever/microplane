package cmd

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/push"
	"github.com/facebookgo/errgroup"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

func mergeOutputPath(repo string) string {
	return path.Join(workDir, repo, "merge", "merge.json")
}

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge short description",
	Long: `merge
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		var initOutput initialize.Output
		if err := loadJSON(initOutputPath(), &initOutput); err != nil {
			log.Fatal(err)
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
			var pushOutput push.Output
			if loadJSON(pushOutputPath(r.Name), &pushOutput) != nil || !pushOutput.Success {
				log.Printf("skipping %s, must successfully push first", r.Name)
				continue
			}
			segments := strings.Split(pushOutput.PullRequestURL, "/")
			prNumber, err := strconv.Atoi(strings.TrimSpace(segments[len(segments)-1]))
			if err != nil {
				log.Fatal(err)
			}

			outputPath := mergeOutputPath(r.Name)
			mergeWorkDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(mergeWorkDir, 0755); err != nil {
				log.Fatal(err)
			}

			eg.Add(1)
			go func(input merge.Input) {
				parallelLimit.Acquire(ctx, 1)
				defer parallelLimit.Release(1)
				defer eg.Done()
				log.Printf("merging: %v", input)
				output, err := merge.Merge(ctx, input)
				// TODO: should we also write the error? only saving output means "status" command only has Success: true/false to work with
				writeJSON(output, outputPath)
				if err != nil {
					eg.Error(err)
					return
				}
			}(merge.Input{
				Org:       "Clever", // TODO
				Repo:      r.Name,
				PRNumber:  prNumber,
				CommitSHA: pushOutput.CommitSHA,
			})
		}
		if err := eg.Wait(); err != nil {
			// TODO: dig into errors and display them with more detail
			log.Fatal(err)
		}

	},
}
