package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"database/sql"

	"github.com/jasonlvhit/gocron"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
)

var db *sql.DB

func updateLoop() {
	log.Infof("Updating metrics from Github: %s", time.Now())
	issues, err := FetchAllIssues()
	if err != nil {
		log.Error("Not able to fetch issues from Github: ", err)
	} else {
		metrics := NewMetrics(issues)
		metrics.writeToDB(db)
		log.Infof("Update finished: %s", time.Now())
		log.Debugf("Update interval: %d", config.Repository.UpdateInterval)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		debugFlag      = flags.Bool("debug", false, "Sets log level to debug.")
		configFileFlag = flags.String("config", "./config.toml", "Path to config file")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	// Setting logger to debug level when debug flag was set.
	if *debugFlag == true {
		log.SetLevel(log.DebugLevel)
	}

	// globally load toml config
	if fileExists(*configFileFlag) {
		loadConfig(*configFileFlag)
	} else {
		log.Fatal("Please provide a config file with `-config <yourconfig>` or just create `config.toml` in this directory")
	}

	// Make sure update interval has a default value
	updateInterval := uint64(config.Repository.UpdateInterval)
	if updateInterval <= 0 {
		updateInterval = 1800 // 30 mins
	}

	// Initialize connection to sqlite database
	var err error
	var psqlConfig = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBname)

	// Connect to PostgreSQL
	db, _ = sql.Open("postgres", psqlConfig)
	defer db.Close()
	// Test if our connection actually works
	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %s", err)
	}

	// Poll Github and update DB on a regular interval
	updateLoop()
	gocron.Every(updateInterval).Seconds().Do(updateLoop)
	<-gocron.Start()
	return nil
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
