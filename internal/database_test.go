package internal

import (
	"testing"
)

func TestLoadDB(t *testing.T) {
	chirps, err := GetChirps()
	if len(chirps) == 0 {
		t.Fatalf("err: %s", err)
	}
}
