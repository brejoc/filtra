package main

import (
	"testing"
	"time"

	"github.com/brejoc/githubv4"
)

func TestCalculateCycleTime(t *testing.T) {
	// loading test config
	loadConfig("./test-data/test_config.toml")

	currentTime := time.Now()

	node1 := &node{}
	node1.Typename = "MovedColumnsInProjectEvent"
	node1.AddedEvent = addedEvent{}
	node1.AddedEvent.Project = project{}
	node1.AddedEvent.Project.Name = "test"
	node1.AddedEvent.CreatedAt = githubv4.DateTime{currentTime.Add(time.Hour * -24)}
	node1.MovedEvent = movedEvent{}
	node1.MovedEvent.Project = project{}
	node1.MovedEvent.Project.Name = "test"
	node1.MovedEvent.PreviousProjectColumnName = "Planned"
	node1.MovedEvent.ProjectColumnName = "Done"
	node1.MovedEvent.CreatedAt = githubv4.DateTime{currentTime.Add(time.Hour * -24)}

	timelineItems := queryTimelineItems{}
	timelineItems.Nodes = []node{}
	timelineItems.Nodes = append(timelineItems.Nodes, *node1)

	want := time.Hour * -24
	got := calculateCycleTime(timelineItems, githubv4.DateTime{currentTime})
	if got != want {
		t.Errorf("Got %s for cycle time, but expected %s", got, want)
	}
}

func TestCalculateLeadTime(t *testing.T) {
	currentTime := time.Now()
	want := (24 * time.Hour)
	createdAt := githubv4.DateTime{currentTime}
	closedAt := githubv4.DateTime{currentTime.Add(want)}
	got := calculateLeadTime(createdAt, closedAt)
	if got != want {
		t.Errorf("Expected %s, but got %s for 'leadTime'", want, got)
	}
}
