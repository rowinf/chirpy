package main

import "testing"

func TestCensorString(t *testing.T) {
	case1 := "ready for a kerfuffle"
	s := CensorString(case1)
	if s != "ready for a ****" {
		t.Fatalf("wrong: %s -> %s", case1, s)
	}
	case2 := "ready for a Kerfuffle!"
	s2 := CensorString(case2)
	if s2 != "ready for a ****!" {
		t.Fatalf("wrong: %s -> %s", case2, s2)
	}
}
