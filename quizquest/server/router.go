
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()

    for {

		_, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }
       
		log.Println("Received via WS:", string(msg))
	
		var data map[string]string				
		
		er := json.Unmarshal(msg, &data)
		stdErrHandling(er)
	
		switch data["function"] {
		
		case "join-game":

			if joinGame(data, conn) {
				newMap := make(map[string]string)
				newMap["status"] = "success"				
				writedata, _ := json.Marshal(newMap)
				conn.WriteMessage(websocket.TextMessage, writedata)				
			} else {
				newMap := make(map[string]string)
				newMap["status"] = "failed"				
				writedata, _ := json.Marshal(newMap)
				conn.WriteMessage(websocket.TextMessage, writedata)				
			}

		case "submit-question" :

			userResult := submitQuestion(data)	
			
			if userResult {
				newMap := make(map[string]string)
				newMap["status"] = "success"				
				newMap["result"] = "correct"				
				writedata, _ := json.Marshal(newMap)
				conn.WriteMessage(websocket.TextMessage, writedata)				
			} else {
				newMap := make(map[string]string)
				newMap["status"] = "success"
				newMap["result"] = "wrong"
				writedata, _ := json.Marshal(newMap)
				conn.WriteMessage(websocket.TextMessage, writedata)				
			}
			
		default :
			fmt.Println("do something")
		}

		// conn.WriteMessage(websocket.TextMessage, []byte)
        
    }
}

func startGame(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Content-Type", "application/json")
	
	var data map[string]string

    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&data); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
    }

	gamedbLock.RLock()
	defer gamedbLock.RUnlock()

	var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM games WHERE gamename = ? LIMIT 1)`
    gamedb.QueryRow(query, data["gamename"]).Scan(&exists)
  	
	if !exists {
		response := make(map[string]string)	
		response["status"] = "Game Name Does Not Exist"

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)

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
	
	response := make(map[string]string)
	response["status"] = "Game Created"

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

