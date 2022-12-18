package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(parseDayStrings("Tuesday, Thur-Mon"))
	sleep := &GeneralEvent{
		Name:      "Sleeping",
		Rotation:  "both",
		Days:      "Sun-Sat",
		StartTime: 23,
		Duration:  8,
	}
	fmt.Println(getNextBlockedHours(time.Now(), []*GeneralEvent{sleep}))

	// blockedEvents, tasks := readFromFile()
}

// run as web server
// read/write events to json file
// func readFromFile() ([]*GeneralEvent, []*Task) {

// }

// read/write blocked days to json file
// change blocked hours to 2 weekly schedule instead of weekly schedule (primary, secondary week)
// when serving, calculate urgency of task or tasks and return
// every hour write out human readable list of most urgent tasks with date and time (for async devices)
