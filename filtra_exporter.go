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

var debugFlag = flag.Bool("debug", false, "Sets log level to debug.")
var ownerFlag = flag.String("owner", "brejoc", "Defines owner of repository")
var repoFlag = flag.String("repo", "test", "Defines repository name")

//Define the metrics
var ghAllIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_all_issues", Help: "All issues"})

var ghOpenIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_open_issues", Help: "Open issues"})

var ghInProgress = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_in_progress", Help: "Issues that are currently in progress"})

var ghBlockedIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_blocked_issues", Help: "Issues that are currently blocked or waiting for response"})

var ghClosedIssues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_closed_issues", Help: "Closed issues"})

var ghOpenL3Issues = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_open_l3_issues", Help: "Open L3 issues"})

var ghOpenBugs = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_open_bug_issues", Help: "Open bugs"})

var ghLeadTime = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_lead_time", Help: "Average lead time of closed issues"})

var ghCycleTime = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gh_cycle_time", Help: "Average cycle time of closed issues"})

func init() {
	//Register metrics with prometheus
	prometheus.MustRegister(ghAllIssues)
	prometheus.MustRegister(ghOpenIssues)
	prometheus.MustRegister(ghInProgress)
	prometheus.MustRegister(ghBlockedIssues)
	prometheus.MustRegister(ghClosedIssues)
	prometheus.MustRegister(ghOpenL3Issues)
	prometheus.MustRegister(ghOpenBugs)
	prometheus.MustRegister(ghLeadTime)
	prometheus.MustRegister(ghCycleTime)
}

func updatePrometheusMetrics(results *QueryPages) {
	allIssuesCounter := 0
	closedIssueCounter := 0
	openIssueCounter := 0
	openBugsCounter := 0
	openL3Counter := 0
	blockedIssueCounter := 0

	for _, result := range results.Queries {
		// All issues
		allIssuesCounter += len(result.Repository.Issues.Nodes)

		// Closed and open issues
		for _, issue := range result.Repository.Issues.Nodes {
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
	}

	// Calculate average lead and cycle times
	var leadTimes []time.Duration
	var sumLeadTimes time.Duration
	var cycleTimes []time.Duration
	var sumCycleTimes time.Duration
	for _, result := range results.Queries {
		for _, issue := range result.Repository.Issues.Nodes {
			if issue.State == "CLOSED" {
				// get and append lead time of issue
				leadTime := calculateLeadTime(issue.CreatedAt, issue.ClosedAt)
				leadTimes = append(leadTimes, leadTime)

				// get and append cycle time of issue
				// TODO: Maybe it would be better to pass the whole issue here.
				cycleTime := calculateCycleTime(issue.TimelineItems, issue.CreatedAt)
				if cycleTime != time.Duration(0*time.Second) {
					cycleTimes = append(cycleTimes, cycleTime)
				}
			}
		}
	}
	// Calculate average of lead times
	for _, leadTime := range leadTimes {
		sumLeadTimes += leadTime
	}
	averageLeadTime := float64(sumLeadTimes.Hours()/24) / float64(closedIssueCounter)

	// Calculate average of cycle time
	for _, cycleTime := range cycleTimes {
		sumCycleTimes += cycleTime
	}
	averageCycleTime := float64(sumCycleTimes.Hours()/24) / float64(len(cycleTimes))

	//TODO: get in progress issues

	ghAllIssues.Set(float64(allIssuesCounter))
	ghOpenIssues.Set(float64(openIssueCounter))
	ghClosedIssues.Set(float64(closedIssueCounter))
	ghOpenBugs.Set(float64(openBugsCounter))
	ghOpenL3Issues.Set(float64(openL3Counter))
	ghBlockedIssues.Set(float64(blockedIssueCounter))
	ghLeadTime.Set(averageLeadTime)
	ghCycleTime.Set(averageCycleTime)
}

func main() {
	// Setting loggger to debug level when debug flag was set.
	flag.Parse()
	if *debugFlag == true {
		log.SetLevel(log.DebugLevel)
	}

	// Start go routine that updates values continously in the background
	go func() {
		for {
			log.Info("Updating metrics from Github: %", time.Now())
			updatePrometheusMetrics(FetchAllIssues())

			// Sleeping for some time, so that we don't update constantly
			// and run into the request limit of Github.
			log.Info("Update finished: %", time.Now())
			log.Debug("Update interval: %", config.Repository.UpdateInterval)
			time.Sleep(time.Duration(config.Repository.UpdateInterval) * time.Second)
		}
	}()

	// Start the websever
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
