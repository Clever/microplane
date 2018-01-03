package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/plan"
	"github.com/Clever/microplane/push"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/waigani/diffparser"
)

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

func getRepoStatus(repo string) (status, details string) {
	status = "initialized"
	details = ""
	var cloneOutput struct {
		clone.Output
		Error string
	}
	if !(loadJSON(outputPath(repo, "clone"), &cloneOutput) == nil && cloneOutput.Success) {
		if cloneOutput.Error != "" {
			details = color.RedString("(clone error) ") + cloneOutput.Error
		}
		return
	}
	status = "cloned"

	var planOutput struct {
		plan.Output
		Error string
	}
	if !(loadJSON(outputPath(repo, "plan"), &planOutput) == nil && planOutput.Success) {
		if planOutput.Error != "" {
			details = color.RedString("(plan error) ") + planOutput.Error
		}
		return
	}
	status = "planned"
	diff, err := diffparser.Parse(planOutput.GitDiff)
	if err == nil {
		details = fmt.Sprintf("%d file(s) modified", len(diff.Files))
	}

	var pushOutput struct {
		push.Output
		Error string
	}
	if !(loadJSON(outputPath(repo, "push"), &pushOutput) == nil && pushOutput.Success) {
		if pushOutput.Error != "" {
			details = color.RedString("(push error) ") + pushOutput.Error
		}
		return
	}
	status = "pushed"
	details = pushOutput.String()

	var mergeOutput struct {
		merge.Output
		Error string
	}
	if !(loadJSON(outputPath(repo, "merge"), &mergeOutput) == nil && mergeOutput.Success) {
		if mergeOutput.Error != "" {
			details = color.RedString("(merge error) ") + mergeOutput.Error
		}
		return
	}
	status = "merged"

	return
}
