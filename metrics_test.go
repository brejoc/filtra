package main

import (
	"github.com/brejoc/filtra/persist"
	log "github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestNewMetrics(t *testing.T) {
	// loading test config
	loadConfig("./test-data/test_config.toml")

	var results QueryPages
	if err := persist.Load("./test-data/query_pages.dump", &results); err != nil {
		log.Fatal(err)
	}
	got := NewMetrics(&results)

	testBoard := map[string]*BoardMetrics{"test": &BoardMetrics{
		closedIssueCounter:  4,
		openIssueCounter:    8,
		blockedIssueCounter: 3,
		plannedIssueCounter: 3,
		openBugsCounter:     1,
		openL3Counter:       1,
		averageLeadTime:     165.7965451388889,
		averageCycleTime:    148.83974247685185,
	}}

	want := GithubMetrics{
		closedIssueCounter: 5,
		openIssueCounter:   9,
		openBugsCounter:    1,
		openL3Counter:      2,
		Board:              testBoard,
	}

	if !reflect.DeepEqual(want, got) {
		t.Logf("Got this:    %v %v\n", got, got.Board["test"])
		t.Logf("Wanted this: %v %v\n", want, want.Board["test"])
		t.Error("Metrics are not matching.")
	}
}
