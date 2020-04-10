package main

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// GithubMetrics stores all of the metrics gathered from graphql.
type GithubMetrics struct {
	closedIssueCounter int
	openIssueCounter   int
	openBugsCounter    int
	openL3Counter      int
	Board              map[string]*BoardMetrics
}

// BoardMetrics stores the metrics of a particular board inside a repository.
type BoardMetrics struct {
	openIssueCounter    int
	closedIssueCounter  int
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

func (metrics GithubMetrics) writeToDB(db *sql.DB) {
	// Prepare to insert in DB
	tx, _ := db.Begin()
	// Aux func to insert map values into DB
	timeNow := time.Now()
	mapToDb := func(query string, m map[string]interface{}, extraArgs ...interface{}) {
		stmt, err := tx.Prepare(query)
		if err != nil {
			log.Fatalf("Query error: %s - %s", query, err)
		}
		defer stmt.Close()
		for k, v := range m {
			args := append([]interface{}{timeNow, k, v}, extraArgs...)
			_, err := stmt.Exec(args...)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Totals for the repo
	repoIssueMap := map[string]interface{}{
		"OPEN":        metrics.openIssueCounter,
		"CLOSED":      metrics.closedIssueCounter,
		"OPEN_BUG":    metrics.openBugsCounter,
		"OPEN_L3_BUG": metrics.openL3Counter,
	}
	mapToDb("insert into repo_counter(ts, type, value) values ($1, $2, $3)", repoIssueMap)

	// Board issue counters
	for boardName, boardMetrics := range metrics.Board {
		boardIssueMap := map[string]interface{}{
			"OPEN":        boardMetrics.openIssueCounter,
			"CLOSED":      boardMetrics.closedIssueCounter,
			"BLOCKED":     boardMetrics.blockedIssueCounter,
			"PLANNED":     boardMetrics.plannedIssueCounter,
			"OPEN_BUG":    boardMetrics.openBugsCounter,
			"OPEN_L3_BUG": boardMetrics.openL3Counter,
		}
		mapToDb("insert into board_counter(ts, type, value, board) values ($1, $2, $3, $4)", boardIssueMap, boardName)
	}

	// Board issue counters
	for boardName, boardMetrics := range metrics.Board {
		boardFlowMap := map[string]interface{}{
			"LEAD_TIME":  boardMetrics.averageLeadTime,
			"CYCLE_TIME": boardMetrics.averageCycleTime,
		}
		mapToDb("insert into board_flow(ts, type, value, board) values ($1, $2, $3, $4)", boardFlowMap, boardName)
	}
	tx.Commit()
}

// NewMetrics returns a GithubMetrics struct.
func NewMetrics(results *QueryPages) GithubMetrics {
	type boardLeadCycle struct {
		accLeadTime  time.Duration
		accCycleTime time.Duration
	}

	metrics := GithubMetrics{Board: map[string]*BoardMetrics{}}
	boardList := make([]string, len(config.Boards))
	timeCalc := map[string]*boardLeadCycle{}

	for k := range config.Boards {
		boardList = append(boardList, k)
		timeCalc[k] = &boardLeadCycle{}
		metrics.Board[k] = &BoardMetrics{}
	}

	for _, result := range results.Queries {
		for _, issue := range result.Repository.Issues.Nodes {

			isBug := false
			isL3 := false

			//  Repository Total Open and Closed issues
			if issue.State == "CLOSED" {
				metrics.closedIssueCounter++
			} else if issue.State == "OPEN" {
				metrics.openIssueCounter++

				// Check labels
				for _, label := range issue.Labels.Nodes {
					labelName := strings.ToLower(string(label.Name))

					// Is it a bug?
					for _, bugLabel := range config.Repository.BugLabels {
						if labelName == strings.ToLower(bugLabel) {
							metrics.openBugsCounter++
							isBug = true
							break
						}
					}

					// Is it a support issue?
					for _, supportLabel := range config.Repository.SupportLabels {
						if labelName == strings.ToLower(supportLabel) {
							metrics.openL3Counter++
							isL3 = true
							break
						}
					}
				}
			}

			// Iterate over project boards
			for _, column := range issue.ProjectCards.Nodes {
				boardName := string(column.Column.Project.Name)
				columnName := strings.ToLower(string(column.Column.Name))

				// Skip boards that are not part of the configured list
				if !isColumnInColumnSlice(boardName, boardList) {
					continue
				}

				// Open / Closed issues inside board
				if issue.State == "CLOSED" {
					metrics.Board[boardName].closedIssueCounter++

					// get and append lead time of issue
					leadTime := calculateLeadTime(issue.CreatedAt, issue.ClosedAt)
					timeCalc[boardName].accLeadTime += leadTime

					// get and append cycle time of issue
					cycleTime := calculateCycleTime(issue.TimelineItems, issue.CreatedAt, issue.ClosedAt, boardName)
					timeCalc[boardName].accCycleTime += cycleTime

					if log.IsLevelEnabled(log.DebugLevel) {
						fmtOut, _ := json.MarshalIndent(issue.TimelineItems, "", "  ")
						log.Debugf("Issue: %+v, Board: %s, Lead time: %v, Cycle time: %v, Created:%v, Closed: %v, timeline:%s\n",
							issue.Url, boardName, leadTime, cycleTime, issue.CreatedAt, issue.ClosedAt, fmtOut)
					}

				} else if issue.State == "OPEN" {
					metrics.Board[boardName].openIssueCounter++

					// Open Bugs and L3s inside board
					if isBug {
						metrics.Board[boardName].openBugsCounter++
					}
					if isL3 {
						metrics.Board[boardName].openL3Counter++
					}

					// Check Columns for Planned and Blocked issues
					if isColumnInColumnSlice(columnName, config.Boards[boardName].BlockedColumns) {
						metrics.Board[boardName].blockedIssueCounter++
					} else if isColumnInColumnSlice(columnName, config.Boards[boardName].PlannedColumns) {
						metrics.Board[boardName].plannedIssueCounter++
					}
				}
			}
		}
	}

	// Calculate average lead and cycle times
	for boardName, boardMetrics := range metrics.Board {
		metrics.Board[boardName].averageLeadTime =
			timeCalc[boardName].accLeadTime.Hours() / 24 / float64(boardMetrics.closedIssueCounter)
		metrics.Board[boardName].averageCycleTime =
			timeCalc[boardName].accCycleTime.Hours() / 24 / float64(boardMetrics.closedIssueCounter)
	}

	return metrics
}
