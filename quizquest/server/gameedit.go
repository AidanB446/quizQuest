package main

import (
	"net/http"
	"encoding/json"
	"strconv"
)

func nextquestion(w http.ResponseWriter, r *http.Request) {

	var data map[string]string
   
	decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&data); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

	if !gameAuth(data) {
		w.Header().Set("Content-Type", "application/json")
		
		response := make(map[string]string)
		response["status"] = "Auth Failed"

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)
		return 
	}


	questionNumber := GAME_REGISTRY[data["sessionid"]].question_number
	questionNumber++
	
	// remove answer from new question
	question := getQuestion(GAME_REGISTRY[data["sessionid"]].gamename, strconv.Itoa(questionNumber))
	delete(question, "answer")	
	
	// iter through connections and send question map	
	
	connections := GAME_REGISTRY[data["sessionid"]].connections






	w.Header().Set("Content-Type", "application/json")

	response := make(map[string]string)
	response["status"] = "Success"

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}
