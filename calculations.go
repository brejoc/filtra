package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
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
func calculateCycleTime(timelineItems queryTimelineItems) time.Duration {
	//TODO: Implement me
	for _, event := range timelineItems.Nodes {
		log.Info("==============")
		if event.Typename == "MovedColumnsInProjectEvent" {
			log.Info("name    : ", event.Typename)
			log.Info("node    : ", event.MovedEvent.CreatedAt)
			log.Info("from    : ", event.MovedEvent.PreviousProjectColumnName)
			log.Info("to      : ", event.MovedEvent.ProjectColumnName)
			log.Info("project : ", event.MovedEvent.Project.Name)
		}
	}
	return time.Now().Sub(time.Now())
}

// Calculates the lead time of an issue.
func calculateLeadTime(createdAt githubv4.DateTime, closedAt githubv4.DateTime) time.Duration {
	return closedAt.Sub(createdAt.Time)
}
