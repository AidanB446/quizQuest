package main

import (
	"fmt"
	"slices"
	"math/rand"
	_ "modernc.org/sqlite"
)

func stdErrHandling(err error) {
	if err != nil {
		fmt.Println(err)
	}
}


func createGameDB(gamename string, questions string) {
	gamedbLock.RLock()	
	defer gamedbLock.RUnlock()	

	stmt, err := gamedb.Prepare("INSERT INTO games VALUES (?, ?)")	
	stdErrHandling(err)
	stmt.Exec(gamename, questions)
}

func createGameSessionID() string {

	var charlist string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

	var newSessionID string = ""

	for i := 0; i<10; i++ {
		newSessionID = newSessionID + string(charlist[rand.Intn(len(charlist))])	
	}	

    keys := make([]string, 0, len(GAME_REGISTRY))

	for k := range GAME_REGISTRY {
        keys = append(keys, k)
    }

	if slices.Contains(keys, newSessionID) {
		return createGameSessionID()
	}

	return newSessionID
}

func createNewSessionID(username string) string {

	// revise to avoid rlock deadlock

	var charlist string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

	var newSessionID string = ""

	for i := 0; i<100; i++ {
		newSessionID = newSessionID + string(charlist[rand.Intn(len(charlist))])	
	}	

	userdbLock.RLock()

    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE sessionID = ? LIMIT 1)`
    userdb.QueryRow(query, newSessionID).Scan(&exists)
   
	if exists {
		userdbLock.RUnlock()
		return createNewSessionID(username)
	}	
	
	// add to db
		
	stmt, err := userdb.Prepare("UPDATE users SET sessionID = ? WHERE username = ?")
	stdErrHandling(err)	

	stmt.Exec(newSessionID, username)

	userdbLock.RUnlock()

	return newSessionID
}

func login(username string, password string) (bool, string) {

	userdbLock.RLock()
	defer userdbLock.RUnlock()
	
	var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?, password = ? LIMIT 1)`
    userdb.QueryRow(query, username, password).Scan(&exists)
  	
	if exists {
		return true, createNewSessionID(username)	
	}	

	return false, "" 
}

func createUser(username string, password string) bool {

	userdbLock.RLock()
	defer userdbLock.RUnlock()
	
	var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ? LIMIT 1)`
    userdb.QueryRow(query, username, password).Scan(&exists)
 
	if exists {
		return false
	} else {
		query := "INSERT INTO users VALUES (?, ?, ?)"
		userdb.Exec(query, username, password, "")
		return true
	}	
}

func sessionAuth(username string, token string) bool {

	userdbLock.RLock()
	defer userdbLock.RUnlock()
	
	var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?, sessionID = ? LIMIT 1)`
    userdb.QueryRow(query, username, token).Scan(&exists)

	return exists
} 

