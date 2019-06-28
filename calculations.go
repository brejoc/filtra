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
func calculateCycleTime(timelineItems queryTimelineItems, issueClosedAt githubv4.DateTime) time.Duration {
	for _, event := range timelineItems.Nodes {
		if event.Typename == "MovedColumnsInProjectEvent" {
			if strings.ToLower(string(event.MovedEvent.Project.Name)) == strings.ToLower(config.Board.Name) {
				previousColumn := strings.ToLower(string(event.MovedEvent.PreviousProjectColumnName))
				if previousColumn == strings.ToLower(config.Board.Planned) {
					return event.MovedEvent.CreatedAt.Sub(issueClosedAt.Time)
				}
			}
		}
	}
	// If an issue was handled correctly, this shouldn't happen. But we have to reaturn anything nevertheless.
	return time.Duration(0 * time.Second)
}

// Calculates the lead time of an issue.
func calculateLeadTime(createdAt githubv4.DateTime, closedAt githubv4.DateTime) time.Duration {
	return closedAt.Sub(createdAt.Time)
}
