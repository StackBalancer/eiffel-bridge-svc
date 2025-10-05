package eiffel

import (
	"time"

	"github.com/google/uuid"
)

type Meta struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Time    int64  `json:"time"`
}

type SourceChangeCreatedData struct {
	Author  string `json:"author"`
	RepoURL string `json:"repoUrl"`
	Branch  string `json:"branch"`
	Commit  string `json:"commit"`
}

type Event struct {
	Meta  Meta        `json:"meta"`
	Data  interface{} `json:"data"`
	Links []any       `json:"links"`
}

func NewSourceChangeCreatedEvent(author, repo, branch, commit string) Event {
	return Event{
		Meta: Meta{
			ID:      uuid.NewString(),
			Type:    "EiffelSourceChangeCreatedEvent",
			Version: "4.0.0",
			Time:    time.Now().UnixMilli(),
		},
		Data: SourceChangeCreatedData{
			Author:  author,
			RepoURL: repo,
			Branch:  branch,
			Commit:  commit,
		},
		Links: []any{},
	}
}

type ActivityTriggeredData struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type ActivityFinishedData struct {
	Outcome string `json:"outcome"`
}

func NewActivityTriggeredEvent(name string) Event {
	return Event{
		Meta: Meta{
			ID:      uuid.NewString(),
			Type:    "EiffelActivityTriggeredEvent",
			Version: "4.0.0",
			Time:    time.Now().UnixMilli(),
		},
		Data: ActivityTriggeredData{
			Name:     name,
			Category: "CI",
		},
		Links: []any{},
	}
}

func NewActivityFinishedEvent(outcome string) Event {
	return Event{
		Meta: Meta{
			ID:      uuid.NewString(),
			Type:    "EiffelActivityFinishedEvent",
			Version: "4.0.0",
			Time:    time.Now().UnixMilli(),
		},
		Data: ActivityFinishedData{
			Outcome: outcome,
		},
		Links: []any{},
	}
}
