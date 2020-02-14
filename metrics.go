package main

import (
	"strings"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type GithubMetrics struct {
	allIssuesCounter    int
	closedIssueCounter  int
	openIssueCounter    int
	openBugsCounter     int
	openL3Counter       int
	blockedIssueCounter int
	plannedIssueCounter int
	averageLeadTime     float64
	averageCycleTime    float64
}

type dbWriter interface {
	writeToDb()
}

func (metrics GithubMetrics) writeToDB() {
	// Prepare to insert in DB
	tx, _ := db.Begin()
	// Aux func to insert map values into DB
	mapToDb := func(query string, m map[string]interface{}) {
		stmt, err := tx.Prepare(query)
		if err != nil {
			log.Fatalf("Query error: %s - %s", query, err)
		}
		defer stmt.Close()
		timeNow := time.Now()
		for k, v := range m {
			_, err := stmt.Exec(timeNow, k, v)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Insert issue counters
	issueMap := map[string]interface{}{
		"ALL":         metrics.allIssuesCounter,
		"BLOCKED":     metrics.blockedIssueCounter,
		"CLOSED":      metrics.closedIssueCounter,
		"OPEN_ISSUE":  metrics.openIssueCounter,
		"OPEN_BUG":    metrics.openBugsCounter,
		"OPEN_L3_BUG": metrics.openL3Counter,
		"PLANNED":     metrics.plannedIssueCounter,
		// TODO: get in progress issues
	}
	mapToDb("insert into issue_counter(ts, type, value) values ($1, $2, $3)", issueMap)

	// Insert issue flows
	flowMap := map[string]interface{}{
		"LEAD_TIME":  metrics.averageLeadTime,
		"CYCLE_TIME": metrics.averageCycleTime,
	}
	mapToDb("insert into issue_flow(ts, type, value) values ($1, $2, $3)", flowMap)
	tx.Commit()
}

// NewMetrics returns a githubMetrics struct.
func NewMetrics(results *QueryPages) GithubMetrics {
	metrics := GithubMetrics{}

	for _, result := range results.Queries {
		// All issues
		metrics.allIssuesCounter += len(result.Repository.Issues.Nodes)

		// Closed and open issues
		for _, issue := range result.Repository.Issues.Nodes {
			if issue.State == "CLOSED" {
				metrics.closedIssueCounter++
			} else if issue.State == "OPEN" {
				metrics.openIssueCounter++
				for _, label := range issue.Labels.Nodes {
					labelName := strings.ToLower(string(label.Name))
					// Issues can only be counted, when they are on the right board. we have to
					// check this by iterating over the columns.
					for _, column := range issue.ProjectCards.Nodes {
						boardName := strings.ToLower(string(column.Column.Project.Name))
						if boardName == strings.ToLower(config.Board.Name) {
							// Is it a bug?
							for _, bugLabel := range config.Board.BugLabels {
								if labelName == strings.ToLower(bugLabel) {
									metrics.openBugsCounter++
									break
								}
							}
							// Is it a support issue?
							for _, supportLabel := range config.Board.SupportLabels {
								if labelName == strings.ToLower(supportLabel) {
									metrics.openL3Counter++
									break
								}
							}
							break // The issue can't be two times on the board, so we can break here.
						}
					}

				}
			}
			// There might be a closed issue in the columns… so we are doing this
			// for all of the issues and not only for the open ones.
			//
			// Looking for issues in blocked and planned columns.
			for _, column := range issue.ProjectCards.Nodes {
				boardName := strings.ToLower(string(column.Column.Project.Name))
				columnName := strings.ToLower(string(column.Column.Name))
				if boardName == strings.ToLower(config.Board.Name) {
					if isColumnInColumSlice(columnName, config.Board.BlockedColumns) {
						metrics.blockedIssueCounter++
					} else if isColumnInColumSlice(columnName, config.Board.PlannedColumns) {
						metrics.plannedIssueCounter++
					}
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
					log.Debug(cycleTime)
				}
			}
		}
	}
	// Calculate average of lead times
	for _, leadTime := range leadTimes {
		sumLeadTimes += leadTime
	}
	metrics.averageLeadTime = float64(sumLeadTimes.Hours()/24) / float64(metrics.closedIssueCounter)

	// Calculate average of cycle time
	for _, cycleTime := range cycleTimes {
		sumCycleTimes += cycleTime
	}
	metrics.averageCycleTime = float64(sumCycleTimes.Hours()/24) / float64(len(cycleTimes))

	return metrics
}
