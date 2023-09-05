package main

import "time"

// import ("testing")

func generateTestingTimes() map[string]time.Time {
	early, _ := time.Parse(taskDateFmt, "08:00 11/20/2022 "+"EST") // first Sun morning
	mid, _ := time.Parse(taskDateFmt, "22:00 11/26/2022 "+"EST")   // first sat night
	late, _ := time.Parse(taskDateFmt, "13:00 12/02/2022 "+"EST")  // second Fri afternoon
	return map[string]time.Time{
		"early": early,
		"mid":   mid,
		"late":  late,
	}
}

// check that nextHourBlock correctly rounds Now and generates the correct hourBlock value

// check parseDayStrings
