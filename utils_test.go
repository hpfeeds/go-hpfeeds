package hpfeeds

import (
	"testing"
)

var testEmptySlice = []string{}

var testSlice1 = []string{"asdf"}

var testSlice2 = []string{"1", "2", "3", "4"}

func TestUtils_stringInSlice(t *testing.T) {
	if stringInSlice("asdf", testEmptySlice) {
		t.Error("String not found in slice when expected.")
	}
	if stringInSlice("", testEmptySlice) {
		t.Error("String not found in slice when expected.")
	}

	if !stringInSlice("asdf", testSlice1) {
		t.Error("String not found in slice when expected.")
	}
	if stringInSlice("not_found", testSlice1) {
		t.Error("String found in slice when NOT expected.")
	}

	if !stringInSlice("2", testSlice2) {
		t.Error("String not found in slice when expected.")
	}
	if stringInSlice("5", testSlice2) {
		t.Error("String not found in slice when expected.")
	}
}
