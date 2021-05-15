package push

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Clever/microplane/lib"
	"github.com/xanzy/go-gitlab"
)

// GitlabPush pushes the commit to Gitlab and opens a pull request
func GitlabPush(ctx context.Context, input Input, repoLimiter *time.Ticker, pushLimiter *time.Ticker) (Output, error) {
	// Create client
	p := lib.NewProviderFromConfig(input.Repo.ProviderConfig)
	client, err := p.GitlabClient()
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
	head := input.BranchName
	base := "master"

	// Determine MR title and body
	// Title is first line of commit message.
	// Body is given by body-file if it exists or is the remainder of the commit message after title.
	title := input.CommitMessage
	body := ""
	splitMsg := strings.SplitN(input.CommitMessage, "\n", 2)
	if len(splitMsg) == 2 {
		title = splitMsg[0]
		if input.PRBody == "" {
			body = splitMsg[1]
		}
	}

	pr, err := findOrCreateGitlabMR(ctx, client, input.Repo.Owner, input.Repo.Name, &gitlab.CreateMergeRequestOptions{
		Title:        &title,
		Description:  &body,
		SourceBranch: &head,
		TargetBranch: &base,
	}, repoLimiter, pushLimiter)
	if err != nil {
		return Output{Success: false}, err
	}

	pipelineStatus, err := GetPipelineStatus(client, input.Repo.Owner, input.Repo.Name, &gitlab.ListProjectPipelinesOptions{SHA: &pr.SHA})
	if err != nil {
		return Output{Success: false}, err
	}

	buildURL := ""
	if pr.Pipeline != nil {
		buildURL = pr.Pipeline.Ref
	}

	return Output{
		Success:                   true,
		CommitSHA:                 pr.SHA,
		PullRequestNumber:         pr.IID,
		PullRequestURL:            pr.WebURL,
		PullRequestCombinedStatus: pipelineStatus,
		PullRequestAssignee:       input.PRAssignee,
		CircleCIBuildURL:          buildURL,
	}, nil
}

func findOrCreateGitlabMR(ctx context.Context, client *gitlab.Client, owner string, name string, pull *gitlab.CreateMergeRequestOptions, repoLimiter *time.Ticker, pushLimiter *time.Ticker) (*gitlab.MergeRequest, error) {
	var pr *gitlab.MergeRequest
	prStatus := "opened"
	<-pushLimiter.C
	<-repoLimiter.C
	pid := fmt.Sprintf("%s/%s", owner, name)
	newMR, _, err := client.MergeRequests.CreateMergeRequest(pid, pull)
	if err != nil && strings.Contains(err.Error(), "merge request already exists") {
		<-repoLimiter.C
		existingMRs, _, err := client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
			SourceBranch: pull.SourceBranch,
			TargetBranch: pull.TargetBranch,
			State:        &prStatus,
		})
		if err != nil {
			return nil, err
		} else if len(existingMRs) != 1 {
			return nil, errors.New("unexpected: found more than 1 MR for branch")
		}
		pr = existingMRs[0]
		//If needed, update MR title and body
		if different(&pr.Title, pull.Title) || different(&pr.Description, pull.Description) {
			pr.Title = *pull.Title
			pr.Description = *pull.Description
			<-repoLimiter.C
			pr, _, err = client.MergeRequests.UpdateMergeRequest(pid, existingMRs[0].ID, &gitlab.UpdateMergeRequestOptions{
				TargetBranch: pull.TargetBranch,
			})
			if err != nil {
				return nil, err
			}
		}

	} else if err != nil {
		return nil, err
	} else {
		pr = newMR
	}
	return pr, nil
}

// GetPipelineStatus returns status of pipeline, if pipeline is absent, returns unknown string
func GetPipelineStatus(client *gitlab.Client, owner string, name string, opts *gitlab.ListProjectPipelinesOptions) (string, error) {
	pid := fmt.Sprintf("%s/%s", owner, name)
	pipeline, _, err := client.Pipelines.ListProjectPipelines(pid, opts)
	if err != nil {
		return "", errors.New("unexpected: cannot get pipeline status")
	} else if len(pipeline) == 0 {
		return "No pipeline was found", nil
	}
	return pipeline[0].Status, nil
}
