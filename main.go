package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/inconshreveable/log15"
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

var genTextMatcher *regexp.Regexp

func main() {

	logger := log15.New()

	genTextMatcher, _ = regexp.Compile(`\t{2}- \*(.+)\*`)

	quitChan := make(chan bool)

	writeTimer := time.NewTimer(0)
	writeDelay := 20 * time.Second

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	previousTasks := []*Task{}

	refreshList := func() {
		logger.Info("Updating task list")
		ourEvents, ourTasks, err := readFromFile(logger)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		sortTasks(ourTasks, time.Now(), ourEvents, logger)
		if compareLists(previousTasks, ourTasks, logger) {
			writeToFile(ourEvents, ourTasks, logger)
			time.Sleep(1 * time.Second)
			writeTimer.Stop()
		} else {
			logger.Debug("Task list not different enough, skipping write.")
		}
		previousTasks = ourTasks
	}

	// get api password from arguments

	// define handlers
	// quit
	// mostUrgent

	// thread for hourly updates
	// change this to on the hour updates
	// only write update if order of tasks is different
	// file watcher lets us update 20 seconds? after last change

	// refresh list on startup
	refreshList()

	// refresh every hour on the hour
	go func() {
		for {
			waitForTopOfHour()
			refreshList()
		}
	}()

	// refresh on delay after file change
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.Error("event not OK")
					return
				}
				if event.Has(fsnotify.Write) {
					logger.Debug("File has been modified", "event", event.Name)
					writeTimer.Stop()
					writeTimer = time.AfterFunc(writeDelay, refreshList)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					logger.Error("error not OK")
					return
				}
				logger.Error("File watching error", "err", err.Error())
			}
		}
	}()

	// Add a path.
	dir := os.Getenv("NOTESDIR")
	err = watcher.Add(dir + "/" + tasksFile)
	if err != nil {
		logger.Error("Problem watching path", "err", err.Error())
	}

	// run server
	<-quitChan
}

// read/write events to md file
func readFromFile(logger log15.Logger) ([]*GeneralEvent, []*Task, error) {
	dir := os.Getenv("NOTESDIR")
	if dir == "" {
		fmt.Println("Don't forget to set NOTESDIR")
		dir = "~/Documents/Logseq/personal/pages"
		logger.Warn("Empty notes directory, adding default", "default", dir)
	}
	// pull in tasks file
	readFile, err := os.Open(dir + "/" + tasksFile)
	if err != nil {
		logger.Error("Error reading Task file", "err", err.Error())
	}
	defer func() {
		err := readFile.Close()
		if err != nil {
			logger.Warn("Unable to close writing file handler; perhaps restart is needed", "err", err.Error())
		}
	}()
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	lines := []string{}
	for fileScanner.Scan() {
		line := fileScanner.Text()
		lines = append(lines, line)
	}
	if len(lines) < 3 {
		return []*GeneralEvent{}, []*Task{}, errors.New("tried the default notes directory but no dice")
	}
	ev, ta := mdToStructs(lines, logger)
	return ev, ta, nil
}

func writeToFile(events []*GeneralEvent, tasks []*Task, logger log15.Logger) {
	dir := os.Getenv("NOTESDIR")
	if dir == "" {
		fmt.Println("Don't forget to set NOTESDIR")
		dir = "~/Documents/Logseq/personal/pages"
		logger.Warn("Empty notes directory, adding default", "default", dir)
	}
	w, err := os.Create(dir + "/" + tasksFile)
	defer func() {
		err := w.Close()
		if err != nil {
			logger.Warn("Unable to close writing file handler; perhaps restart is needed", "err", err.Error())
		}
	}()
	if err != nil {
		logger.Error("Error writing to Task file", "err", err.Error())
	}
	w.WriteString(fmt.Sprintf("Updated at %s: %s", time.Now().Format(updateLineFmt), whatDayIsIt(time.Now(), logger)))
	w.WriteString(outputTasks(tasks))
	w.WriteString(outputEvents(events))
	logger.Info("Updated To Do List file")
}

func organizeLines(rawLine string) (string, int) {
	tokens := strings.Split(rawLine, "	")
	return strings.Trim(rawLine, "- 	"), len(tokens)
}

func mdToStructs(rawLines []string, logger log15.Logger) ([]*GeneralEvent, []*Task) {
	events := []*GeneralEvent{}
	tasks := []*Task{}
	offsets := []int{}
	lines := []string{}

	// add a newline to the raw text
	rawLines = append(rawLines, "\n")

	// generate list of lines and matching offsets
	for _, raw := range rawLines {
		line, offset := organizeLines(raw)
		lines = append(lines, line)
		offsets = append(offsets, offset)
	}

	// loop through lines until we get to one of the literature headers
	for ind := 0; ind < len(lines); {
		line := lines[ind]
		offset := offsets[ind]

		if line == upcomingTasks || line == overdueTasks || line == completedTasks || line == repeatingEvents || line == inactiveEvents {
			newInd := ind
			// loop over all lines within the header, judged by waiting until the offset matches the header's offset (new potential header)
			for newInd < len(lines)-1 {
				newInd++
				newOffset := offsets[newInd]
				if newOffset == offset {
					// now we have an index corresponding to the end of this section
					break
				}
			}
			switch line {
			case upcomingTasks, overdueTasks, completedTasks:
				tasks = append(tasks, mdToTasks(rawLines[ind+1:newInd], lines[ind+1:newInd], offsets[ind+1:newInd], logger)...)
			case repeatingEvents, inactiveEvents:
				events = append(events, mdToEvents(rawLines[ind+1:newInd], lines[ind+1:newInd], offsets[ind+1:newInd], logger)...)
			}
			ind = newInd
			continue
		}
		ind++
	}
	return events, tasks
}

