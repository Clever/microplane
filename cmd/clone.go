package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/lib"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone all repos targeted by init",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}

		err = parallelize(repos, cloneOneRepo)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func cloneOneRepo(r lib.Repo, ctx context.Context) error {
	log.Printf("cloning: %s/%s", r.Owner, r.Name)

	// Prepare workdir for current step's output
	cloneOutputPath := outputPath(r.Name, "clone")
	cloneWorkDir := filepath.Dir(cloneOutputPath)
	if err := os.MkdirAll(cloneWorkDir, 0755); err != nil {
		return err
	}

	// Execute
	input := clone.Input{
		WorkDir: cloneWorkDir,
		GitURL:  r.CloneURL,
	}
	output, err := clone.Clone(ctx, input)
	if err != nil {
		o := struct {
			clone.Output
			Error string
		}{output, err.Error()}
		writeJSON(o, cloneOutputPath)
		return err
	}
	writeJSON(output, cloneOutputPath)
	return nil
}
