package lib

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
)

type GithubStatusContext struct {
	Context   string
	TargetURL *string
}

type GithubPRStatus struct {
	// "error", "expected", "failure", "pending", or "success"
	State string

	// Results from the Status Checks. Does not include results from the Checks API (incl. Github Actions)
	Statuses []GithubStatusContext
}

func GetGithubPRStatus(ctx context.Context, repoLimiter *time.Ticker, repo Repo, prNumber int) (GithubPRStatus, error) {
	p := NewProviderFromConfig(repo.ProviderConfig)
	graphqlClient, err := p.GithubGraphqlClient(ctx)
	if err != nil {
		return GithubPRStatus{}, err
	}

	// Fetch the rolled up status checks from github
	// This includes both the Statuses API and Checks API
	var query struct {
		Repository struct {
			PullRequest struct {
				Commits struct {
					Nodes []struct {
						Commit struct {
							StatusCheckRollup struct {
								State    string
								Contexts struct {
									Nodes []struct {
										Typename      string `graphql:"__typename"`
										StatusContext struct {
											Context   string
											TargetURL string
										} `graphql:"... on StatusContext"`
									}
								} `graphql:"contexts(first: 10)"`
							}
						}
					}
				} `graphql:"commits(last: 1)"`
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	<-repoLimiter.C
	err = graphqlClient.Query(ctx, &query, map[string]interface{}{
		"owner":  graphql.String(repo.Owner),
		"name":   graphql.String(repo.Name),
		"number": graphql.Int(prNumber),
	})
	if err != nil {
		return GithubPRStatus{}, err
	}

	commits := query.Repository.PullRequest.Commits.Nodes
	if len(commits) != 1 {
		return GithubPRStatus{}, fmt.Errorf("unexpected number of commits in PR")
	}

	lastCommit := commits[0].Commit.StatusCheckRollup
	var statuses []GithubStatusContext

	for _, status := range lastCommit.Contexts.Nodes {
		if status.Typename == "StatusContext" {
			statuses = append(statuses, GithubStatusContext{
				Context:   status.StatusContext.Context,
				TargetURL: &status.StatusContext.TargetURL,
			})
		}
	}

	return GithubPRStatus{
		State:    strings.ToLower(lastCommit.State),
		Statuses: statuses,
	}, nil
}