func mdToTasks(rawLines []string, lines []string, offsets []int, topLogger log15.Logger) []*Task {
	topLogger = topLogger.New("function", "mdToTasks")
	// offset goes up; that's the name, beginning of new Task
	// offset stays equal or goes down; that's a field
	newTask := &Task{}
	tasks := []*Task{}
	namesOffset := -1
	for index, line := range lines {
		line = strings.Trim(line, " ")
		offset := offsets[index]
		if index == 0 {
			namesOffset = offset
		}
		loopLogger := topLogger.New("line", line, "offset", offset, "NamesOffset", namesOffset, "index", index)
		loopLogger.Debug("Analyzing new Task line")
		switch {
		case (offset == namesOffset):
			loopLogger.Debug("Adding Name")
			newTask = &Task{
				Name: line,
			}
			tasks = append(tasks, newTask)
		case offset > namesOffset:
			tokens := strings.Split(line, "; ")
			switch tokens[0] {
			case "Deadline":
				loopLogger.Debug("Adding Deadline", "deadline", tokens[1])
				newTask.Deadline = tokens[1]
			case "Estimated Hours":
				loopLogger.Debug("Adding Estimated Hours", "hours", tokens[1])
				num, err := strconv.Atoi(tokens[1])
				if err != nil {
					loopLogger.Error("Error transforming hours into number", "err", err.Error(), "hours", tokens[1])
				}
				newTask.EstimatedHours = num
			default:
				loopLogger.Debug("Line didn't correspond to Name, Deadline, or Hours; skipping")
			}
		case offset < namesOffset:
			loopLogger.Warn("We're parsing a line that has fewer offsets than the first line did...")
		}
		newTask.AddRaw(rawLines[index])
	}
	return tasks
}

func mdToEvents(rawLines []string, lines []string, offsets []int, topLogger log15.Logger) []*GeneralEvent {
	topLogger = topLogger.New("function", "mdToEvents")
	// offset goes up; that's the name, beginning of new Task
	// offset stays equal or goes down; that's a field
	newEvent := &GeneralEvent{}
	namesOffset := -1
	events := []*GeneralEvent{}
	for index, line := range lines {
		line = strings.Trim(line, " ")
		offset := offsets[index]
		if index == 0 {
			namesOffset = offset
		}
		loopLogger := topLogger.New("line", line, "offset", offset, "namesOffset", namesOffset, "index", index)
		loopLogger.Debug("Analyzing new Event line")
		switch {
		case (offset == namesOffset):
			loopLogger.Debug("Adding Name")
			newEvent = &GeneralEvent{
				Name: line,
			}
			events = append(events, newEvent)
		case offset > namesOffset:
			tokens := strings.Split(line, "; ")
			switch tokens[0] {
			case "Rotation":
				loopLogger.Debug("Adding Rotation", "rotation", tokens[1])
				newEvent.Rotation = rotation(tokens[1])
			case "Days":
				loopLogger.Debug("Adding Days", "days", tokens[1])
				newEvent.Days = tokens[1]
			case "Start Time":
				loopLogger.Debug("Adding Start Time", "start", tokens[1])
				num, err := strconv.Atoi(tokens[1])
				if err != nil {
					loopLogger.Error("Error transforming start time into number", "err", err.Error(), "start", tokens[1])
				}
				newEvent.StartTime = num
			case "Duration":
				loopLogger.Debug("Adding Duration", "duration", tokens[1])
				num, err := strconv.Atoi(tokens[1])
				if err != nil {
					loopLogger.Error("Error transforming duration into number", "err", err.Error(), "duration", tokens[1])
				}
				newEvent.Duration = num
			case "Inactive":
				loopLogger.Debug("Adding Inactivity", "inactive", tokens[1])
				ans, err := strconv.ParseBool(tokens[1])
				if err != nil {
					loopLogger.Error("Error transforming inactivity into boolean", "err", err.Error(), "inactive", tokens[1])
				}
				newEvent.Inactive = ans
			default:
				loopLogger.Debug("Line didn't correspond to Name, Rotation, Days, Start, Duration, or Inactivity; skipping")
			}
		case offset < namesOffset:
			loopLogger.Warn("We're parsing a line that has fewer offsets than the first line did...")

		}
		newEvent.AddRaw(rawLines[index])
	}
	return events
}
