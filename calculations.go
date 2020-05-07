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
// Therefore we need to get the date of when the issue was moved to one of the "planned" columns first. The planned
// columns are defined in the config. The cycle time is the difference between this date and when the issue was closed.
func calculateCycleTime(timelineItems queryTimelineItems, createdAt githubv4.DateTime,
	closedAt githubv4.DateTime, boardName string) time.Duration {

	for _, event := range timelineItems.Nodes {
		if event.Typename == "MovedColumnsInProjectEvent" {
			if strings.ToLower(string(event.MovedEvent.Project.Name)) == strings.ToLower(boardName) {
				// previousColumn := strings.ToLower(string(event.MovedEvent.PreviousProjectColumnName))
				targetColumn := strings.ToLower(string(event.MovedEvent.ProjectColumnName))
				// Start time is when an issue is moved to a planned (backlog) column
				if isColumnInColumnSlice(targetColumn, config.Boards[boardName].PlannedColumns) &&
					event.AddedEvent.CreatedAt.Before(closedAt.Time) {
					return closedAt.Sub(event.MovedEvent.CreatedAt.Time)
				}
			}
		}
	}

	// There are cases when issues are added to boards directly in backlog or "in progress" (skipping inbox)
	// In those cases we consider the time the issue was added to the board as the initial cycle time
	for _, event := range timelineItems.Nodes {
		if strings.ToLower(string(event.AddedEvent.Project.Name)) == strings.ToLower(boardName) &&
			event.AddedEvent.CreatedAt.Before(closedAt.Time) {
			return closedAt.Sub(event.AddedEvent.CreatedAt.Time)
		}
	}

	// The issue was not handled correctly. Assume cycle time = lead time in such cases
	return calculateLeadTime(createdAt, closedAt)
}

// Calculates the lead time of an issue.
// This is the difference between when the issues was created and closed.
func calculateLeadTime(createdAt githubv4.DateTime, closedAt githubv4.DateTime) time.Duration {
	return closedAt.Sub(createdAt.Time)
}

// isColumenInColumnSlice checks if a column is in a slice of columns. Cases are ignored.
func isColumnInColumnSlice(column string, list []string) bool {
	for _, sliceColumn := range list {
		if strings.ToLower(sliceColumn) == strings.ToLower(column) {
			return true
		}
	}
	return false
}
