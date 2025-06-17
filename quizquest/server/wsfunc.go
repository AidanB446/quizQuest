package main

import (
	"encoding/json"
	"slices"
	"strconv"
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

	var questions map[string]map[string]string

	err := json.Unmarshal([]byte(questionstring), &questions)	
	stdErrHandling(err)
	
	
	question := questions[strconv.Itoa(GAME_REGISTRY[data["sessionid"]].question_number)]	

	if question["answer"] == data["answer"] {
		// highten user score
		score := GAME_REGISTRY[data["sessionid"]].userboard
		score["username"] = score["username"] + 1

		return true
	}

	return false
}




