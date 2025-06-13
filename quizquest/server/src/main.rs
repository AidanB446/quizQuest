
mod db;
mod livegame;

use db::{create_game_sessionid, create_gamedb, create_new_sessionid};

use warp::Filter;
use warp::ws::{Message, WebSocket};
use futures_util::{SinkExt, StreamExt, TryFutureExt};
use rusqlite::{params, Connection};
use std::collections::HashMap;
use std::intrinsics::discriminant_value;
use std::sync::{Arc, Mutex};
use once_cell::sync::Lazy;
use serde_json::{Result, Value};

use crate::db::session_auth;

#[derive(serde::Deserialize, serde::Serialize)]
struct Response {
    status: String,
    message: String,
}

#[derive(serde::Deserialize, serde::Serialize)]
struct GameSubmission {
    gamename : String,
    questions : String,
}

#[derive(serde::Deserialize, serde::Serialize)]
struct CreateGameField{
    gamename : String,
    admintoken : String,
}

struct ActiveGame {
    gamename : String,    
    session_id : String,    
    admintoken : String,    
    userboard : HashMap<String, i32>,
    connections : Vec<Arc<Mutex<WebSocket>>>,
    question_number : i8,
}

// GAME_REGISTRY.lock().unwrap() to obtain the hashmap
// will auto unlock when out of scope

static GAME_REGISTRY: Lazy<Arc<Mutex<HashMap<String, ActiveGame>>>> = Lazy::new(|| {
    Arc::new(Mutex::new(HashMap::new()))
});

#[tokio::main]
async fn main() {
    
    // WebSocket route
    let ws_route = warp::path("ws")
        .and(warp::ws())
        .map(|ws: warp::ws::Ws| {
            ws.on_upgrade(handle_socket)
    });
    
    let submit_route = warp::path("start-game")
        .and(warp::post())
        .and(warp::body::json())
        .map(|game : CreateGameField| {
            let game_name = game.gamename;      
            let admin_token = game.admintoken;
            let game_sessionid = create_game_sessionid();

            let new_game = ActiveGame {
                gamename : game_name,
                session_id : game_sessionid.clone(),
                admintoken : admin_token,
                userboard : HashMap::new(),
                connections : Vec::new(),
                question_number : 0.into(),
            };

            GAME_REGISTRY.lock().unwrap().insert(game_sessionid, new_game);
            
            let new_response = Response {
                status : String::from(""),
                message : String::from("")
            };
            
            return warp::reply::json(&new_response);
        }); 

    // POST /submit with JSON body
    let submit_route = warp::path("create-game")
        .and(warp::post())
        .and(warp::body::json())
        .map(|submission: GameSubmission| {
            
            // first check if value exists already, if so handle, if not create

            let connection = Connection::open("./game.db").unwrap();

            let query = "SELECT EXISTS(SELECT 1 FROM games WHERE gamename= ?)";
            
            let mut stmt = connection.prepare(query).unwrap();
            let exists: bool = stmt.query_row(params![submission.gamename], |row| row.get(0)).unwrap();
            
            drop(stmt);

            if exists {
                let new_response = Response {
                    status : String::from("error"),
                    message : String::from("already exists"),
                }; 
                
                return warp::reply::json(&new_response);
            } else {
            
                create_gamedb(submission.gamename, submission.questions, connection);

                let new_response = Response {
                    status : String::from("success"),
                    message : String::from(""),
                }; 
                
                return warp::reply::json(&new_response);
            }

        });

    // Combine routes
    let routes = ws_route.or(submit_route);

    println!("🚀 Server running on http://localhost:3030");
    warp::serve(routes).run(([127, 0, 0, 1], 3030)).await;
}

async fn handle_socket(ws: WebSocket) {
    
    let (mut tx, mut rx) = ws.split();

    while let Some(Ok(msg)) = rx.next().await {

        // write event flow 
        // joingame 


        // expect msg is json, parse json    
        // every message will have a function

        if msg.is_text() {
            let received = msg.to_str().unwrap();
            let data : Value = serde_json::from_str(received).unwrap(); 
            let function = &data.clone()["function"]; 
            if function == "join game" {
                
                let username = &data.clone()["username"]; 
                let token = &data.clone()["token"]; 
                let sessionid = &data.clone()["sessionid"]; 

                // call session auth                 
                if !session_auth(username.to_string(), token.to_string()) {
                    tx.send(Message::text("Auth Failed")).await.unwrap();
                    continue; 
                } 

                let activegames = GAME_REGISTRY.lock().unwrap().keys().cloned().collect::<Vec<String>>();
                
                if !activegames.contains(&sessionid.to_string()) {
                    tx.send(Message::text("Game Does Not Exist")).await.unwrap();
                    continue; 
                }
                
                drop(activegames);

            }
                

        }

    }
}

