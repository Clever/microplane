package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Clever/microplane/clone"
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

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone all repos targeted by init",
	Run: func(cmd *cobra.Command, args []string) {
		var initOutput initialize.Output
		if err := loadJSON(outputPath("", "init"), &initOutput); err != nil {
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
			cloneOutputPath := outputPath(r.Name, "clone")
			cloneWorkDir := filepath.Dir(cloneOutputPath)
			if err := os.MkdirAll(cloneWorkDir, 0755); err != nil {
				log.Fatal(err)
			}

			eg.Add(1)
			go func(cloneInput clone.Input) {
				parallelLimit.Acquire(ctx, 1)
				defer parallelLimit.Release(1)
				defer eg.Done()
				log.Printf("cloning: %s", cloneInput.GitURL)
				output, err := clone.Clone(ctx, cloneInput)
				// TODO: should we also write the error? only saving output means "status" command only has Success: true/false to work with
				writeJSON(output, cloneOutputPath)
				if err != nil {
					eg.Error(err)
					return
				}
			}(clone.Input{
				WorkDir: cloneWorkDir,
				GitURL:  r.CloneURL,
			})
		}
		if err := eg.Wait(); err != nil {
			log.Fatal(err)
		}
	},
}
