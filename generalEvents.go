package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
)

// General Events are blocks of time on the Blocked Hours calendar. They represent days and hours of a normal two week period
// where I'm busy. Given a reference date (such as today) they can be translated into the datetime of the next occurance.
// Events can wrap around days or weeks (such as for Sleeping)
type GeneralEvent struct {
	// Friendly name for the event, i.e. Sleeping, School, Work, Church, etc
	Name string
	// first, second, or both
	Rotation rotation
	// a comma separated list of days of the week, also accepts hyphenated ranges and abbreviations
	Days string
	// StartTime is in military time e.g. "18" is the one hour block starting at 6 pm
	StartTime int
	// Duration is the number of hours that an event lasts; round up when needed.
	Duration int
	// Inactive is used to turn on and off events as needed, for example when traveling long term, without
	// having to remove the events. Inactive events are not counted towards busy hours.
	Inactive bool
	Raw      string
}

func (ge *GeneralEvent) validate() error {
	if ge.Name == "" {
		return fmt.Errorf("Event is missing a name")
	}
	if ge.Rotation == "" {
		return fmt.Errorf("Event '%s' has no deadline", ge.Name)
	}
	if ge.Days == "" {
		return fmt.Errorf("Event '%s' has no listed days", ge.Name)
	}
	if ge.Duration == 0 {
		return fmt.Errorf("Event '%s' has missing or zero duration", ge.Name)
	}
	return nil
}

func (e *GeneralEvent) AddRaw(line string) {
	e.Raw += line + "\n"
}

func (e *GeneralEvent) PrintRaw() string {
	return e.Raw
}

func (e *GeneralEvent) String() string {
	return e.PrintRaw()
}

// Method for a GeneralEvent that returns a list of hourBlocks for an arbitrary 2 week rotation
func (event *GeneralEvent) generateBlockedHours(logger log15.Logger) []int {
	days, err := parseDayStrings(event.Days)
	if err != nil {
		logger.Error("Problem parsing day strings", "days", event.Days, "err", err.Error())
	}
	return generateBlockedHours(days, event.Rotation, event.StartTime, event.Duration)
}

func getNextBlockedHours(now time.Time, genEvents []*GeneralEvent, topLogger log15.Logger) ([]int, error) {
	hourblocks := []int{}
	upcomingHour := nextHourBlock(now, topLogger)
	for _, event := range genEvents {
		logger := topLogger.New("event", event, "function", "getNextBlockedHours", "upcomingHour", upcomingHour)
		err := event.validate()
		if err != nil {
			return hourblocks, err
		}
		hours := event.generateBlockedHours(logger)
		for _, hour := range hours {
			if hour <= upcomingHour {
				hour += fullTwoWeeks
			}
			hourblocks = append(hourblocks, hour)
		}
	}
	sort.Ints([]int(hourblocks))
	return hourblocks, nil
}

func outputEvents(eventList []*GeneralEvent) string {
	outStr := ""
	active := []*GeneralEvent{}
	inactive := []*GeneralEvent{}
	for _, event := range eventList {
		if event.Inactive {
			inactive = append(inactive, event)
		} else {
			active = append(active, event)

		}
	}
	if len(active) != 0 {
		outStr += fmt.Sprintf(headerLineFmt, repeatingEvents)
		for _, task := range active {
			outStr += task.PrintRaw()
		}
	}
	if len(inactive) != 0 {
		outStr += fmt.Sprintf(headerLineFmt, inactiveEvents)
		for _, task := range inactive {
			outStr += task.PrintRaw()
		}
	}
	outStr = strings.Replace(outStr, "\n\n", "\n", -1)
	return outStr
}
