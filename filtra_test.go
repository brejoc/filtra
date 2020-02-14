package main

import (
	"testing"

	"github.com/brejoc/filtra/persist"
	log "github.com/sirupsen/logrus"
)

func TestGetMetrics(t *testing.T) {
	var results QueryPages
	if err := persist.Load("./test-data/query_pages.dump", &results); err != nil {
		log.Fatal(err)
	}
	got := getMetrics(&results)
	want := githubMetrics{
		allIssuesCounter:    14,
		closedIssueCounter:  5,
		openIssueCounter:    9,
		openBugsCounter:     1,
		openL3Counter:       1,
		blockedIssueCounter: 3,
		plannedIssueCounter: 3,
		averageLeadTime:     312.8581111111111,
		averageCycleTime:    57.33476851851852,
	}
	if want != got {
		t.Error("Metrics are not matching.")
	}
	// TODO: Make this nice!
}
