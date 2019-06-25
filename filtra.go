// filtra …
//
// Usage: filtra

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	var query struct {
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

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("    foo:", query.Repository.Issues.TotalCount)
	fmt.Println("    foo:", query.Repository.Issues.PageInfo.StartCursor)
	fmt.Println("  nodes:", query.Repository.Issues.Nodes)
	fmt.Println("   size:", len(query.Repository.Issues.Nodes))
	for _, issue := range query.Repository.Issues.Nodes {
		fmt.Println("====================================")
		fmt.Println("    title:", issue.Title)
		fmt.Println("createdAt:", issue.CreatedAt)
		fmt.Println("      URL:", issue.Url)
		fmt.Println("   Labels:")
		for _, label := range issue.Labels.Nodes {
			fmt.Println("          ", label.Name)
		}

		for _, ghColumn := range issue.ProjectCards.Nodes {
			fmt.Println("   Column:", ghColumn.Column.Name)
		}
	}
}
