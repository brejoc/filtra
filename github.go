package main

import (
	"context"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// QueryPages holds the multiple pages (Query) we can get from Github.
type QueryPages struct {
	Queries []Query
}

// Query is used to perform the Graphql query and also
// holds the results afterwards.
type Query struct {
	Repository struct {
		Issues struct {
			TotalCount int
			PageInfo   struct {
				StartCursor githubv4.String
				EndCursor   githubv4.String
				HasNextPage bool
			}
			Nodes []struct {
				CreatedAt     githubv4.DateTime
				ClosedAt      githubv4.DateTime
				Title         githubv4.String
				Url           githubv4.URI
				State         githubv4.StatusState
				TimelineItems struct {
					PageInfo struct {
						StartCursor githubv4.String
						EndCursor   githubv4.String
						HasNextPage bool
					}
					Nodes []struct {
						Typename   string `graphql:"__typename"`
						AddedEvent struct {
							CreatedAt githubv4.DateTime
						} `graphql:"...on AddedToProjectEvent"`
						MovedEvent struct {
							PreviousProjectColumnName githubv4.String
							ProjectColumnName         githubv4.String
							CreatedAt                 githubv4.DateTime
						} `graphql:"...on MovedColumnsInProjectEvent"`
					}
				} `graphql:"timelineItems(itemTypes: [ADDED_TO_PROJECT_EVENT, MOVED_COLUMNS_IN_PROJECT_EVENT], first: 250)"`
				ProjectCards struct {
					Nodes []struct {
						Column struct {
							Name githubv4.String
						}
					}
				} `graphql:"projectCards"`
				Labels struct {
					Nodes []struct {
						Name githubv4.String
					}
				} `graphql:"labels(first: 100)"`
			}
		} `graphql:"issues(first: 100, after: $startCursor)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// FetchAllIssues fetches all of the issues from Github and returns
// a pointer to the query struct.
func FetchAllIssues(config *Config) *QueryPages {
	queryPages := QueryPages{}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	variables := map[string]interface{}{
		"startCursor": (*githubv4.String)(nil),
		"owner":       githubv4.String(config.Owner),
		"repo":        githubv4.String(config.Repo),
	}

	pageCount := 0
	for {
		pageCount++
		log.Debug("Fetching page: ", pageCount)
		query := Query{}
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			log.Fatal(err)
		}

		log.Debug("resultCount:", query.Repository.Issues.TotalCount)
		log.Debug("      nodes:", query.Repository.Issues.Nodes)
		log.Debug(" Issue size:", len(query.Repository.Issues.Nodes))
		for _, issue := range query.Repository.Issues.Nodes {
			log.Debug("====================================")
			log.Debug("        title:", issue.Title)
			log.Debug("    createdAt:", issue.CreatedAt)
			log.Debug("          URL:", issue.Url)
			log.Debug("       Labels:")
			for _, label := range issue.Labels.Nodes {
				log.Debug("              ", label.Name)
			}

			for _, ghColumn := range issue.ProjectCards.Nodes {
				log.Debug("       Column:", ghColumn.Column.Name)
			}
		}
		queryPages.Queries = append(queryPages.Queries, query)
		if query.Repository.Issues.PageInfo.HasNextPage == true {
			variables["startCursor"] = githubv4.NewString(query.Repository.Issues.PageInfo.EndCursor)
			continue
		}
		return &queryPages
	}
}
