package main

import (
	"database/sql"
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


// key string is the games sessionid
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

func main() {

	userdb, _ = sql.Open("sqlite", "./users.db")
	gamedb, _ = sql.Open("sqlite", "./game.db")

	GAME_REGISTRY = make(map[string]ActiveGame)

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/start-game", startGame)	
	http.HandleFunc("/create-game", createGame)	
	http.HandleFunc("/next-question", nextquestion)	
	http.HandleFunc("/game-end", gameEnd)	


    log.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

