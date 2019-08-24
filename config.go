package main

import (
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// Config stores the values read from the TOML config
type Config struct {
	Repository repository
	Boards     map[string]board
}

type repository struct {
	Owner          string
	Name           string
	UpdateInterval int64
}

type board struct {
	Name           string
	PlannedColumns []string
	BlockedColumns []string
	BugLabels      []string
	SupportLabels  []string
}

type boards struct {
	Boards []board
}

// A global config variable
var config Config

func init() {
	if _, err := toml.DecodeFile("./config.toml", &config); err != nil {
		log.Fatal(err)
	}
	log.Debug("%#v\n", config)
}
