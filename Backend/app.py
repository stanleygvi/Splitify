from flask import Flask, request, redirect, jsonify, url_for, make_response, session
from flask_session import Session
from datetime import timedelta
from flask_cors import CORS
from urllib.parse import urlparse
import redis
import os
from Backend.spotify_api import (
    is_access_token_valid,
    refresh_access_token,
    get_all_playlists,
    exchange_code_for_token,
    get_user_id,
)
from Backend.playlist_processing import process_playlists
from Backend.helpers import generate_random_string
url = urlparse(os.environ.get("REDIS_URL"))

db = redis.Redis(host=url.hostname, port=url.port, password=url.password, ssl=(url.scheme == "rediss"), ssl_cert_reqs=None)

app = Flask(__name__)

app.config["SECRET_KEY"] = os.getenv("FLASK_SECRET_KEY")
app.config["SESSION_TYPE"] = "redis"
app.config["SESSION_PERMANENT"] = False
app.config["SESSION_USE_SIGNER"] = True
app.config["PERMANENT_SESSION_LIFETIME"] = timedelta(hours=1)

app.config["SESSION_REDIS"] = db

app.config["SESSION_COOKIE_DOMAIN"] = ".splitifytool.com"
app.config["SESSION_COOKIE_SECURE"] = True
app.config["SESSION_COOKIE_SAMESITE"] = "None"

redis_url = os.getenv("REDIS_URL")
sess = Session()
sess.init_app(app)

CORS(app, origins=["https://www.splitifytool.com", "https://splitify-fac76.web.app"], supports_credentials=True)




@app.route("/login")
def login_handler():
    uid = session.get("uid")
    if uid:
        auth_token = db.get(f"{uid}_TOKEN")
        refresh_token = db.get(f"{uid}_REFRESH_TOKEN")
        if not is_access_token_valid(auth_token):
            if refresh_token:
                new_auth_token = refresh_access_token(refresh_token)
                db.set(f"{uid}_TOKEN", new_auth_token)
            else:
                return redirect_to_spotify_login()

        response = make_response(redirect("https://www.splitifytool.com/input-playlist"))
        response.set_cookie(
            "auth_token", auth_token, httponly=True, secure=True, samesite="None"
        )
        return response
    return redirect_to_spotify_login()


def redirect_to_spotify_login():
    client_id = os.getenv("CLIENT_ID")
    state = generate_random_string(16)
    scope = "user-read-private playlist-modify-public playlist-read-private"

    params = {
        "response_type": "code",
        "client_id": client_id,
        "scope": scope,
        "show_dialog": "true",
        "redirect_uri": url_for("callback_handler", _external=True),
        "state": state,
    }

    url = "https://accounts.spotify.com/authorize?" + "&".join(
        [f"{key}={value}" for key, value in params.items()]
    )
    return redirect(url)


@app.route("/callback")
def callback_handler():
    code = request.args.get("code")

    if not code:
        return "No code present in callback", 400

    token_data = exchange_code_for_token(code)

    if not token_data:
        return "Error exchanging code for token", 500

    auth_token = token_data.get("access_token")
    user_id = get_user_id(auth_token)

    db.set(f"{user_id}_TOKEN", auth_token)
    db.set(f"{user_id}_REFRESH_TOKEN", token_data.get("refresh_token"))

    response = make_response(redirect("https://splitify-fac76.web.app/input-playlist"))
    response.set_cookie(
        "auth_token", auth_token, httponly=True, secure=True, samesite="None"
    )

    return response


@app.route("/user-playlists")
def get_playlist_handler():
    auth_token = request.cookies.get("auth_token")

    if not auth_token:
        print(f"NO AUTH: {auth_token}")
        return {"Code": 401, "Error": "Authorization token required"}

    playlists = get_all_playlists(auth_token)

    if not playlists:
        return {"Code": 500, "Error": "Failed to get playlists"}

    return jsonify(playlists)


@app.route("/process-playlist", methods=["POST"])
def process_playlist_handler():

    auth_token = request.cookies.get("auth_token")

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
    app.run(host="0.0.0.0", port=int(port))
