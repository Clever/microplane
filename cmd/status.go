package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/status"
	"github.com/spf13/cobra"
)

// TODO: Move status to its own package
// Parallelize, as all else
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Status shows a workflow's progress",
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}

		err = parallelize(repos, statusOneRepo)
		if err != nil {
			// TODO: dig into errors and display them with more detail
			log.Fatal(err)
		}

		printStatus(repos)
	},
}

func tabWriterWithDefaults() *tabwriter.Writer {
	w := new(tabwriter.Writer)
	minWidth := 0
	tabWidth := 8
	padding := 3
	padchar := '\t'
	flags := uint(0)
	w.Init(os.Stdout, minWidth, tabWidth, padding, byte(padchar), flags)
	return w
}

func joinWithTab(s ...string) string {
	return strings.Join(s, "\t")
}

func printStatus(repos []initialize.Repo) {
	out := tabWriterWithDefaults()
	fmt.Fprintln(out, joinWithTab("REPO", "STATUS", "DETAILS"))
	for _, r := range repos {
		status, details := readRepoStatus(r.Name)
		d2 := strings.TrimSpace(details)
		d3 := strings.Join(strings.Split(d2, "\n"), " ")
		if len(d3) > 150 {
			d3 = d3[:150] + "..."
		}
		fmt.Fprintln(out, joinWithTab(r.Name, status, d3))
	}
	out.Flush()
}

func readRepoStatus(repo string) (currentStep, details string) {
	var statusOutput struct {
		status.Output
		Error string
	}
	if loadJSON(outputPath(repo, "status"), &statusOutput) == nil && statusOutput.Success {
		currentStep = statusOutput.CurrentStep
		details = statusOutput.Details
	}
	return
}

func statusOneRepo(r initialize.Repo, ctx context.Context) error {
	log.Printf("%s/%s - getting status...", r.Owner, r.Name)

	// Prepare workdir for current step's output
	statusOutputPath := outputPath(r.Name, "status")
	statusWorkDir := filepath.Dir(statusOutputPath)
	if err := os.MkdirAll(statusWorkDir, 0755); err != nil {
		return err
	}

	// Execute
	input := status.Input{
		Org:     r.Owner,
		Repo:    r.Name,
		Workdir: workDir,
	}

	output, err := status.Status(ctx, input, githubLimiter)
	if err != nil {
		log.Printf("%s/%s - status error: %s", r.Owner, r.Name, err.Error())
		o := struct {
			status.Output
			Error string
		}{output, err.Error()}
		writeJSON(o, statusOutputPath)
		return err
	}
	writeJSON(output, statusOutputPath)
	return nil
}
