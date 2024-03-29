package main

import (
	"testing"
	"time"

	"github.com/inconshreveable/log15"
)

// check that tasks correctly parse a deadline string with parseRepeatingDays
func TestParseRepeatedDays(t *testing.T) {
	tLogger := log15.New()
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
		hour, week, days, err := ourTask.parseRepeatingDays(tLogger)
		if err != nil {
			t.Errorf("right now none of these tests expect an error so this is a fail, but in the future we should add some error testing too")
			t.FailNow()
		}
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
	tLogger := log15.New()
	bookClub := &Task{
		Name:           "Read For Book Club",
		Deadline:       "18:00 both Tuesday, Thursday",
		EstimatedHours: 1,
	}
	times := generateTestingTimes()
	bookReport1, bookReport2, syllabus := createThreeTasks()

	sleepEvent := &GeneralEvent{
		Name:      "sleeping",
		Rotation:  "both",
		Days:      "Sun-Sat",
		StartTime: 23,
		Duration:  8,
	}

	tests := []struct {
		title               string
		task                *Task
		remainingHoursEarly int
		remainingHoursMid   int
		remainingHoursLate  int
	}{
		{
			title:               "Repeating task uses correct instance of task to generate correct hours",
			task:                bookClub,
			remainingHoursEarly: 41,
			remainingHoursMid:   43,
			remainingHoursLate:  68,
		},
		{
			title:               "Upcoming task generates correct hours, even when in the recent past",
			task:                bookReport1,
			remainingHoursEarly: 136,
			remainingHoursMid:   26,
			remainingHoursLate:  -93,
		},
		{
			title:               "Far out task generates correct hours, even when far in the future with multiple intervening fortnights",
			task:                bookReport2,
			remainingHoursEarly: 616,
			remainingHoursMid:   506,
			remainingHoursLate:  419,
		},
		{
			title:               "Task far back in the past should keep iterating into the negatives instead of ending up positive again",
			task:                syllabus,
			remainingHoursEarly: -928,
			remainingHoursMid:   -1086,
			remainingHoursLate:  -1221,
		},
	}
	for _, test := range tests {
		// early test
		remainingEarly, err := test.task.getHoursLeft(times["early"], sleepEvent.generateBlockedHours(tLogger), tLogger)
		if err != nil {
			t.Errorf("right now none of these tests expect an error so this is a fail, but in the future we should add some error testing too")
			t.FailNow()
		}
		if remainingEarly != test.remainingHoursEarly {
			t.Errorf("Problem with '%s'; Remaining hours didn't match for early;\nExpected: %v\nActual: %v", test.title, test.remainingHoursEarly, remainingEarly)
			t.FailNow()
		}
		// mid test
		remainingMid, err := test.task.getHoursLeft(times["mid"], sleepEvent.generateBlockedHours(tLogger), tLogger)
		if err != nil {
			t.Errorf("right now none of these tests expect an error so this is a fail, but in the future we should add some error testing too")
			t.FailNow()
		}
		if remainingMid != test.remainingHoursMid {
			t.Errorf("Problem with '%s'; Remaining hours didn't match for mid;\nExpected: %v\nActual: %v", test.title, test.remainingHoursMid, remainingMid)
			t.FailNow()
		}
		// late test
		remainingLate, err := test.task.getHoursLeft(times["late"], sleepEvent.generateBlockedHours(tLogger), tLogger)
		if err != nil {
			t.Errorf("right now none of these tests expect an error so this is a fail, but in the future we should add some error testing too")
			t.FailNow()
		}
		if remainingLate != test.remainingHoursLate {
			t.Errorf("Problem with '%s'; Remaining hours didn't match for late;\nExpected: %v\nActual: %v", test.title, test.remainingHoursLate, remainingLate)
			t.FailNow()
		}
	}
}

func TestTaskUrgency(t *testing.T) {
	tLogger := log15.New()
	bookClub := &Task{
		Name:           "book club",
		Deadline:       "08:00 11/27/2022 EST",
		EstimatedHours: 24,
	}
	report1, report2, syllabus := createThreeTasks()
	taskList := []*Task{bookClub, report1, report2, syllabus}
	mid := generateTestingTimes()["mid"]
	err := sortTasks(taskList, mid, []*GeneralEvent{}, tLogger)
	if err != nil {
		t.Errorf("right now none of these tests expect an error so this is a fail, but in the future we should add some error testing too")
		t.FailNow()
	}
	if bookClub.Urgency != 2.4 {
		t.Errorf("Expected urgency of %f, got %f\n", 2.4, bookClub.Urgency)
	}
	// test sorted order
	if taskList[0] != bookClub || taskList[1] != report1 || taskList[2] != report2 || taskList[3] != syllabus {
		t.Errorf("Task list out of order for mid time:\n%v: %f\n%v: %f\n%v: %f\n%v: %f\n",
			taskList[0].Name, taskList[0].Urgency,
			taskList[1].Name, taskList[1].Urgency,
			taskList[2].Name, taskList[2].Urgency,
			taskList[3].Name, taskList[3].Urgency)
	}

	late := generateTestingTimes()["late"]
	err = sortTasks(taskList, late, []*GeneralEvent{}, tLogger)
	if err != nil {
		t.Errorf("right now none of these tests expect an error so this is a fail, but in the future we should add some error testing too")
		t.FailNow()
	}
	if bookClub.Urgency >= 0 {
		t.Errorf("Expected urgency of less than zero, got %f\n", bookClub.Urgency)
	}

	// test sorted order
	if taskList[0] != report2 || taskList[1] != syllabus || taskList[2] != report1 || taskList[3] != bookClub {
		t.Errorf("Task list out of order for late time:\n%v: %f\n%v: %f\n%v: %f\n%v: %f\n",
			taskList[0].Name, taskList[0].Urgency,
			taskList[1].Name, taskList[1].Urgency,
			taskList[2].Name, taskList[2].Urgency,
			taskList[3].Name, taskList[3].Urgency)
	}
}

func createThreeTasks() (*Task, *Task, *Task) {
	earlySyllabus := &Task{
		Name:           "Read the syllabus and get it signed",
		Deadline:       "16:00 09/28/2022 EST",
		EstimatedHours: 1,
	}
	bookReport1 := &Task{
		Name:           "Finish first book report for class",
		Deadline:       "16:00 11/28/2022 EST",
		EstimatedHours: 1,
	}
	bookReport2 := &Task{
		Name:           "Finish second book report for class",
		Deadline:       "16:00 12/28/2022 EST",
		EstimatedHours: 1,
	}
	return bookReport1, bookReport2, earlySyllabus
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
