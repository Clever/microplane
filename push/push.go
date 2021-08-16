package push

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/Clever/microplane/lib"
	"github.com/google/go-github/v35/github"
)

// Command represents a command to run.
type Command struct {
	Path string
	Args []string
}

// Input to Push()
type Input struct {
	// Repo is the git repo
	Repo lib.Repo
	// PlanDir is where the git repo that has been modified lives.
	PlanDir string
	// WorkDir is where the work associated with the Push operation happens
	WorkDir string
	// CommitMessage is the commit message for the PR
	// Its first line is used as the PR title.
	// Subsequent lines are used as the PR body if there is no body file.
	CommitMessage string
	// PRBody is the body of the PR submitted to Github
	PRBody string
	// PRAssignee is the user who will be assigned the PR
	PRAssignee string
	// BranchName is the branch name in Git
	BranchName string
	// Labels
	Labels []string
	// Draft controls whether it should be a draft PR
	Draft bool
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
func GithubPush(ctx context.Context, input Input, repoLimiter *time.Ticker, pushLimiter *time.Ticker) (Output, error) {
	// Create Github Client
	p := lib.NewProviderFromConfig(input.Repo.ProviderConfig)
	client, err := p.GithubClient(ctx)
	if err != nil {
		return Output{}, err
	}

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
	head := fmt.Sprintf("%s:%s", input.Repo.Owner, input.BranchName)
	repository, _, err := client.Repositories.Get(ctx, input.Repo.Owner, input.Repo.Name)
	if err != nil {
		return Output{Success: false}, err
	}
	base := *repository.DefaultBranch

	title, body := getTitleBody(input)
	pr, err := findOrCreatePR(ctx, client, input.Repo.Owner, input.Repo.Name, &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &head,
		Base:  &base,
		Draft: &input.Draft,
	}, repoLimiter, pushLimiter)
	if err != nil {
		return Output{Success: false}, err
	}

	if pr.Assignee == nil || pr.Assignee.Login == nil || *pr.Assignee.Login != input.PRAssignee {
		<-repoLimiter.C
		_, _, err := client.Issues.AddAssignees(ctx, input.Repo.Owner, input.Repo.Name, *pr.Number, []string{input.PRAssignee})
		if err != nil {
			return Output{Success: false}, err
		}
	}

	if pr.Labels == nil || len(input.Labels) > 0 {
		<-repoLimiter.C
		// TODO: Compare current labels
		_, _, err := client.Issues.AddLabelsToIssue(ctx, input.Repo.Owner, input.Repo.Name, *pr.Number, input.Labels)
		if err != nil {
			return Output{Success: false}, err
		}
	}

	<-repoLimiter.C
	cs, _, err := client.Repositories.GetCombinedStatus(ctx, input.Repo.Owner, input.Repo.Name, *pr.Head.SHA, nil)
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

func findOrCreatePR(ctx context.Context, client *github.Client, owner string, name string, pull *github.NewPullRequest, repoLimiter *time.Ticker, pushLimiter *time.Ticker) (*github.PullRequest, error) {
	var pr *github.PullRequest
	<-pushLimiter.C
	<-repoLimiter.C
	newPR, _, err := client.PullRequests.Create(ctx, owner, name, pull)
	if err != nil && strings.Contains(err.Error(), "pull request already exists") {
		<-repoLimiter.C
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

		// If needed, update PR title and body
		if different(pr.Title, pull.Title) || different(pr.Body, pull.Body) {
			pr.Title = pull.Title
			pr.Body = pull.Body
			<-repoLimiter.C
			pr, _, err = client.PullRequests.Edit(ctx, owner, name, *pr.Number, pr)
			if err != nil {
				return nil, err
			}
		}

	} else if err != nil {
		return nil, err
	} else {
		pr = newPR
	}
	return pr, nil
}

func different(s1, s2 *string) bool {
	return s1 != nil && s2 != nil && *s1 != *s2
}

// Determine PR title and body
// Title is first line of commit message.
// Body is the remainder of the commit message after title AND/OR `body-file` content if given
func getTitleBody(input Input) (string, string) {
	title := input.CommitMessage
	body := input.PRBody

	splitMsg := strings.SplitN(input.CommitMessage, "\n", 2)
	if len(splitMsg) == 2 {
		title = splitMsg[0]
		body = splitMsg[1] + "\n" + input.PRBody
	}

	return title, body
}
