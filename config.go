package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

// Config stores the values read from the TOML config
type Config struct {
	Repository repository
	Board      board
}

type repository struct {
	Owner          string
	Name           string
	UpdateInterval int64
}

type board struct {
	Name    string
	Planned string
	Blocked string
}

// A global config variable
var config Config

func init() {
	if _, err := toml.DecodeFile("./config.toml", &config); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", config)
}
