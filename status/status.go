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

var workDir string

// Input to Status
type Input struct {
	// Org on Github, e.g. "Clever"
	Org string
	// Repo is the name of the repo on Github, e.g. "microplane"
	Repo string
	// Workdir is the working dictory
	Workdir string
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

func pushStatusString(status string) string {
	s := ""
	switch status {
	case pushStatusFailedBuild:
		s += "💔"
		s += "  CI failed"
	case pushStatusRejectedReview:
		s += "🙅"
		s += "  rejected by reviewer"
	case pushStatusBuildPending:
		s += "🕑"
		s += "  CI status pending"
	case pushStatusAwaitingReview:
		s += "👀"
		s += "  awaiting review"
	case pushStatusReadyToMerge:
		s += "✅"
		s += "  ready to merge"
	default:
		s += "❓"
		s += "  unknown status: "
		s += status
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

func Status(ctx context.Context, input Input, githubLimiter *time.Ticker) (Output, error) {
	// set global, used in helper methods
	// TODO: pass more explicitly
	workDir = input.Workdir

	out := Output{
		Success:     true,
		CurrentStep: "initialized",
		Details:     "",
	}

	// has clone been run?
	var cloneOutput struct {
		clone.Output
		Error string
	}
	if !(loadJSON(outputPath(input.Repo, "clone"), &cloneOutput) == nil && cloneOutput.Success) {
		if cloneOutput.Error != "" {
			out.Details = color.RedString("(clone error) ") + cloneOutput.Error
		}
		return out, nil
	}
	out.CurrentStep = "cloned"

	// has plan been run?
	var planOutput struct {
		plan.Output
		Error string
	}
	if !(loadJSON(outputPath(input.Repo, "plan"), &planOutput) == nil && planOutput.Success) {
		if planOutput.Error != "" {
			out.Details = color.RedString("(plan error) ") + planOutput.Error
		}
		return out, nil
	}

	out.CurrentStep = "planned"
	diff, err := diffparser.Parse(planOutput.GitDiff)
	if err == nil {
		out.Details = fmt.Sprintf("%d file(s) modified", len(diff.Files))
		out.GitDiff = planOutput.GitDiff
	} else {
		out.Details = fmt.Sprintf("? file(s) modified", len(diff.Files))
		out.GitDiff = "error determining git diff"
	}

	var pushOutput struct {
		push.Output
		Error string
	}
	if !(loadJSON(outputPath(input.Repo, "push"), &pushOutput) == nil && pushOutput.Success) {
		if pushOutput.Error != "" {
			out.Details = color.RedString("(push error) ") + pushOutput.Error
		}

		// TODO: move this up into the cmd
		//// Print diff if status is planned
		//if isSingle(input.Repo) || true {
		//fmt.Println("********", input.Repo, "********")
		//fmt.Println(planOutput.GitDiff)
		//fmt.Println("")
		//}
		return out, nil
	}
	out.CurrentStep = "pushed"
	out.Details = pushOutput.String()

	var mergeOutput struct {
		merge.Output
		Error string
	}
	if !(loadJSON(outputPath(input.Repo, "merge"), &mergeOutput) == nil && mergeOutput.Success) {
		if mergeOutput.Error != "" {
			out.Details = color.RedString("(merge error) ") + mergeOutput.Error
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
				out.Details = fmt.Sprintf("error determining push status: %s", err)
			} else {
				out.Details = pushStatusString(text)
			}
		}
		return out, nil
	}
	out.CurrentStep = "merged"

	return out, nil
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
