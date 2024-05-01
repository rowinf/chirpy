package internal

import (
	"testing"
)

func TestLoadDB(t *testing.T) {
	db, err := NewDB("./database.json")
	if err != nil {
		panic("no db")
	}
	chirps, cerr := db.GetChirps()
	if cerr != nil {
		panic("no chirps")
	}
	if len(chirps) == 0 {
		t.Fatalf("err: %s", err)
	}
	t.Log(db)
}
