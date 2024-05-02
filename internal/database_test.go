package internal

import (
	"testing"
)

func TestLoadDB(t *testing.T) {
	db, err := NewDB("./database.json")
	if err != nil {
		t.Fatalf("no db %s", err)
	}
	chirps, cerr := db.GetChirps()
	if cerr != nil {
		t.Fatalf("error loading chirps %s", cerr)
	}
	if len(chirps) == 0 {
		t.Fatal("no chirps loaded")
	}
	t.Log(chirps)
	if len(chirps) > 0 && chirps[len(chirps)-1].Body == "" {
		t.Fatalf("invalid chirp: %v", chirps)
	}
}

func TestCreateChirp(t *testing.T) {
	db, err := NewDB("./database.json")
	if err != nil {
		t.Fatalf("no db %s", err)
	}
	body := "test chirp body"
	if chirp, cerr := db.CreateChirp(body); cerr == nil {
		if chirp.Body == body {
			t.Logf("created chirp %s", body)
		} else {
			t.Fatalf("couldnt create chirp: '%s'", body)
		}
	} else {
		t.Fatalf("couldnt create chirp: '%s'", body)
	}
}
