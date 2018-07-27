package status

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/Clever/microplane/clone"
	"github.com/Clever/microplane/merge"
	"github.com/Clever/microplane/plan"
	"github.com/Clever/microplane/push"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/nathanleiby/diffparser"
	"golang.org/x/oauth2"
)

const pushStatusFailedBuild = "failed"
const pushStatusNotMergeable = "not-mergeable"
const pushStatusRejectedReview = "rejected-review"
const pushStatusBuildPending = "build-pending"
const pushStatusAwaitingReview = "awaiting-review"
const pushStatusReadyToMerge = "ready-to-merge"
const pushStatusAlreadyMerged = "already-merged"

// Input to Status
type Input struct {
	// Org on Github, e.g. "Clever"
	Org string
	// Repo is the name of the repo on Github, e.g. "microplane"
	Repo string
}

// Output from Status()
type Output struct {
	Success     bool
	CurrentStep string
	Details     string
	// include diff
	GitDiff string
	// allow cached version
	Timestamp time.Time
}

// TODO: Write status output to file
// Include a timestamp
// Ideally, only refresh if hasn't been refreshed in past n seconds
// Make a best effort to refresh older ones, but only do a few updates-- try to keep whole command execution < 2seconds
// Also allow a --refresh=false flag to disable polling Github (not sure which is better default behavior - refresh or no)
func pushStatusString(status string) string {
	s := ""
	switch status {
	case pushStatusFailedBuild:
		s += "💔"
		s += "  build failed"
	case pushStatusRejectedReview:
		s += "🙅"
		s += "  rejected by reviewer"
	case pushStatusBuildPending:
		s += "🕑"
		s += "  build in progress"
	case pushStatusAwaitingReview:
		s += "👀"
		s += "  awaiting review"
	case pushStatusReadyToMerge:
		s += "✅"
		s += "  ready to merge"
	default:
		s += "❓"
		s += "  PR status unknown"
	}

	return s
}

func getPushStatus(ctx context.Context, input merge.Input, githubLimiter *time.Ticker) (string, error) {
	// Create Github Client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	// OK to merge?

	// (1) Check if the PR is mergeable
	<-githubLimiter.C
	pr, _, err := client.PullRequests.Get(ctx, input.Org, input.Repo, input.PRNumber)
	if err != nil {
		return "", err
	}

	if pr.GetMerged() {
		// already merged -- we should write the merge output, if possible
		return pushStatusAlreadyMerged, nil
	}

	if !pr.GetMergeable() {
		return pushStatusNotMergeable, nil
	}

	// (2) Check commit status
	<-githubLimiter.C
	combinedStatus, _, err := client.Repositories.GetCombinedStatus(ctx, input.Org, input.Repo, input.CommitSHA, &github.ListOptions{})
	if err != nil {
		return "", err
	}

	state := combinedStatus.GetState()
	if state != "success" {
		// TODO: correct state text?
		if state == "pending" {
			return pushStatusBuildPending, nil
		}
		return pushStatusFailedBuild, nil
	}

	// (3) check if PR has been approved by a reviewer
	<-githubLimiter.C
	reviews, _, err := client.PullRequests.ListReviews(ctx, input.Org, input.Repo, input.PRNumber, &github.ListOptions{})
	if err != nil {
		return "", err
	}
	if input.RequireReviewApproval {
		if len(reviews) == 0 {
			return pushStatusAwaitingReview, nil
		}
		for _, r := range reviews {
			if r.GetState() != "APPROVED" {
				return pushStatusRejectedReview, nil
			}
		}
	}

	return pushStatusReadyToMerge, nil
}

func Status(input Input) (status, details string) {
	status = "initialized"
	details = ""
	var cloneOutput struct {
		clone.Output
		Error string
	}
	if !(loadJSON(outputPath(input.Repo, "clone"), &cloneOutput) == nil && cloneOutput.Success) {
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
	if !(loadJSON(outputPath(input.Repo, "plan"), &planOutput) == nil && planOutput.Success) {
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
	if !(loadJSON(outputPath(input.Repo, "push"), &pushOutput) == nil && pushOutput.Success) {
		if pushOutput.Error != "" {
			details = color.RedString("(push error) ") + pushOutput.Error
		}

		// TODO: move this up into the cmd
		//// Print diff if status is planned
		//if isSingle(input.Repo) || true {
		//fmt.Println("********", input.Repo, "********")
		//fmt.Println(planOutput.GitDiff)
		//fmt.Println("")
		//}

		return
	}
	status = "pushed"
	details = pushOutput.String()

	var mergeOutput struct {
		merge.Output
		Error string
	}
	if !(loadJSON(outputPath(input.Repo, "merge"), &mergeOutput) == nil && mergeOutput.Success) {
		if mergeOutput.Error != "" {
			details = color.RedString("(merge error) ") + mergeOutput.Error
		} else {
			// Lookup latest push status
			text, err := getPushStatus(context.Background(), merge.Input{
				Org:                   input.Org,
				Repo:                  input.Repo,
				PRNumber:              pushOutput.PullRequestNumber,
				CommitSHA:             pushOutput.CommitSHA,
				RequireReviewApproval: true,
			}, githubLimiter)

			if err != nil {
				fmt.Println("Error", err.Error())
			} else {
				details = merge.PushStatusString(text)
			}
		}
		return
	}
	status = "merged"
	details = ""

	return
}

// TODO
// Copied from cmd/helpers.go

func loadJSON(path string, obj interface{}) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, obj)
}

// outputPath helper constructs the output path string for each step
func outputPath(repoName string, step string) string {
	if step == "init" {
		return path.Join(workDir, "init.json")
	}
	return path.Join(workDir, repoName, step, fmt.Sprintf("%s.json", step))
}
