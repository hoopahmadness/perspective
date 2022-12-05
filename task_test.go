package main

import (
	"testing"
	"time"
)

// check that tasks correctly parse a deadline string with parseRepeatingDays
func TestParseRepeatedDays(t *testing.T) {
	tests := []struct {
		inputStr       string
		resultHour     int
		resultRotation rotation
		resultDays     []time.Weekday
	}{
		{
			inputStr:       "09:00 first Monday",
			resultHour:     9,
			resultRotation: firstWeek,
			resultDays:     []time.Weekday{time.Monday},
		},
		{
			inputStr:       "19:00 first Monday",
			resultHour:     19,
			resultRotation: firstWeek,
			resultDays:     []time.Weekday{time.Monday},
		},
		{
			inputStr:       "19:00 FIRST MONDAY",
			resultHour:     19,
			resultRotation: firstWeek,
			resultDays:     []time.Weekday{time.Monday},
		},
		{
			inputStr:       "09:00 second Monday",
			resultHour:     9,
			resultRotation: secondWeek,
			resultDays:     []time.Weekday{time.Monday},
		},
		{
			inputStr:       "09:00 both Monday",
			resultHour:     9,
			resultRotation: bothWeeks,
			resultDays:     []time.Weekday{time.Monday},
		},
		{
			inputStr:       "09:00 first Mon",
			resultHour:     9,
			resultRotation: firstWeek,
			resultDays:     []time.Weekday{time.Monday},
		},
		{
			inputStr:       "09:00 first Monday - Friday",
			resultHour:     9,
			resultRotation: firstWeek,
			resultDays:     []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		},
		{
			inputStr:       "09:00 first Sat-Monday",
			resultHour:     9,
			resultRotation: firstWeek,
			resultDays:     []time.Weekday{time.Saturday, time.Sunday, time.Monday},
		},
		{
			inputStr:       "09:00 second Tue, Thu",
			resultHour:     9,
			resultRotation: secondWeek,
			resultDays:     []time.Weekday{time.Tuesday, time.Thursday},
		},
	}
	for _, test := range tests {
		ourTask := &Task{
			Deadline: test.inputStr,
		}
		hour, week, days := ourTask.parseRepeatingDays()
		if hour != test.resultHour {
			t.Errorf("Got wrong hour when trying to parse deadline '%s';\nActual: %d\nExpected: %d", test.inputStr, hour, test.resultHour)
			t.FailNow()
		}
		if week != test.resultRotation {
			t.Errorf("Got wrong week when trying to parse deadline '%s';\nActual: %s\nExpected: %s", test.inputStr, week, test.resultRotation)
			t.FailNow()
		}
		if !compareWeekdayArr(days, test.resultDays) {
			t.Errorf("Got wrong weekdays when trying to parse deadline '%s';\nActual: %v\nExpected: %v", test.inputStr, days, test.resultDays)
			t.FailNow()
		}
	}
}

// check that repeating tasks correctly return remaining hours for early, mid, and late-fortnite Nows
func TestTasksReturnedHours(t *testing.T) {
	bookClub := &Task{
		Name:           "Read For Book Club",
		Deadline:       "18:00 both Tuesday, Thursday",
		EstimatedHours: 1,
	}
	times := generateTestingTimes()
	bookReport1, bookReport2 := createTwoTasks()

	sleepEvent := &GeneralEvent{
		Name:      "sleeping",
		Rotation:  "both",
		Days:      "Sun-Sat",
		StartTime: 23,
		Duration:  8,
	}

	tests := []struct {
		title string
		task *Task
		remainingHoursEarly int
		remainingHoursMid int
		remainingHoursLate int
	}{
		{
			title: "Repeating task uses correct instance of task to generate correct hours",
			task: bookClub,
			remainingHoursEarly: 41,
			remainingHoursMid: 43,
			remainingHoursLate: 68,
		},
		{
			title: "Upcoming task generates correct hours, even when in the past",
			task: bookReport1,
			remainingHoursEarly: 136,
			remainingHoursMid: 26,
			remainingHoursLate: -93,
		},
		{
			title: "Far out task generates correct hours, even when far in the future with multiple intervening fortnights",
			task: bookReport2,
			remainingHoursEarly: 616,
			remainingHoursMid: 506,
			remainingHoursLate: 419,
		},
	}
	for _, test := range tests {
		// early test
		remainingEarly := test.task.getHoursLeft(times["early"], sleepEvent.generateBlockedHours())
		if remainingEarly != test.remainingHoursEarly {
			t.Errorf("Problem with '%s'; Remaining hours didn't match for early;\nExpected: %v\nActual: %v", test.title, test.remainingHoursEarly, remainingEarly)
			t.FailNow()
		}
		// mid test
		remainingMid := test.task.getHoursLeft(times["mid"], sleepEvent.generateBlockedHours())
		if remainingMid != test.remainingHoursMid {
			t.Errorf("Problem with '%s'; Remaining hours didn't match for mid;\nExpected: %v\nActual: %v", test.title, test.remainingHoursMid, remainingMid)
			t.FailNow()
		}
		// late test
		remainingLate := test.task.getHoursLeft(times["late"], sleepEvent.generateBlockedHours())
		if remainingLate != test.remainingHoursLate {
			t.Errorf("Problem with '%s'; Remaining hours didn't match for late;\nExpected: %v\nActual: %v", test.title, test.remainingHoursLate, remainingLate)
			t.FailNow()
		}
	}
}

func TestTaskUrgency(t *testing.T) {
	bookClub := &Task{
		Name: "book club",
		Deadline: "08:00 11/27/2022 EST",
		EstimatedHours: 24,
	}
	mid := generateTestingTimes()["mid"]
	urgency := bookClub.getUrgency(mid, []*GeneralEvent{})
	if urgency != 2.4 {
		t.Errorf("Expected urgency of %f, got %f\n", 2.4, urgency)
	}

	late := generateTestingTimes()["late"]
	urgency = bookClub.getUrgency(late, []*GeneralEvent{})
	if urgency >=0 {
		t.Errorf("Expected urgency of less than zero, got %f\n", urgency)
	}


}

func createTwoTasks() (*Task, *Task) {
		bookReport1 := &Task{
		Name: "Finish first book report for class",
		Deadline: "16:00 11/28/2022 EST",
		EstimatedHours: 1,
	}
	bookReport2 := &Task{
		Name: "Finish second book report for class",
		Deadline: "16:00 12/28/2022 EST",
		EstimatedHours: 1,
	}
	return bookReport1, bookReport2
}

func compareWeekdayArr(arr1, arr2 []time.Weekday) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for index, element := range arr1 {
		if element != arr2[index] {
			return false
		}
	}
	return true
}
