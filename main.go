package main

import (
	"fmt"
)

func main() {
	fmt.Println(parseDayStrings("Tuesday, Thur-Mon"))
}

// run as web server
// read/write events to json file
// read/write blocked days to json file
// when serving, calculate urgency of task or tasks and return
