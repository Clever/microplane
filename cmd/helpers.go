package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/Clever/microplane/initialize"
	"github.com/facebookgo/errgroup"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

func loadJSON(path string, obj interface{}) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, obj)
}

func writeJSON(obj interface{}, path string) error {
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

// TODO: Make parallelism configurable?
var parallelism = int64(10)

// parallelize take a list of repos and applies a function (clone, plan, ...) to them
func parallelize(repos []initialize.Repo, f func(initialize.Repo, context.Context) error) error {
	ctx := context.Background()
	var eg errgroup.Group
	parallelLimit := semaphore.NewWeighted(parallelism)
	for _, r := range repos {
		eg.Add(1)
		go func(repo initialize.Repo) {
			parallelLimit.Acquire(ctx, 1)
			defer parallelLimit.Release(1)
			defer eg.Done()

			err := f(repo, ctx)
			if err != nil {
				eg.Error(err)
				return
			}
		}(r)
	}

	return eg.Wait()
}

// whichRepos determines which repos are relevant to the current command.
// It also handles the `singleRepo` flag, allowing a user to target just one repo.
func whichRepos(cmd *cobra.Command) ([]initialize.Repo, error) {
	var initOutput initialize.Output
	if err := loadJSON(outputPath("", "init"), &initOutput); err != nil {
		return []initialize.Repo{}, err
	}

	singleRepo, err := cmd.Flags().GetString("repo")
	if err != nil {
		return []initialize.Repo{}, err
	}

	// All repos
	if singleRepo == "" {
		return initOutput.Repos, nil
	}

	// Single repo
	for _, r := range initOutput.Repos {
		if r.Name == singleRepo {
			return []initialize.Repo{r}, nil
		}
	}
	// TODO: showing valid repo names would be helpful
	return []initialize.Repo{}, fmt.Errorf("%s not a targeted repo name", singleRepo)
}

// outputPath helper constructs the output path string for each step
func outputPath(repoName string, step string) string {
	if step == "init" {
		return path.Join(workDir, "init.json")
	}
	return path.Join(workDir, repoName, step, fmt.Sprintf("%s.json", step))
}
