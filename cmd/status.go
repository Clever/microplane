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
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status short description",
	Long: `status
                long
                description`,
	Run: func(cmd *cobra.Command, args []string) {
		// find files and folders to explain the status of each repo
		initPath := initOutputPath()

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
	fmt.Fprintln(out, joinWithTab("REPO", "STATUS"))
	for _, r := range repos {
		status := getRepoStatus(r)
		fmt.Fprintln(out, joinWithTab(r, status))
	}
	out.Flush()
}

func getRepoStatus(repo string) string {
	status := "initialized"
	var cloneOutput clone.Output
	if !(loadJSON(cloneOutputPath(repo), &cloneOutput) == nil && cloneOutput.Success) {
		return status
	}
	status = "cloned"

	var planOutput plan.Output
	if !(loadJSON(planOutputPath(repo), &planOutput) == nil && planOutput.Success) {
		return status
	}
	status = "planned"

	var pushOutput push.Output
	if !(loadJSON(pushOutputPath(repo), &pushOutput) == nil && pushOutput.Success) {
		return status
	}
	status = "pushed"

	var mergeOutput merge.Output
	if !(loadJSON(mergeOutputPath(repo), &mergeOutput) == nil && mergeOutput.Success) {
		return status
	}
	status = "merged"

	return status
}
