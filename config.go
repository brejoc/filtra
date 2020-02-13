package main

import (
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// Config stores the values read from the TOML config
type Config struct {
	Repository repository
	Board      board
	Database   database
}

type repository struct {
	Owner          string
	Name           string
	UpdateInterval uint64
}

type board struct {
	Name           string
	PlannedColumns []string
	BlockedColumns []string
	BugLabels      []string
	SupportLabels  []string
}

type database struct {
	URI string
}

// A global config variable
var config Config

func loadConfig(pathToConfig string) {
	if _, err := toml.DecodeFile(pathToConfig, &config); err != nil {
		log.Fatal(err)
	}
	log.Debugf("%#v\n", config)
}
