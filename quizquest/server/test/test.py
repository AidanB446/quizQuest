
import requests

url = "http://localhost:3030/create-game"

r = requests.post(url, json={"gamename": "tutorial", "questions": "{}"})

print(r.json())



