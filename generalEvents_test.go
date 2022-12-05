package main

import (
	"sort"
	"testing"
)

// test that a list of events returns the correct list of blocked off hours for early, mid, and late-fortnite Nows
func TestGenEventBlockedHours(t *testing.T) {
	times := generateTestingTimes()
	sleepEvent := &GeneralEvent{
		Name:      "sleeping",
		Rotation:  "both",
		Days:      "Sun-Sat",
		StartTime: 23,
		Duration:  8,
	}
	earlySleepHours := []int{336, 24, 48, 72, 96, 120, 144, 168, 192, 216, 240, 264, 288, 312, 337, 25, 49, 73, 97, 121, 145, 169, 193, 217, 241, 265, 289, 313, 338, 26, 50, 74, 98, 122, 146, 170, 194, 218, 242, 266, 290, 314, 339, 27, 51, 75, 99, 123, 147, 171, 195, 219, 243, 267, 291, 315, 340, 28, 52, 76, 100, 124, 148, 172, 196, 220, 244, 268, 292, 316, 341, 29, 53, 77, 101, 125, 149, 173, 197, 221, 245, 269, 293, 317, 342, 30, 54, 78, 102, 126, 150, 174, 198, 222, 246, 270, 294, 318, 23, 47, 71, 95, 119, 143, 167, 191, 215, 239, 263, 287, 311, 335}
	sort.Ints(earlySleepHours)
	midSleepHours := []int{336, 360, 384, 408, 432, 456, 480, 168, 192, 216, 240, 264, 288, 312, 337, 361, 385, 409, 433, 457, 481, 169, 193, 217, 241, 265, 289, 313, 338, 362, 386, 410, 434, 458, 482, 170, 194, 218, 242, 266, 290, 314, 339, 363, 387, 411, 435, 459, 483, 171, 195, 219, 243, 267, 291, 315, 340, 364, 388, 412, 436, 460, 484, 172, 196, 220, 244, 268, 292, 316, 341, 365, 389, 413, 437, 461, 485, 173, 197, 221, 245, 269, 293, 317, 342, 366, 390, 414, 438, 462, 486, 174, 198, 222, 246, 270, 294, 318, 359, 383, 407, 431, 455, 479, 503, 191, 215, 239, 263, 287, 311, 335}
	sort.Ints(midSleepHours)
	lateSleepHours := []int{336, 360, 384, 408, 432, 456, 480, 504, 528, 552, 576, 600, 624, 312, 337, 361, 385, 409, 433, 457, 481, 505, 529, 553, 577, 601, 625, 313, 338, 362, 386, 410, 434, 458, 482, 506, 530, 554, 578, 602, 626, 314, 339, 363, 387, 411, 435, 459, 483, 507, 531, 555, 579, 603, 627, 315, 340, 364, 388, 412, 436, 460, 484, 508, 532, 556, 580, 604, 628, 316, 341, 365, 389, 413, 437, 461, 485, 509, 533, 557, 581, 605, 629, 317, 342, 366, 390, 414, 438, 462, 486, 510, 534, 558, 582, 606, 630, 318, 311, 335, 359, 383, 407, 431, 455, 479, 503, 527, 551, 575, 599, 623}
	sort.Ints(lateSleepHours)
	tests := []struct {
		testNow        string
		generatedHours []int
	}{
		{
			testNow:        "early",
			generatedHours: earlySleepHours,
		},
		{
			testNow:        "mid",
			generatedHours: midSleepHours,
		},
		{
			testNow:        "late",
			generatedHours: lateSleepHours,
		},
	}

	for _, test := range tests {
		hours := getNextBlockedHours(times[test.testNow], []*GeneralEvent{sleepEvent})
		if !compareIntArr(hours, test.generatedHours) {
			t.Errorf("Sleeping hours didn't match for %s;\nExpected: %v\nActual: %v", test.testNow, test.generatedHours, hours)
			t.FailNow()
		}
	}
}

func compareIntArr(arr1, arr2 []int) bool {
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
