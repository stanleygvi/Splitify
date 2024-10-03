from flask import Flask, request, redirect, jsonify, session
from datetime import timedelta
from flask_cors import CORS
import redis
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
CORS(app, supports_credentials=True, origins=["https://splitifytool.com"])

app.config["SECRET_KEY"] = os.getenv("FLASK_SECRET_KEY", "supersecretkey")
app.config["SESSION_TYPE"] = "redis"
app.config["SESSION_REDIS"] = redis.from_url(os.getenv("REDIS_URL"))
app.config["PERMANENT_SESSION_LIFETIME"] = timedelta(hours=1) # Life of auth token for spotify
app.config["SESSION_COOKIE_SAMESITE"] = "None"
app.config["SESSION_COOKIE_SECURE"] = True
app.config["SESSION_COOKIE_HTTPONLY"] = True

app.config["SESSION_COOKIE_SECURE"] = True
Session(app)


@app.before_request
def before_request():
    if not request.is_secure and os.getenv("FLASK_ENV") == "production":
        return redirect(request.url.replace("http://", "https://"))
@app.after_request
def apply_cors(response):
    response.headers["Access-Control-Allow-Origin"] = "https://splitifytool.com"
    response.headers["Access-Control-Allow-Credentials"] = "true"
    response.headers["Cache-Control"] = "no-store, no-cache, must-revalidate, public, max-age=0"
    response.headers["Pragma"] = "no-cache"
    response.headers["Expires"] = "0"
    return response

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

    assert request.json
    playlist_ids = request.json.get("playlistIds", [])
    
    if not playlist_ids:
        return "No playlist IDs provided", 400

    process_playlists(auth_token, playlist_ids)

    return jsonify({"message": "Playlists processed successfully!"}), 200

if __name__ == "__main__":
    port = os.getenv("PORT", "8080")
    assert type(port) == int
    app.run(host="0.0.0.0", port=port)
