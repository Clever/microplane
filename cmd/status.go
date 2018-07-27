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
	"github.com/spf13/cobra"
)

// TODO: Move status to its own package
// Parallelize, as all else
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Status shows a workflow's progress",
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

		singleRepo, err := cmd.Flags().GetString("repo")
		if err == nil && singleRepo != "" {
			valid := false
			validRepoNames := []string{}
			for _, r := range initOutput.Repos {
				if r.Name == singleRepo {
					valid = true
					break
				}
				validRepoNames = append(validRepoNames, r.Name)
			}
			if !valid {
				log.Fatalf("%s not a targeted repo name (valid target repos are: %s)", singleRepo, strings.Join(validRepoNames, ", "))
			}
			isSingleRepo = true
		}

		repos := []string{}
		for _, r := range initOutput.Repos {
			if singleRepo != "" && r.Name != singleRepo {
				continue
			}
			repos = append(repos, r.Name)
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

func printStatus(repos []string) {
	out := tabWriterWithDefaults()
	fmt.Fprintln(out, joinWithTab("REPO", "STATUS", "DETAILS"))
	for _, r := range repos {
		status, details := getRepoStatus(r)
		d2 := strings.TrimSpace(details)
		d3 := strings.Join(strings.Split(d2, "\n"), " ")
		if len(d3) > 150 {
			d3 = d3[:150] + "..."
		}
		fmt.Fprintln(out, joinWithTab(r, status, d3))
	}
	out.Flush()
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
		Org:  r.Owner,
		Repo: r.Name,
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
