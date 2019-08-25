package main

import (
	"strings"
	"time"

	"github.com/brejoc/githubv4"
)

// Calculates how long an issue was blocked
func calculateBlockedTime() {
	//TODO: Implement me
}

// Calculates how long an issue was worked on
func calculateWipTime() {
	//TODO: Implement me
}

// Calculates the cycle time of an issue.
func calculateCycleTime(boardName string, timelineItems queryTimelineItems, issueClosedAt githubv4.DateTime) time.Duration {
	for _, board := range config.Boards {
		if boardName == board.Name {
			for _, event := range timelineItems.Nodes {
				if event.Typename == "MovedColumnsInProjectEvent" {
					if strings.ToLower(string(event.MovedEvent.Project.Name)) == strings.ToLower(boardName) {
						previousColumn := strings.ToLower(string(event.MovedEvent.PreviousProjectColumnName))
						targetColumn := strings.ToLower(string(event.MovedEvent.ProjectColumnName))
						// We only need to calculate if the target column is not also a planned column.
						if isColumnInColumSlice(previousColumn, board.PlannedColumns) && !isColumnInColumSlice(targetColumn, board.PlannedColumns) {
							return event.MovedEvent.CreatedAt.Sub(issueClosedAt.Time)
						}
					}
				}
			}
		}
	}
	// If an issue was handled correctly, this shouldn't happen. But we have to return anything nevertheless.
	return time.Duration(0 * time.Second)
}

// Calculates the lead time of an issue.
func calculateLeadTime(createdAt githubv4.DateTime, closedAt githubv4.DateTime) time.Duration {
	return closedAt.Sub(createdAt.Time)
}

// isColumenInColumnSlice checks if a column is in a slice of columns. Cases are ignored.
func isColumnInColumSlice(column string, list []string) bool {
	for _, sliceColumn := range list {
		if strings.ToLower(sliceColumn) == strings.ToLower(column) {
			return true
		}
	}
	return false
}
