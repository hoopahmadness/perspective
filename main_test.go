package main

import (
	"testing"
	"time"

	"github.com/inconshreveable/log15"
)

func TestMDToStructs(t *testing.T) {
	tLogger := log15.New()
	tests := []struct {
		description    string
		lines          []string
		returnedEvents []*GeneralEvent
		returnedTasks  []*Task
	}{
		{
			description: "happy path",
			lines: []string{
				"- Just a list of things  ",
				"- Tasks  ",
				"	- Task1  ",
				"		- Deadline; 14:00 first Monday  ",
				"		- Estimated Hours; 10  ",
				"	- Task2  ",
				"		- Deadline; 14:00 12/25/2023 EST  ",
				"		- Estimated Hours; 10  ",
				"- Events  ",
				"	- Sleeping  ",
				"		- Rotation; both  ",
				"		- Days; Sun - Sat  ",
				"		- Start Time; 23  ",
				"		- Duration; 7  ",
				"	- Conjugate  ",
				"		- Rotation; first  ",
				"		- Days; Tue, Thur  ",
				"		- Start Time; 16  ",
				"		- Duration; 2  ",
			},
		},
	}
	for _, test := range tests {
		events, tasks := mdToStructs(test.lines, tLogger)
		for _, task := range tasks {
			err := task.calculateUrgency(time.Now(), events, tLogger)
			if err != nil {
				t.Errorf("right now none of these tests expect an error so this is a fail, but in the future we should add some error testing too")
				t.FailNow()
			}
			if task.Urgency == 0 {
				t.FailNow()
			}
		}
	}
}
