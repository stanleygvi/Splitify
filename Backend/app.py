from flask import Flask, request, redirect, jsonify, session
from redis import Redis
from flask_cors import CORS
import os
from Backend.spotify_api import (
    is_access_token_valid,
    refresh_access_token,
    get_all_playlists,
    exchange_code_for_token
)
from Backend.playlist_processing import process_playlists
from Backend.helpers import generate_random_string
from flask import url_for
from flask_session import Session

app = Flask(__name__)
CORS(app, origins=["https://splitifytool.com"])

app.config["SECRET_KEY"] = os.getenv("FLASK_SECRET_KEY", "supersecretkey")
app.config["SESSION_TYPE"] = "redis"
redis_url = os.getenv("REDIS_URL")
assert redis_url
app.config["SESSION_REDIS"] = Redis.from_url(redis_url)
Session(app)

@app.route("/login")
def login_handler():
    auth_token = session.get("TOKEN")
    
    if auth_token and is_access_token_valid(auth_token):
        return redirect("https://splitifytool.com/input-playlist")
    
    refresh_token = session.get("REFRESH_TOKEN")
    
    if refresh_token:
        new_access_token = refresh_access_token(refresh_token)
        
        if new_access_token:
            session["TOKEN"] = new_access_token

            return redirect("https://splitifytool.com/input-playlist")
    
    return redirect_to_spotify_login()

def redirect_to_spotify_login():
    client_id = os.getenv("CLIENT_ID")
    state = generate_random_string(16)
    scope = "user-read-private user-read-email playlist-modify-public playlist-modify-private playlist-read-collaborative"
    
    params = {
        "response_type": "code",
        "client_id": client_id,
        "scope": scope,
        "show_dialog": "true",
        "redirect_uri": url_for('callback_handler', _external=True),
        "state": state
    }
    
    url = "https://accounts.spotify.com/authorize?" + "&".join([f"{key}={value}" for key, value in params.items()])
    return redirect(url)

@app.route("/callback")
def callback_handler():
    code = request.args.get("code")
    
    if not code:
        return "No code present in callback", 400

    token_data = exchange_code_for_token(code)
    
    if not token_data:
        return "Error exchanging code for token", 500

    session["TOKEN"] = token_data.get("access_token")
    session["REFRESH_TOKEN"] = token_data.get("refresh_token")

    return redirect("https://splitifytool.com/input-playlist")

@app.route("/user-playlists")
def get_playlist_handler():
    auth_token = session.get("TOKEN")
    print(f"Access Token: {session.get('TOKEN')}")
    print(f"Refresh Token: {session.get('REFRESH_TOKEN')}")
    if not auth_token or not is_access_token_valid(auth_token):
        refresh_token = session.get("REFRESH_TOKEN")
        
        if refresh_token:
            new_access_token = refresh_access_token(refresh_token)
            
            if new_access_token:
                session["TOKEN"] = new_access_token
                auth_token = new_access_token
            else:
                return {"Code": 401, "Error": "Failed to refresh access token"}
        else:
            return {"Code": 401, "Error": "Authorization required. Please log in."}

    playlists = get_all_playlists(auth_token)
    
    if not playlists:
        return {"Code": 500, "Error": "Failed to get playlists"}

    return jsonify(playlists)

@app.route("/process-playlist", methods=["POST"])
def process_playlist_handler():
    auth_token = session.get("TOKEN")

    if not auth_token or not is_access_token_valid(auth_token):
        return "Authorization required", 401

    playlist_ids = request.json.get("playlistIds", [])
    
    if not playlist_ids:
        return "No playlist IDs provided", 400

    process_playlists(auth_token, playlist_ids)

    return jsonify({"message": "Playlists processed successfully!"}), 200

if __name__ == "__main__":
    port = os.getenv("PORT", "8080")
    app.run(host="0.0.0.0", port=port)
