package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/initialize"
	"github.com/Clever/microplane/lib"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/plan"
	"github.com/Clever/microplane/push"
	"github.com/fatih/color"
	"github.com/nathanleiby/diffparser"
	"github.com/spf13/cobra"
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

		repos, err := whichRepos(cmd)
		if err != nil {
			log.Fatal(err)
		}
		sync, err := cmd.Flags().GetBool("sync")
		if err != nil {
			log.Fatal(err)
		}
		if sync {
			err = parallelize(repos, syncOneRepo)
			if err != nil {
				// TODO: dig into errors and display them with more detail
				log.Fatal(err)
			}
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

func printStatus(repos []lib.Repo) {
	out := tabWriterWithDefaults()
	fmt.Fprintln(out, joinWithTab("REPO", "STATUS", "DETAILS"))
	for _, r := range repos {
		status, details := getRepoStatus(r)
		d2 := strings.TrimSpace(details)
		d3 := strings.Join(strings.Split(d2, "\n"), " ")
		if len(d3) > 150 {
			d3 = d3[:150] + "..."
		}
		fmt.Fprintln(out, joinWithTab(r.Name, status, d3))
	}
	out.Flush()
}

func getRepoStatus(repo lib.Repo) (status, details string) {
	repoName := repo.Name
	status = "initialized"
	details = ""
	var cloneOutput struct {
		clone.Output
		Error string
	}
	if !(loadJSON(outputPath(repoName, "clone"), &cloneOutput) == nil && cloneOutput.Success) {
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
	if !(loadJSON(outputPath(repoName, "plan"), &planOutput) == nil && planOutput.Success) {
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
	if isSingleRepo {
		fmt.Println(planOutput.GitDiff)
	}

	var pushOutput struct {
		push.Output
		Error string
	}
	if !(loadJSON(outputPath(repoName, "push"), &pushOutput) == nil && pushOutput.Success) {
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
	// check PR was merged
	if !(loadJSON(outputPath(repoName, "merge"), &mergeOutput) == nil && mergeOutput.Success) {
		if mergeOutput.Error != "" {
			details = color.RedString("(merge error) ") + mergeOutput.Error
		}
		return
	}
	status = "merged"
	details = ""

	return
}

func init() {
	statusCmd.Flags().BoolP("sync", "s", false, "Sync workflow status with repo origin")
}
