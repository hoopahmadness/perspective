package main

import (
	"encoding/csv"
	"time"
)

type Weekday int

const (
	SUN = Weekday(iota)
	MON
	TUES
	WED
	THUR
	FRI
	SAT
)

type GeneralEvents struct {
	Name      string `csv:"Name"`
	Day       string `csv:"Day"`
	Hours     string `csv:"Hours"`
	datetimes []time.Time
}
