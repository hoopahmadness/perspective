package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	tasksFile = "To Do List.md"

	overdueTasks    = "Overdue Tasks"
	upcomingTasks   = "Upcoming Tasks"
	completedTasks  = "Completed Tasks"
	repeatingEvents = "Regular Events"
	inactiveEvents  = "Inactive Events"
	backgroundStuff = "Background Perspective Stuff"

	headerLineFmt = "\n- %s\n"
	genTextFmt    = "\n\t\t- *%s*"
)

var ourEvents []*GeneralEvent
var ourTasks []*Task
var genTextMatcher *regexp.Regexp

func main() {

	genTextMatcher, _ = regexp.Compile(`\t{2}- \*(.+)\*`)

	ourEvents, ourTasks = readFromFile()
	fmt.Println(ourEvents)
	fmt.Println(ourTasks)
	sortTasks(ourTasks, time.Now(), ourEvents)
	writeToFile(ourEvents, ourTasks)

}

// run as web server
// read/write events to md file
func readFromFile() ([]*GeneralEvent, []*Task) {
	dir := os.Getenv("NOTESDIR")
	if dir == "" {
		fmt.Println("Don't forget to set NOTESDIR")
	}
	// pull in tasks file
	readFile, err := os.Open(dir + "/" + tasksFile)
	if err != nil {
		fmt.Println(err)
	}
	defer readFile.Close()
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	lines := []string{}
	for fileScanner.Scan() {
		line := fileScanner.Text()
		lines = append(lines, line)

	}
	return mdToStructs(lines)
}

func writeToFile(events []*GeneralEvent, tasks []*Task) {
	dir := os.Getenv("NOTESDIR")
	if dir == "" {
		fmt.Println("Don't forget to set NOTESDIR")
	}
	w, err := os.Open(dir + "/" + tasksFile)
	defer w.Close()
	if err != nil {
		fmt.Println(err)
	}
	w.WriteString(outputTasks(tasks))
	w.WriteString(outputEvents(events))
}

func organizeLines(rawLine string) (string, int) {
	tokens := strings.Split(rawLine, "	")
	return strings.Trim(rawLine, "- 	"), len(tokens)
}

// read/write blocked days to json file
// change blocked hours to 2 weekly schedule instead of weekly schedule (primary, secondary week)
// when serving, calculate urgency of task or tasks and return
// every hour write out human readable list of most urgent tasks with date and time (for async devices)

func mdToStructs(rawLines []string) ([]*GeneralEvent, []*Task) {
	events := []*GeneralEvent{}
	tasks := []*Task{}
	offsets := []int{}
	lines := []string{}
	rawLines = append(rawLines, "\n")
	for _, raw := range rawLines {
		line, offset := organizeLines(raw)
		lines = append(lines, line)
		offsets = append(offsets, offset)
	}
	for ind := 0; ind < len(lines); {
		line := lines[ind]
		offset := offsets[ind]
		if line == upcomingTasks || line == overdueTasks || line == completedTasks || line == repeatingEvents || line == inactiveEvents {
			newInd := ind
			for newInd < len(lines)-1 {
				newInd++
				newOffset := offsets[newInd]
				if newOffset == offset {
					break
				}
			}
			switch line {
			case upcomingTasks, overdueTasks, completedTasks:
				tasks = append(tasks, mdToTasks(rawLines[ind+1:newInd], lines[ind+1:newInd], offsets[ind+1:newInd])...)
			case repeatingEvents, inactiveEvents:
				events = append(events, mdToEvents(rawLines[ind+1:newInd], lines[ind+1:newInd], offsets[ind+1:newInd])...)
			}
			ind = newInd
			continue
		}
		ind++
	}
	return events, tasks
}

func mdToTasks(rawLines []string, lines []string, offsets []int) []*Task {
	// offset goes up; that's the name, beginning of new Task
	// offset stays equal or goes down; that's a field
	newTask := &Task{}
	previousOffset := -1
	tasks := []*Task{}
	for index, line := range lines {
		line = strings.Trim(line, " ")
		offset := offsets[index]
		fmt.Printf("Tasks - line # %d says %s, offset is %d, previousOffset is %d \n", index, line, offset, previousOffset)
		switch {
		case previousOffset == -1 || offset < previousOffset:
			fmt.Println("Adding Name")
			newTask = &Task{
				Name: line,
			}
			tasks = append(tasks, newTask)
		case offset >= previousOffset:
			tokens := strings.Split(line, "; ")
			switch tokens[0] {
			case "Deadline":
				fmt.Printf("Adding Deadline: '%s'\n", tokens[1])
				newTask.Deadline = tokens[1]
			case "Estimated Hours":
				fmt.Printf("Adding Estimated Hours: '%s'\n", tokens[1])
				num, err := strconv.Atoi(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
				newTask.EstimatedHours = num
			default:
				fmt.Println("Whoopsie!")
				fmt.Println(tokens[0])
			}
		}
		newTask.AddRaw(rawLines[index])
		previousOffset = offset
	}
	return tasks
}

func mdToEvents(rawLines []string, lines []string, offsets []int) []*GeneralEvent {
	// offset goes up; that's the name, beginning of new Task
	// offset stays equal or goes down; that's a field
	newEvent := &GeneralEvent{}
	previousOffset := -1
	events := []*GeneralEvent{}
	for index, line := range lines {
		line = strings.Trim(line, " ")
		offset := offsets[index]
		fmt.Printf("Events - line # %d says %s, offset is %d, previousOffset is %d \n", index, line, offset, previousOffset)
		switch {
		case previousOffset == -1 || offset < previousOffset:
			fmt.Println("Adding Name")
			newEvent = &GeneralEvent{
				Name: line,
			}
			events = append(events, newEvent)
		case offset >= previousOffset:
			tokens := strings.Split(line, "; ")
			switch tokens[0] {
			case "Rotation":
				fmt.Printf("Adding Rotation: '%s'\n", tokens[1])
				newEvent.Rotation = rotation(tokens[1])
			case "Days":
				fmt.Printf("Adding Days: '%s'\n", tokens[1])
				newEvent.Days = tokens[1]
			case "Start Time":
				fmt.Printf("Adding Start Time: '%s'\n", tokens[1])
				num, err := strconv.Atoi(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
				newEvent.StartTime = num
			case "Duration":
				fmt.Printf("Adding Duration: '%s'\n", tokens[1])
				num, err := strconv.Atoi(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
				newEvent.Duration = num
			case "Inactive":
				fmt.Printf("Adding Inactivity: '%s'\n", tokens[1])
				ans, err := strconv.ParseBool(tokens[1])
				if err != nil {
					fmt.Println(err)
				}
				newEvent.Inactive = ans
			default:
				fmt.Println("Whoopsie!")
				fmt.Println(tokens[0])
			}
		}
		newEvent.AddRaw(rawLines[index])
		previousOffset = offset
	}
	return events
}
