package push

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Command represents a command to run.
type Command struct {
	Path string
	Args []string
}

// Input to Push()
type Input struct {
	// RepoName is the name of the repo, without the owner.
	RepoName string
	// PlanDir is where the git repo that has been modified lives.
	PlanDir string
	// WorkDir is where the work associated with the Push operation happens
	WorkDir string
	// PRMessage is the text of the PR submitted to Github
	PRMessage string
	// PRAssignee is the user who will be assigned the PR
	PRAssignee string
	// RepoOwner is the name of the user who owns the Github repo
	RepoOwner string
	// BranchName is the branch name in Git
	BranchName string
}

// Output from Push()
type Output struct {
	Success                   bool
	CommitSHA                 string
	PullRequestURL            string
	PullRequestNumber         int
	PullRequestCombinedStatus string // failure, pending, or success
	PullRequestAssignee       string
	CircleCIBuildURL          string
}

func (o Output) String() string {
	s := "status:"
	switch o.PullRequestCombinedStatus {
	case "failure":
		s += "❌"
	case "pending":
		s += "🕐"
	case "success":
		s += "✅"
	default:
		s += "?"
	}

	s += fmt.Sprintf("  assignee:%s %s", o.PullRequestAssignee, o.PullRequestURL)
	if o.CircleCIBuildURL != "" {
		s += fmt.Sprintf(" %s", o.CircleCIBuildURL)
	}
	return s
}

// Push pushes the commit to Github and opens a pull request
func Push(ctx context.Context, input Input) (Output, error) {
	// Get the commit SHA from the last commit
	cmd := Command{Path: "git", Args: []string{"log", "-1", "--pretty=format:%H"}}
	gitLog := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	gitLog.Dir = input.PlanDir
	gitLogOutput, err := gitLog.CombinedOutput()
	if err != nil {
		return Output{Success: false}, errors.New(string(gitLogOutput))
	}

	// Push the commit
	gitHeadBranch := fmt.Sprintf("HEAD:%s", input.BranchName)
	cmd = Command{Path: "git", Args: []string{"push", "-f", "origin", gitHeadBranch}}
	gitPush := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	gitPush.Dir = input.PlanDir
	if output, err := gitPush.CombinedOutput(); err != nil {
		return Output{Success: false}, errors.New(string(output))
	}

	// Open a pull request, if one doesn't exist already
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	head := fmt.Sprintf("%s:%s", input.RepoOwner, input.BranchName)
	base := "master"

	// Configure a retrier - exponential backoff upon hitting Github API rate limit errors
	r := retrier.New(retrier.ExponentialBackoff(10, time.Second), retrier.WhitelistClassifier{
		&github.RateLimitError{},
		&github.AbuseRateLimitError{},
	})
	r.SetJitter(0.5)

	var pr *github.PullRequest
	err = r.Run(func() error {
		var prErr error
		pr, prErr = findOrCreatePR(ctx, client, input.RepoOwner, input.RepoName, &github.NewPullRequest{
			Title: &input.PRMessage,
			Head:  &head,
			Base:  &base,
		})
		return prErr
	})
	if err != nil {
		return Output{Success: false}, err
	}

	if pr.Assignee == nil || pr.Assignee.Login == nil || *pr.Assignee.Login != input.PRAssignee {
		err = r.Run(func() error {
			var addErr error
			_, _, addErr = client.Issues.AddAssignees(ctx, input.RepoOwner, input.RepoName,
				*pr.Number, []string{input.PRAssignee})
			return addErr
		})
		if err != nil {
			return Output{Success: false}, err
		}
	}

	var cs *github.CombinedStatus
	err = r.Run(func() error {
		var getErr error
		cs, _, getErr = client.Repositories.GetCombinedStatus(ctx, input.RepoOwner, input.RepoName, *pr.Head.SHA, nil)
		return getErr
	})
	if err != nil {
		return Output{Success: false}, err
	}

	var circleCIBuildURL string
	for _, status := range cs.Statuses {
		if status.Context != nil && *status.Context == "ci/circleci" && status.TargetURL != nil {
			circleCIBuildURL = *status.TargetURL
			// url has lots of ugly tracking query params, get rid of them
			if parsedURL, err := url.Parse(circleCIBuildURL); err == nil {
				query := parsedURL.Query()
				query.Del("utm_campaign")
				query.Del("utm_medium")
				query.Del("utm_source")
				parsedURL.RawQuery = query.Encode()
				circleCIBuildURL = parsedURL.String()
			}
		}
	}

	// TODO: if pr title != PRMessage, update it

	return Output{
		Success:                   true,
		CommitSHA:                 *pr.Head.SHA,
		PullRequestNumber:         *pr.Number,
		PullRequestURL:            *pr.HTMLURL,
		PullRequestCombinedStatus: *cs.State,
		PullRequestAssignee:       input.PRAssignee,
		CircleCIBuildURL:          circleCIBuildURL,
	}, nil
}

func findOrCreatePR(ctx context.Context, client *github.Client, owner string, name string, pull *github.NewPullRequest) (*github.PullRequest, error) {
	var pr *github.PullRequest
	newPR, _, err := client.PullRequests.Create(ctx, owner, name, pull)
	if err != nil && strings.Contains(err.Error(), "pull request already exists") {
		existingPRs, _, err := client.PullRequests.List(ctx, owner, name, &github.PullRequestListOptions{
			Head: *pull.Head,
			Base: *pull.Base,
		})
		if err != nil {
			return nil, err
		} else if len(existingPRs) != 1 {
			return nil, errors.New("unexpected: found more than 1 PR for branch")
		}
		pr = existingPRs[0]
	} else if err != nil {
		return nil, err
	} else {
		pr = newPR
	}
	return pr, nil
}
