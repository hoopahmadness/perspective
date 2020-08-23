package main

import (
	"encoding/json"
	"time"
)

type Task struct {
	Name           string
	Deadline       time.Time
	EstimatedHours time.Duration
	Urgency        float32
	HoursLeft      time.Duration
	BusyHours      int
}

/*
config file will have file names for the following:
a list of events corresponding to a normal calendar week,
a list of special events with specific dates as needed, and
a list of tasks

program will run with a flag to config file and flag to output file

if no output, program will just print to console

program takes in task list, calculates urgency and orders them in output
program also prints warnings for expired events or tasks but does not remove them
maybe add a flag that will automatically remove those, but default is false.

commands:
	run --config --output
	update --config [editor]
		regular
		special
		tasks <<these will just open the files in an editor
*/
