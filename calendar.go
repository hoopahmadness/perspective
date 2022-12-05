package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Off Sunday 09/11 2022 No DST
// Or Sunday 11/20 2022 DST
var primeSunday = getPrimeSunday()
var primeDateFmt = "Mon 01/02 2006 MST"
var taskDateFmt = "15:00 01/02/2006 MST"

// function used to set the global variable PrimeSunday, which is one of two literature values used to define
// which sundays are prime (as opposed to being second sundays)
func getPrimeSunday() time.Time {
	zone, _ := time.Now().Zone()
	primeSundayNoDST, err := time.Parse(primeDateFmt, "Sun 09/11 2022 "+zone)
	if err != nil {
		fmt.Println(err.Error())
	}
	NoDSTZone, _ := primeSundayNoDST.Zone()
	primeSundayDST, err := time.Parse(primeDateFmt, "Sun 11/20 2022 "+zone)
	if err != nil {
		fmt.Println(err.Error())
	}
	if NoDSTZone == zone {
		return primeSundayNoDST
	} else {
		return primeSundayDST
	}
}

// function used to calculate the First Sunday corresponding to a given time.Time
func getZeroSunday(now time.Time) time.Time {
	zeroSunday := primeSunday
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
//0 is  midnight thru one AM on the morning of First Sunday
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
func nextHourBlock(now time.Time) int {
	nowHour := now.Add(30 * time.Minute).Round(time.Hour) // round up to the next hour
	primeDiff := nowHour.Sub(primeSunday)                 // should be a whole number of hours
	primeDiffHours := int(primeDiff / time.Hour)

	return primeDiffHours % fullTwoWeeks
}

// func parseTimeStrings(input string) ([]time.Duration, error) { //whole hour steps only, for now
// 	hoursList := []time.Duration{}
// 	tokens := strings.Split(input, ",")
// 	for _, phrase := range tokens {
// 		phrase = strings.Trim(phrase, " ")
// 		hours := strings.Split(phrase, "-")
// 		if len(hours) == 1 { //this is a single listed hour
// 			hour, err := strconv.Atoi(hours[0])
// 			if err != nil {
// 				return hoursList, err
// 			}
// 			hoursList = append(hoursList, time.Duration(hour)*time.Hour)
// 		} else if len(hours) == 2 { //this is a range of hours
// 			firstHour, err := strconv.Atoi(hours[0])
// 			if err != nil {
// 				return hoursList, err
// 			}
// 			hoursList = append(hoursList, time.Duration(firstHour)*time.Hour)
// 			lastHour, err := strconv.Atoi(hours[1])
// 			if err != nil {
// 				return hoursList, err
// 			}
// 			if lastHour < firstHour { //temporarily "un" wrap the day
// 				lastHour = lastHour + 24
// 			}
// 			//iterate over loop to fill in ranges
// 			for middleHour := firstHour + 1; middleHour < lastHour; middleHour++ {
// 				if middleHour == 24 { //wrap
// 					middleHour = 0
// 					lastHour = lastHour - 24
// 				}
// 				hoursList = append(hoursList, time.Duration(middleHour)*time.Hour)
// 			}
// 			hoursList = append(hoursList, time.Duration(lastHour)*time.Hour)

// 		}
// 	}
// 	return hoursList, nil
// }

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
