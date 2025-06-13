
use rusqlite::{params, Connection};
use rand::prelude::*;

use crate::*;

pub fn create_gamedb(gamename : String, questions : String, conn : rusqlite::Connection) {
    let querystring = "INSERT INTO games VALUES (?, ?);"; 
    let mut stmt = conn.prepare(querystring).unwrap();
    stmt.execute(params![gamename.as_str(), questions.as_str()]).unwrap();
}

pub fn create_game_sessionid() -> String {
    
    // all alpha numeric characters
    let charlist: Vec<char> = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
        .chars()
        .collect();

    let mut rng = rand::rng();
    let mut new_sessionid = String::new();

    for _ in 0..100 {
        let index = rng.random_range(0..charlist.len());
        new_sessionid.push(charlist[index]);
    }
    
    let current_ids : Vec<String> = GAME_REGISTRY.lock().unwrap().keys().cloned().collect(); 

    if current_ids.contains(&new_sessionid) {
        drop(current_ids);    
        return create_game_sessionid();
    } 

    drop(current_ids); 
    return new_sessionid;
}

pub fn create_new_sessionid(username : String) -> String {
    
    // all alpha numeric characters
    let charlist: Vec<char> = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
        .chars()
        .collect();

    let mut rng = rand::rng();
    let mut new_sessionid = String::new();

    for _ in 0..100 {
        let index = rng.random_range(0..charlist.len());
        new_sessionid.push(charlist[index]);
    }
    
    let conn = Connection::open("./users.db").unwrap();
    
    let query = "SELECT * FROM users WHERE sessionID = ?";
    
    let mut stmt = conn.prepare(query).unwrap();
    
    let exists: bool = stmt.query_row(params![new_sessionid], |row| row.get(0)).unwrap();
    
    drop(stmt);


    if exists {
        conn.close().unwrap();
        return create_new_sessionid(username);
    } 
        
    // update database
    let query = "UPDATE users SET sessionID = ? WHERE username = ?"; 
    
    let mut stmt = conn.prepare(query).unwrap(); 

    stmt.execute(params![new_sessionid, username]).unwrap();

    return new_sessionid;
}

// returns true or false and a sessionID if permitted 
pub fn login(username : String, password : String) -> (bool, String) {
    
    // check to see if username and password match, return tuple    

    let conn = Connection::open("./users.db").unwrap();
    
    let query = "SELECT username, password FROM users WHERE username = ? AND password = ?";
    
    let mut stmt = conn.prepare(query).unwrap();
    
    let exists: bool = stmt.query_row(params![username, password], |row| row.get(0)).unwrap();
    
    drop(stmt);

    if exists {
        // setup and update session in database
        
        let query = "UPDATE users SET sessionID = ? WHERE username = ?";
        
        let mut stmt = conn.prepare(query).unwrap();
        
        let new_sessionid = create_new_sessionid(username.clone());

        stmt.execute(params![new_sessionid.clone(), username]).unwrap();
        
        drop(stmt); 
    
        return (true, new_sessionid);
    } 

    conn.close().unwrap(); 

    return (false, String::from(""));
}

pub fn create_user(username : String, password : String) -> bool {
    
    let conn = Connection::open("./users.db").unwrap();

    let query = "SELECT EXISTS(SELECT 1 FROM users WHERE username= ?)";
    
    let mut stmt = conn.prepare(query).unwrap();
    let exists: bool = stmt.query_row(params![username], |row| row.get(0)).unwrap();
    
    drop(stmt); // connection is freed
    
    if exists {
        return false;
    } else {
        let query = "INSERT INTO users VALUES (?, ?, ?)";
        
        let mut stmt = conn.prepare(query).unwrap();
        
        stmt.execute(params![username, password, String::from("Undefined")]).unwrap();

        return true;     
    }

}

pub fn session_auth(username : String, token : String) {
    
    



}

