package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/lib"
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

func parallelize(repos []lib.Repo, f func(lib.Repo, context.Context) error) error {
	return parallelizeLimited(repos, f, 3)
}

// parallelize take a list of repos and applies a function (clone, plan, ...) to them
func parallelizeLimited(repos []lib.Repo, f func(lib.Repo, context.Context) error, parallelismLimit int64) error {
	ctx := context.Background()
	var eg errgroup.Group
	parallelLimit := semaphore.NewWeighted(parallelismLimit)
	for _, r := range repos {
		eg.Add(1)
		go func(repo lib.Repo) {
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
func whichRepos(cmd *cobra.Command) ([]lib.Repo, error) {
	var initOutput initialize.Output
	if err := loadJSON(outputPath("", "init"), &initOutput); err != nil {
		return []lib.Repo{}, err
	}

	singleRepo, err := cmd.Flags().GetString("repo")
	if err != nil {
		return []lib.Repo{}, err
	}

	// All repos
	if singleRepo == "" {
		return initOutput.Repos, nil
	}

	// Single repo
	for _, r := range initOutput.Repos {
		if r.Name == singleRepo {
			return []lib.Repo{r}, nil
		}
	}
	// TODO: showing valid repo names would be helpful
	return []lib.Repo{}, fmt.Errorf("%s not a targeted repo name", singleRepo)
}
