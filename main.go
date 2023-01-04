package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	tasksFile      = "Task List.md"
	sortedListFile = "Top Priorities.md"
)

var ourEvents []*GeneralEvent
var ourTasks []*Task

func main() {

	ourEvents, ourTasks = readFromFile()
	fmt.Println(ourEvents)
	fmt.Println(ourTasks)
	sortTasks(ourTasks, time.Now(), ourEvents)

	fmt.Println(byUrgency(ourTasks))
	for _, task := range ourTasks {
		fmt.Println(task.Name)
	}

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
		if strings.Contains(line, "-") {
			lines = append(lines, line)
		}
	}
	return mdToStructs(lines)
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
	for _, raw := range rawLines {
		line, offset := organizeLines(raw)
		lines = append(lines, line)
		offsets = append(offsets, offset)
	}
	for ind := 0; ind < len(lines); {
		line := lines[ind]
		offset := offsets[ind]
		if line == "Tasks" || line == "Events" {
			newInd := ind
			for newInd < len(lines)-1 {
				newInd++
				newOffset := offsets[newInd]
				if newOffset == offset {
					break
				}
			}
			switch line {
			case "Tasks":
				tasks = append(tasks, mdToTasks(lines[ind+1:newInd], offsets[ind+1:newInd])...)
			case "Events":
				events = append(events, mdToEvents(lines[ind+1:newInd], offsets[ind+1:newInd])...)
			}
			ind = newInd
			continue
		}
		ind++
	}
	return events, tasks
}

func mdToTasks(lines []string, offsets []int) []*Task {
	// offset goes up; that's the name, beginning of new Task
	// offset stays equal or goes down; that's a field
	newTask := &Task{}
	previousOffset := -1
	tasks := []*Task{}
	fmt.Println(lines)
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
		previousOffset = offset
	}
	return tasks
}

func mdToEvents(lines []string, offsets []int) []*GeneralEvent {
	// offset goes up; that's the name, beginning of new Task
	// offset stays equal or goes down; that's a field
	newEvent := &GeneralEvent{}
	previousOffset := -1
	events := []*GeneralEvent{}
	fmt.Println(lines)
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
			default:
				fmt.Println("Whoopsie!")
				fmt.Println(tokens[0])
			}
		}
		previousOffset = offset
	}
	return events
}
