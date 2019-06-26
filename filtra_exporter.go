package main

import (
	"flag"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	updateInterval = 15 // in seconds
)

var debugFlag = flag.Bool("debug", false, "Sets log level to debug.")

//Define the metrics
var allIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_all_issues", Help: "All issues"})

var openIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_open_issues", Help: "Open issues"})

var inProgress = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_in_progress", Help: "Issues that are currently in progress"})

var blockedIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_blocked_issues", Help: "Issues that are currently blocked or waiting for response"})

var closedIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_closed_issues", Help: "Closed issues"})

var openL3Issues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_open_l3_issues", Help: "Open L3 issues"})

var openBugs = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_open_bug_issues", Help: "Open bugs"})

func init() {
	//Register metrics with prometheus
	prometheus.MustRegister(allIssues)
	prometheus.MustRegister(openIssues)
	prometheus.MustRegister(inProgress)
	prometheus.MustRegister(blockedIssues)
	prometheus.MustRegister(closedIssues)
	prometheus.MustRegister(openL3Issues)
	prometheus.MustRegister(openBugs)
}

func updatePrometheusMetrics(results *Query) {
	// All issues
	allIssues.Set(float64(results.Repository.Issues.TotalCount))

	// Closed and open issues
	closedIssueCounter := 0
	openIssueCounter := 0
	openBugsCounter := 0
	openL3Counter := 0
	blockedIssueCounter := 0
	for _, issue := range results.Repository.Issues.Nodes {
		if issue.State == "CLOSED" {
			closedIssueCounter++
		} else if issue.State == "OPEN" {
			openIssueCounter++
		}
		// looking for bug or L3 labels
		for _, label := range issue.Labels.Nodes {
			labelName := strings.ToLower(string(label.Name))
			if labelName == "l3" && issue.State == "OPEN" {
				openL3Counter++
				break
			} else if labelName == "bugs" && issue.State == "OPEN" {
				openBugsCounter++
				break
			}
		}
		// looking for blocked label
		for _, column := range issue.ProjectCards.Nodes {
			columnName := strings.ToLower(string(column.Column.Name))
			if columnName == "blocked" {
				blockedIssueCounter++
				break
			}
		}
	}

	openIssues.Set(float64(openIssueCounter))
	closedIssues.Set(float64(closedIssueCounter))
	openBugs.Set(float64(openBugsCounter))
	openL3Issues.Set(float64(openL3Counter))
	blockedIssues.Set(float64(blockedIssueCounter))

	//TODO: get in progress issues

}

func main() {
	// Setting loggger to debug level when debug flag was set.
	flag.Parse()
	if *debugFlag == true {
		log.SetLevel(log.DebugLevel)
	}

	go func() {
		for {
			log.Info("Updating metrics from Github: %", time.Now())
			updatePrometheusMetrics(FetchAllIssues())
			// Sleeping for some time, so that we don't update constantly
			// and run into the request limit of Github.
			time.Sleep(time.Duration(updateInterval * time.Second))
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
