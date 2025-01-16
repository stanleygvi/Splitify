import redis
import os
from fastapi import FastAPI, Request, Response, Depends, HTTPException
from fastapi.responses import RedirectResponse, JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from datetime import timedelta
from urllib.parse import urlparse
from Backend.spotify_api import (
    is_access_token_valid,
    refresh_access_token,
    get_all_playlists,
    exchange_code_for_token,
    get_user_id,
)
from Backend.playlist_processing import process_all
from Backend.helpers import generate_random_string
from starlette.middleware.sessions import SessionMiddleware

url = urlparse(os.environ.get("REDIS_URL"))

db = redis.Redis(
    host=url.hostname,
    port=url.port,
    password=url.password,
    ssl=(url.scheme == "rediss"),
    ssl_cert_reqs=None,
)

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://www.splitifytool.com", "https://splitify-fac76.web.app"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.add_middleware(
    SessionMiddleware,
    secret_key=os.getenv("FLASK_SECRET_KEY"),
    session_cookie="splitify_session",
    max_age=3600,
    https_only=True,
    same_site="none",
)

def get_session(request: Request):
    return request.session

@app.get("/login")
async def login_handler(request: Request):
    session = get_session(request)
    uid = session.get("uid")
    if uid:
        auth_token = db.get(f"{uid}_TOKEN")
        refresh_token = db.get(f"{uid}_REFRESH_TOKEN")
        if not is_access_token_valid(auth_token):
            if refresh_token:
                new_auth_token = refresh_access_token(refresh_token)
                db.set(f"{uid}_TOKEN", new_auth_token)
            else:
                return await redirect_to_spotify_login()

        response = RedirectResponse("https://www.splitifytool.com/input-playlist")
        response.set_cookie(
            "auth_token", auth_token, httponly=True, secure=True, samesite="none"
        )
        return response
    return await redirect_to_spotify_login()

async def redirect_to_spotify_login():
    client_id = os.getenv("CLIENT_ID")
    state = generate_random_string(16)
    scope = "user-read-private playlist-modify-public playlist-read-private"

    params = {
        "response_type": "code",
        "client_id": client_id,
        "scope": scope,
        "show_dialog": "true",
        "redirect_uri": "https://splitify-fac76.web.app/callback",
        "state": state,
    }

    url = "https://accounts.spotify.com/authorize?" + "&".join(
        [f"{key}={value}" for key, value in params.items()]
    )
    return RedirectResponse(url)

@app.get("/callback")
async def callback_handler(code: str):
    if not code:
        raise HTTPException(status_code=400, detail="No code present in callback")

    token_data = exchange_code_for_token(code)

    if not token_data:
        raise HTTPException(status_code=500, detail="Error exchanging code for token")

    auth_token = token_data.get("access_token")
    user_id = get_user_id(auth_token)

    db.set(f"{user_id}_TOKEN", auth_token)
    db.set(f"{user_id}_REFRESH_TOKEN", token_data.get("refresh_token"))

    response = RedirectResponse("https://splitify-fac76.web.app/input-playlist")
    response.set_cookie("auth_token", auth_token, secure=True, samesite="none")

    return response

@app.get("/user-playlists")
async def get_playlist_handler(request: Request):
    auth_token = request.cookies.get("auth_token")

    if not auth_token:
        raise HTTPException(status_code=401, detail="Authorization token required")

    playlists = get_all_playlists(auth_token)

    if not playlists:
        raise HTTPException(status_code=500, detail="Failed to get playlists")

    return JSONResponse(playlists)

@app.post("/process-playlist")
async def process_playlist_handler(request: Request):
    auth_token = request.cookies.get("auth_token")

    if not auth_token or not is_access_token_valid(auth_token):
        raise HTTPException(status_code=401, detail="Authorization required")

    data = await request.json()
    playlist_ids = data.get("playlistIds", [])

    if not playlist_ids:
        raise HTTPException(status_code=400, detail="No playlist IDs provided")

    process_all(auth_token, playlist_ids)

    return JSONResponse({"message": "Playlists processed successfully!"}, status_code=200)

if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(app, host="0.0.0.0", port=port)
