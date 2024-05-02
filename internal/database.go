package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := DB{path, &sync.RWMutex{}}
	db.ensureDB()
	return &db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	newChirp := Chirp{Body: body}
	if dbStructure, err := db.loadDB(); err == nil {
		newChirp.Id = len(dbStructure.Chirps)
		dbStructure.Chirps[len(dbStructure.Chirps)] = newChirp
		if werr := db.writeDB(dbStructure); werr == nil {
			return newChirp, nil
		} else {
			return newChirp, werr
		}
	}
	return newChirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		values := make([]Chirp, len(data.Chirps))
		for _, val := range data.Chirps {
			values = append(values, val)
		}
		return values, nil
	}
	return nil, errors.New("doh")
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	f, err := os.OpenFile(db.path, os.O_RDWR, 0755)
	if err != nil {
		os.Create(db.path)
		return nil
	}
	f.Close()
	return err
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	bytes, err := os.ReadFile(db.path)
	if err != nil {
		chirps := DBStructure{Chirps: make(map[int]Chirp)}
		err := json.Unmarshal(bytes, &chirps)
		if err != nil {
			return DBStructure{Chirps: make(map[int]Chirp)}, errors.New("couldnt load json")
		}
		return chirps, nil
	}
	fmt.Printf("log.Logger: %v\n", bytes)
	return DBStructure{Chirps: make(map[int]Chirp)}, errors.New("couldnt load json")
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	j, err := json.Marshal(dbStructure)
	if err != nil {
		panic("json error")
	}
	werr := os.WriteFile(db.path, j, 0666)
	return werr
}
