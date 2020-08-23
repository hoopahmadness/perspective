package main

import (
	// "encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type GeneralEvents struct {
	Name      string `csv:"Name"`
	Day       string `csv:"Day"`
	Hours     string `csv:"Hours"`
	datetimes []time.Time
}

func generateBlockedHours(genEvents []GeneralEvents) []time.Time {
	blockedTimes := []time.Time{}
	now := time.Now()
	todayWeekday := now.Weekday()
	//generate array of a week's worth of blocked events
	//apply the next seven days' dates to create a list of blocked hours
	//add in special events

	return blockedTimes
}

func parseTimeStrings(input string) ([]time.Duration, error) { //whole hour steps only, for now
	hoursList := []time.Duration{}
	tokens := strings.Split(input, ",")
	for _, phrase := range tokens {
		phrase = strings.Trim(phrase, " ")
		hours := strings.Split(phrase, "-")
		if len(hours) == 1 { //this is a single listed hour
			hour, err := strconv.Atoi(hours[0])
			if err != nil {
				return hoursList, err
			}
			hoursList = append(hoursList, time.Duration(hour)*time.Hour)
		} else if len(hours) == 2 { //this is a range of hours
			firstHour, err := strconv.Atoi(hours[0])
			if err != nil {
				return hoursList, err
			}
			hoursList = append(hoursList, time.Duration(firstHour)*time.Hour)
			lastHour, err := strconv.Atoi(hours[1])
			if err != nil {
				return hoursList, err
			}
			if lastHour < firstHour { //temporarily "un" wrap the day
				lastHour = lastHour + 24
			}
			//iterate over loop to fill in ranges
			for middleHour := firstHour + 1; middleHour < lastHour; middleHour++ {
				if middleHour == 24 { //wrap
					middleHour = 0
					lastHour = lastHour - 24
				}
				hoursList = append(hoursList, time.Duration(middleHour)*time.Hour)
			}
			hoursList = append(hoursList, time.Duration(lastHour)*time.Hour)

		}
	}
	return hoursList, nil
}

func parseDayStrings(input string) ([]time.Weekday, error) {
	dayList := []time.Weekday{}
	tokens := strings.Split(input, ",")
	for _, phrase := range tokens {
		phrase = strings.Trim(phrase, " ")
		days := strings.Split(phrase, "-")
		if len(days) == 1 { //this is a single listed day
			day, err := getWeekday(days[0])
			if err != nil {
				return dayList, err
			}
			dayList = append(dayList, day)
		} else if len(days) == 2 { //this is a range of days
			firstDay, err := getWeekday(days[0])
			if err != nil {
				return dayList, err
			}
			dayList = append(dayList, firstDay)
			lastDay, err := getWeekday(days[1])
			if err != nil {
				return dayList, err
			}
			if lastDay < firstDay { //temporarily "un" wrap the week
				lastDay = lastDay + 7
			}
			//iterate over loop to fill in ranges
			for middleDay := firstDay + 1; middleDay < lastDay; middleDay++ {
				if middleDay == time.Weekday(7) { //wrap
					middleDay = time.Weekday(0)
					lastDay = lastDay - 7
				}
				dayList = append(dayList, middleDay)
			}
			dayList = append(dayList, lastDay)

		}
	}
	return dayList, nil
}

func getWeekday(input string) (time.Weekday, error) {
	switch strings.ToLower(input) {
	case "mon", "monday":
		return time.Monday, nil
	case "tues", "tuesday":
		return time.Tuesday, nil
	case "wed", "wednesday":
		return time.Wednesday, nil
	case "thur", "thursday":
		return time.Thursday, nil
	case "fri", "friday":
		return time.Friday, nil
	case "sat", "saturday":
		return time.Saturday, nil
	case "sun", "sunday":
		return time.Sunday, nil
	default:
		return time.Weekday(-1), fmt.Errorf("Can't understand %s as a weekday.", input)
	}
}
