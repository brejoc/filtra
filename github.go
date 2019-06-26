package main

import (
	"context"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

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
				CreatedAt    githubv4.DateTime
				ClosedAt     githubv4.DateTime
				Title        githubv4.String
				Url          githubv4.URI
				State        githubv4.StatusState
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
		} `graphql:"issues(first: 100)"`
	} `graphql:"repository(owner: \"brejoc\", name: \"test\")"`
}

// FetchAllIssues fetches all of the issues from Github and returns
// a pointer to the query struct.
func FetchAllIssues() *Query {
	query := Query{}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	err := client.Query(context.Background(), &query, nil)
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
	return &query
}
