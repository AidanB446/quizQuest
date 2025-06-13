
import sqlite3

conn = sqlite3.connect("./users.db.db")

cur = conn.cursor()

cur.execute("CREATE TABLE users (username TEXT, password TEXT, sessionID TEXT)")

cur.close()


conn.commit()

conn.close()

