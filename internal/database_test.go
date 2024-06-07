package internal

import (
	"testing"
)

func TestLoadDB(t *testing.T) {
	db, err := NewDB("./testdata/database.json")
	if err != nil {
		t.Fatalf("no db %s", err)
	}
	params := struct {
		AuthorId int
		Sort     string
	}{AuthorId: 0, Sort: ""}
	chirps, cerr := db.GetChirps(params)
	if cerr != nil {
		t.Fatalf("error loading chirps %s", cerr)
	}
	if len(chirps) > 0 && chirps[len(chirps)-1].Body == "" {
		t.Fatalf("invalid chirp: %v", chirps)
	}
}

func TestCreateChirp(t *testing.T) {
	db, err := NewDB("./testdata/database.json")
	if err != nil {
		t.Fatalf("no db %s", err)
	}
	body := "test chirp body afwef"
	user := User{Id: 1}
	_, cerr := db.CreateChirp(body, user)
	if cerr != nil {
		t.Fatalf("couldnt create chirp: '%s'", body)
	}
	params := struct {
		AuthorId int
		Sort     string
	}{AuthorId: 0, Sort: ""}
	chirps, gerr := db.GetChirps(params)
	if gerr != nil {
		t.Fatalf("couldnt get chirps %s", gerr)
	}
	if len(chirps) != 1 {
		t.Fatalf("expected len: %d actual length: %d", 1, len(chirps))
	}
	body2 := "test chirp body uahwuef"
	_, cerr2 := db.CreateChirp(body2, user)
	if cerr2 != nil {
		t.Fatalf("couldnt create chirp: '%s'", body2)
	}
	chirps2, gerr2 := db.GetChirps(params)
	if gerr2 != nil {
		t.Fatalf("couldnt load chirps %s", gerr2)
	}
	if len(chirps2) != 2 {
		t.Fatalf("expected len: %d actual length: %d", 2, len(chirps2))
	}
}
