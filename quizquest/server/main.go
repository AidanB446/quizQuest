package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"sync"
	_ "modernc.org/sqlite"
)

type ActiveGame struct {
	gamename string
	session_id string
	admintoken string
	userboard map[string]int
	connections []*websocket.Conn
	question_number int
}

var GAME_REGISTRY map[string]ActiveGame

// call dbLock.RLock() and defer dbLock.RUnlock()
// before using a db connection
var (
	gamedb *sql.DB
	gamedbLock sync.RWMutex	
	
	userdb *sql.DB
	userdbLock sync.RWMutex	
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // For demo only
}

func postHandler(w http.ResponseWriter, r *http.Request) {

	var data map[string]any

    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&data); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    log.Printf("Received POST JSON: %+v\n", data)
    fmt.Fprintf(w, "Data received: %+v\n", data)
}

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

func main() {

	userdb, _ = sql.Open("sqlite", "./users.db")
	gamedb, _ = sql.Open("sqlite", "./game.db")

	GAME_REGISTRY = make(map[string]ActiveGame)

	http.HandleFunc("/ws", wsHandler)
    http.HandleFunc("/submit", postHandler)

    log.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))

}

