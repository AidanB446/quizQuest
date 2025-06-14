package main

import (
	"encoding/json"
	"slices"

	"github.com/gorilla/websocket"
)

func joinGame(data map[string]string, ws *websocket.Conn) bool  {

	// sessionAuth
	sessionid := data["sessionid"] 
	
	keys := make([]string, 0, len(GAME_REGISTRY))
	for k := range GAME_REGISTRY {
		keys = append(keys, k)
	}

	if !sessionAuth(data["username"], data["token"]) || !slices.Contains(keys, sessionid) {
		return false	
	}

	gameToModify := GAME_REGISTRY[sessionid]
	gameToModify.connections = append(gameToModify.connections, ws)
	gameToModify.userboard[data["username"]] = 0

	return true
}

func submitQuestion(data map[string]string) bool {

	// expect data['questionNumber']
	// expect data["answer"] expect 'a' 'b' 'c' 'd'
	// expect sessionid 
	// expect username
	// token
	
	gamedbLock.RLock()
	defer gamedbLock.RUnlock()

	if !sessionAuth(data["username"], data["token"]) {
		return false
	}
	
	gamename := GAME_REGISTRY[data["sessionid"]].gamename	

	var questionstring string		
	
	query := "SELECT questions FROM games WHERE gamename = ?"
	row := gamedb.QueryRow(query, gamename)	
	row.Scan(&questionstring)

	var questions map[string]string

	err := json.Unmarshal([]byte(questionstring), &questions)	
	stdErrHandling(err)
	
	var question string = questions[string(data["question_number"])]	
	
	var mqs map[string]string
	
	er := json.Unmarshal([]byte(question), &mqs)	
	stdErrHandling(er)
	
	if mqs["answer"] == data["answer"] {
		// highten user score
		score := GAME_REGISTRY[data["sessionid"]].userboard
		score["username"] = score["username"] + 1

		return true
	}

	return false
}


func startGame(data map[string]string) bool {
	
	// check if gamename exists in db
	
	gamedbLock.RLock()
	defer gamedbLock.RUnlock()

	var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM games WHERE gamename = ? LIMIT 1)`
    gamedb.QueryRow(query, data["gamename"]).Scan(&exists)
  	
	if !exists {
		return false
	}

	newSessionID := createGameSessionID()

	newGame := ActiveGame {
		data["gamename"],
		newSessionID,
		data["token"],
		make(map[string]int),
		make([]*websocket.Conn, 0),	
		1,
	}	

	GAME_REGISTRY[newSessionID] = newGame

	return true
}





