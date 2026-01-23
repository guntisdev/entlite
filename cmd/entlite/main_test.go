package main

import "testing"

func TestBasic(t *testing.T) {
	expected := 1
	actual := 1

	if actual != expected {
		t.Errorf("Expected %d, but got %d", expected, actual)
	}
}
