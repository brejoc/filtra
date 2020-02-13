package main

import (
	"flag"
	"strings"
	"time"

	"database/sql"
	"github.com/jasonlvhit/gocron"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

var debugFlag = flag.Bool("debug", false, "Sets log level to debug.")
var configFileFlag = flag.String("config", "./config.toml", "Path to config file")
var db *sql.DB

func updateMetrics(results *QueryPages) {
	allIssuesCounter := 0
	closedIssueCounter := 0
	openIssueCounter := 0
	openBugsCounter := 0
	openL3Counter := 0
	blockedIssueCounter := 0
	plannedIssueCounter := 0

	for _, result := range results.Queries {
		// All issues
		allIssuesCounter += len(result.Repository.Issues.Nodes)

		// Closed and open issues
		for _, issue := range result.Repository.Issues.Nodes {
			if issue.State == "CLOSED" {
				closedIssueCounter++
			} else if issue.State == "OPEN" {
				openIssueCounter++
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
									openBugsCounter++
									break
								}
							}
							// Is it a support issue?
							for _, supportLabel := range config.Board.SupportLabels {
								if labelName == strings.ToLower(supportLabel) {
									openL3Counter++
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
						blockedIssueCounter++
					} else if isColumnInColumSlice(columnName, config.Board.PlannedColumns) {
						plannedIssueCounter++
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

	// Prepare to insert in DB
	tx, _ := db.Begin()
	// Aux func to insert map values into DB
	mapToDb := func(stmt *sql.Stmt, m map[string]interface{}) {
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
		"ALL":         allIssuesCounter,
		"BLOCKED":     blockedIssueCounter,
		"CLOSED":      closedIssueCounter,
		"OPEN_ISSUE":  openIssueCounter,
		"OPEN_BUG":    openBugsCounter,
		"OPEN_L3_BUG": openL3Counter,
		"PLANNED":     plannedIssueCounter,
		// TODO: get in progress issues
	}
	stmt, err := tx.Prepare(`insert into issue_counter(ts, type, value) values (?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	mapToDb(stmt, issueMap)

	// Insert issue flows
	flowMap := map[string]interface{}{
		"LEAD_TIME":  averageLeadTime,
		"CYCLE_TIME": averageCycleTime,
	}
	stmt, err = tx.Prepare(`insert into issue_flow(ts, type, value) values (?, ?, ?)`)
	mapToDb(stmt, flowMap)
	defer stmt.Close()
	tx.Commit()
}

func updateLoop() {
	log.Infof("Updating metrics from Github: %s", time.Now())
	updateMetrics(FetchAllIssues())
	log.Infof("Update finished: %s", time.Now())
	log.Debugf("Update interval: %d", config.Repository.UpdateInterval)
}

func main() {
	// Setting logger to debug level when debug flag was set.
	flag.Parse()
	if *debugFlag == true {
		log.SetLevel(log.DebugLevel)
	}

	// globally load toml config
	if fileExists(*configFileFlag) {
		loadConfig(*configFileFlag)
	} else {
		log.Fatal("Please provide a config file with `-config <yourconfig>` or just create `config.toml` in this directory")
	}

	// Make sure update interval has a default value
	updateInterval := uint64(config.Repository.UpdateInterval)
	if updateInterval <= 0 {
		updateInterval = 1800 // 30 mins
	}

	// Initialize connection to sqlite database
	var err error
	db, err = sql.Open("sqlite3", config.Database.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Poll Github and update DB on a regular interval
	updateLoop()
	gocron.Every(updateInterval).Seconds().Do(updateLoop)
	<-gocron.Start()
}