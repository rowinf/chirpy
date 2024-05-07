package internal

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
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
	Chirps    map[int]Chirp  `json:"chirps"`
	Users     map[int]User   `json:"users"`
	Passwords map[int][]byte `json:"passwords"`
}

type User struct {
	Email string `json:"email"`
	Id    int    `json:"id"`
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
	newChirp := Chirp{Body: body, Id: -1}
	if dbStructure, err := db.loadDB(); err == nil {
		for i := range dbStructure.Chirps {
			if _, ok := dbStructure.Chirps[i]; !ok {
				newChirp.Id = i
				break
			}
		}
		if newChirp.Id == -1 {
			newChirp.Id = len(dbStructure.Chirps) + 1
		}

		dbStructure.Chirps[newChirp.Id] = newChirp
		werr := db.writeDB(dbStructure)
		return newChirp, werr
	}
	return newChirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirp(chirpId int) (Chirp, error) {
	data, err := db.loadDB()
	chirp := Chirp{}
	if err == nil {
		for _, val := range data.Chirps {
			if val.Id == chirpId {
				return val, nil
			}
		}
		return chirp, errors.New("not found")
	}
	return chirp, err
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	data, err := db.loadDB()
	values := make([]Chirp, len(data.Chirps))
	if err == nil {
		i := 0
		for _, val := range data.Chirps {
			values[i] = val
			i++
		}
		return values, nil
	}
	return values, nil
}

func (db *DB) CreateUser(email string, password []byte) (User, error) {
	pw, _ := bcrypt.GenerateFromPassword(password, 10)
	newUser := User{Email: email, Id: -1}
	if dbStructure, err := db.loadDB(); err == nil {
		for i := range dbStructure.Users {
			if _, ok := dbStructure.Users[i]; !ok {
				newUser.Id = i
				break
			}
		}
		if newUser.Id == -1 {
			newUser.Id = len(dbStructure.Users) + 1
		}
		dbStructure.Users[newUser.Id] = newUser
		dbStructure.Passwords[newUser.Id] = pw
		werr := db.writeDB(dbStructure)
		return newUser, werr
	}

	return newUser, nil
}

func (db *DB) UserLogin(email string, password []byte) (User, error) {
	user := User{}
	if dbStructure, err := db.loadDB(); err == nil {
		for i := range dbStructure.Users {
			if val, ok := dbStructure.Users[i]; ok {
				if val.Email == email {
					pw := dbStructure.Passwords[i]
					err := bcrypt.CompareHashAndPassword(pw, password)
					if err != nil {
						return user, err
					} else {
						return val, nil
					}
				}
			}
		}
	}
	return user, errors.New("db error")
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	f, err := os.OpenFile(db.path, os.O_RDWR, 0755)
	if err != nil {
		_, cerr := os.Create(db.path)
		if cerr != nil {
			panic("couldnt create file")
		}
		return nil
	}
	f.Close()
	return err
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	bytes, err := os.ReadFile(db.path)
	chirpsDb := DBStructure{
		Chirps:    make(map[int]Chirp),
		Users:     make(map[int]User),
		Passwords: make(map[int][]byte),
	}
	if err == nil {
		var uerr error
		if len(bytes) > 0 {
			uerr = json.Unmarshal(bytes, &chirpsDb)
		}
		return chirpsDb, uerr
	}
	return chirpsDb, err
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
