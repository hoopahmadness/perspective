package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
)

// Off Sunday 09/11 2022 No DST
// Or Sunday 11/20 2022 DST
var primeDateFmt = "Mon 01/02 2006 MST"
var specificDateTimeFmt = "Mon 15:04 01/02/2006 MST"
var taskDateFmt = "15:04 01/02/2006 MST"
var updateLineFmt = "15:04 01/02/2006 MST"

// function used to set the global variable PrimeSunday, which is one of two literature values used to define
// which sundays are prime (as opposed to being second sundays)
func getPrimeSunday(timezone string) time.Time {
	logger := log15.New("function", "getPrimeSunday")
	primeSunday, err := time.Parse(primeDateFmt, "Sun 09/11 2022 "+timezone)
	if err != nil {
		logger.Error("Unable to parse Prime Sunday No DST String", "err", err.Error())
	}
	return primeSunday
}

// function used to calculate the First Sunday corresponding to a given time.Time
func getZeroSunday(now time.Time) time.Time {
	zone, _ := now.Zone()
	zeroSunday := getPrimeSunday(zone)
	for {
		nextZero := zeroSunday.Add(fullTwoWeeks * time.Hour)
		if nextZero.After(now) {
			break
		}
		zeroSunday = nextZero
	}
	return zeroSunday
}

// an HourBlock is a simplified representation of an our of a day of a week.
// 0 is  midnight thru one AM on the morning of First Sunday
// 24 is 11 PM thru midnight on First Sunday
// 336 is 11 PM thru midnight on the Second Saturday...and so on
// type hourBlock int

type rotation string

const (
	firstWeek  = rotation("first")
	secondWeek = rotation("second")
	bothWeeks  = rotation("both")

	fullTwoWeeks = 336
)

// function that takes a time and spits out the number of the upcoming hour block
func nextHourBlock(now time.Time, logger log15.Logger) int {
	zone, _ := now.Zone()
	logger.Debug("The 'now' time is", "time", now.Format(specificDateTimeFmt), "rawTime.Time", now)
	nowHour := now.Add(30 * time.Minute).Round(time.Hour) // round up to the next hour
	primeDiff := nowHour.Sub(getPrimeSunday(zone))        // should be a whole number of hours
	primeDiffHours := int(primeDiff / time.Hour)

	return primeDiffHours % fullTwoWeeks
}

// function that takes a time and generates a string representation of what day and rotation
// that time corresponds to.
func whatDayIsIt(now time.Time, logger log15.Logger) string {
	hourBlock := nextHourBlock(now, logger)
	rotation := "first"
	weekday := ""
	if hourBlock >= 168 {
		rotation = "second"
		hourBlock -= 168
	}
	switch {
	case hourBlock >= 144:
		weekday = "Saturday"
	case hourBlock >= 120:
		weekday = "Friday"
	case hourBlock >= 96:
		weekday = "Thursday"
	case hourBlock >= 72:
		weekday = "Wednesday"
	case hourBlock >= 48:
		weekday = "Tuesday"
	case hourBlock >= 24:
		weekday = "Monday"
	case hourBlock >= 0:
		weekday = "Sunday"
	default:
		logger.Warn("Got a bad hourblock", "function", "whatDayIsIt")
	}

	return fmt.Sprintf("%s %s", rotation, weekday)
}

func parseDayStrings(input string) ([]time.Weekday, error) {
	dayList := []time.Weekday{}
	tokens := strings.Split(input, ",")
	for _, phrase := range tokens {
		phrase = strings.Trim(phrase, " ")
		days := strings.Split(phrase, "-")
		if len(days) == 1 { // this is a single listed day
			day, err := getWeekday(days[0])
			if err != nil {
				return dayList, fmt.Errorf("Unable to parse '%s' as single listed day", phrase)
			}
			dayList = append(dayList, day)
		} else if len(days) == 2 { // this is a range of days
			firstDay, err := getWeekday(days[0])
			if err != nil {
				return dayList, fmt.Errorf("Unable to parse first part of '%s' as range between two days", phrase)
			}
			dayList = append(dayList, firstDay)
			lastDay, err := getWeekday(days[1])
			if err != nil {
				return dayList, fmt.Errorf("Unable to parse second part of '%s' as range between two days", phrase)
			}
			if lastDay < firstDay { // temporarily "un" wrap the week
				lastDay = lastDay + 7
			}
			// iterate over loop to fill in ranges
			for middleDay := firstDay + 1; middleDay < lastDay; middleDay++ {
				if middleDay == time.Weekday(7) { // wrap
					middleDay = time.Weekday(0)
					lastDay = lastDay - 7
				}
				dayList = append(dayList, middleDay)
			}
			dayList = append(dayList, lastDay)
		} else {
			return dayList, fmt.Errorf("Unable to parse out any days from the string '%s'", phrase)
		}
	}
	return dayList, nil
}

func getWeekday(input string) (time.Weekday, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "mon", "monday":
		return time.Monday, nil
	case "tue", "tues", "tuesday":
		return time.Tuesday, nil
	case "wed", "wednesday":
		return time.Wednesday, nil
	case "thu", "thur", "thursday":
		return time.Thursday, nil
	case "fri", "friday":
		return time.Friday, nil
	case "sat", "saturday":
		return time.Saturday, nil
	case "sun", "sunday":
		return time.Sunday, nil
	default:
		return time.Weekday(-1), fmt.Errorf("Can't understand '%s' as a weekday.", input)
	}
}
func generateBlockedHours(days []time.Weekday, rotation rotation, startTime, duration int) []int {
	hourBlocks := []int{}
	for _, day := range days {
		firstWeekDayStartHour := 24 * int(day)
		secondWeekDayStartHour := firstWeekDayStartHour + (24 * 7)
		block := startTime
		for blockOffset := 0; blockOffset < duration; blockOffset++ {
			if rotation == firstWeek || rotation == bothWeeks {
				hourBlocks = append(hourBlocks, (block + blockOffset + firstWeekDayStartHour))
			}
			if rotation == secondWeek || rotation == bothWeeks {
				hourBlocks = append(hourBlocks, (block + blockOffset + secondWeekDayStartHour))
			}
		}
	}
	sort.Ints(hourBlocks)
	return hourBlocks
}

// Blocks until the 0 minute mark in a given hour; if it's currently the 0 minute then we wait a full hour
func waitForTopOfHour() {
	min := time.Now().Minute()
	wait := 60 - min
	time.Sleep(time.Duration(wait) * time.Minute)
}
