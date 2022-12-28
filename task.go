package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	Name string
	// either a specific date with correct format or range of weekdays for repeating events
	// ex 15:00 01/02/2006 EST
	// ex 15:00 first Monday
	// ex 15:00 both Tue, Thur
	Deadline       string
	EstimatedHours int
	Urgency        float32
}

func (t *Task) getHoursLeft(now time.Time, blockedHours []int) int {
	deadlineHourBlock := 0
	interveningFortnites := 0
	nowHourBlock := nextHourBlock(now)
	fmt.Printf("NOW is hour %d\n", nowHourBlock)
	// try to parse Deadline into time
	deadline, err := time.Parse(taskDateFmt, t.Deadline)

	if err == nil { // this is a single dated task
		fmt.Println("This is a single dated task")
		for {
			earlyDeadline := deadline.Add(-1 * time.Hour * fullTwoWeeks)
			if getZeroSunday(now).After(earlyDeadline) { // We should be comparing deadline to the beginning of the two week rotation, not the day itself!
				fmt.Println("We went back too far")
				break
			}
			interveningFortnites++
			fmt.Printf("Fortnights: %d\n", interveningFortnites)
			fmt.Printf("early deadline: %v \n now: %v\n", earlyDeadline, now)
			deadline = earlyDeadline
		}
		deadlineHourBlock = nextHourBlock(deadline)
	} else { // this is a repeating task, we need to find the next instance of this deadline
		fmt.Println("This is a repeating task")
		deadlineHour, weekRotation, days := t.parseRepeatingDays()
		hours := generateBlockedHours(days, weekRotation, deadlineHour, 1)
		wrap := 0
		for index := 0; ; index++ {
			if index == len(hours) {
				index = 0
				wrap = 1
				fmt.Println("wrapping")
			}
			hour := hours[index] + (wrap * fullTwoWeeks)
			fmt.Printf("we're comparing hour %d to nowHourBlock %d\n", hour, nowHourBlock)
			if hour > nowHourBlock {
				deadlineHourBlock = hour
				break
			}
		}
	}
	// now we have a deadline hour block but it's normalized to this rotation; let's un-normalize it
	fmt.Printf("Deadline hour is %d\n", deadlineHourBlock)
	deadlineHourBlock += interveningFortnites * fullTwoWeeks

	remainingFreeHours := deadlineHourBlock - nowHourBlock
	fmt.Printf("remaining = %d - %d = %d\n", deadlineHourBlock, nowHourBlock, remainingFreeHours)

	wraps := 0
	for index := 0; ; index++ {
		if index == len(blockedHours) {
			if index == 0 {
				break
			}
			index = 0
			wraps += 1
		}
		eventHourBlock := blockedHours[index] + (wraps * fullTwoWeeks)
		if eventHourBlock < nowHourBlock {
			continue
		}
		if eventHourBlock >= deadlineHourBlock {
			break
		}
		remainingFreeHours += -1
		// fmt.Printf("Hour %d (%d) is blocked, remainingFreeHours is %d \n", blockedHours[index], eventHourBlock, remainingFreeHours)
	}
	fmt.Println("")
	return remainingFreeHours
}

func (t *Task) calculateUrgency(now time.Time, genEvents []*GeneralEvent) {
	hoursLeft := t.getHoursLeft(now, getNextBlockedHours(now, genEvents))
	t.Urgency = float32(t.EstimatedHours) / float32(hoursLeft)
}

func (t *Task) parseRepeatingDays() (hour int, weekRotation rotation, days []time.Weekday) {
	tokens := strings.Split(strings.ToLower(t.Deadline), " ")
	militaryTime := tokens[0]
	hour, err := strconv.Atoi(strings.Split(militaryTime, ":")[0])
	if err != nil {
		fmt.Println(err.Error())
	}
	weekRotation = rotation(tokens[1])
	//everything else is days
	daysStr := strings.Join(tokens[2:], " ")
	days, err = parseDayStrings(daysStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func sortTasks(tasks []*Task, now time.Time, genEvents []*GeneralEvent) {
	for _, task := range tasks {
		task.calculateUrgency(now, genEvents)
	}
	sort.Sort(byUrgency(tasks))
}

type byUrgency []*Task

func (t byUrgency) Len() int {
	return len(t)
}
func (t byUrgency) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t byUrgency) Less(i, j int) bool {
	return t[i].Urgency > t[j].Urgency
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
