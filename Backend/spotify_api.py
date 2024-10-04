import requests
import os


SPOTIFY_API_URL = "https://api.spotify.com/v1"

def spotify_request(method, endpoint, auth_token, params=None, data=None, json_data=None) -> dict:
    url = f"{SPOTIFY_API_URL}{endpoint}"
    headers = {
        "Authorization": f"Bearer {auth_token}",
        "Content-Type": "application/json"
    }

    response = requests.request(method, url, headers=headers, params=params, data=data, json=json_data)
    
    if response.status_code >= 400:
        print(f"Spotify API request error: {response.status_code}, {response.text}")
        return {}
    return response.json()

def is_access_token_valid(auth_token):
    response = spotify_request("GET", "/me", auth_token)
    return response is not None

def refresh_access_token(refresh_token) -> str:
    client_id = os.getenv('CLIENT_ID')
    client_secret = os.getenv('CLIENT_SECRET')
    url = "https://accounts.spotify.com/api/token"
    data = {
        "grant_type": "refresh_token",
        "refresh_token": refresh_token,
        "client_id": client_id,
        "client_secret": client_secret
    }
    headers = {"Content-Type": "application/x-www-form-urlencoded"}
    
    response = requests.post(url, data=data, headers=headers)
    if response.status_code != 200:
        print(f"Error refreshing access token: {response.status_code}, {response.text}")
        return ""

    token_data = response.json()
    return token_data.get('access_token')

def exchange_code_for_token(code):
    client_id = os.getenv('CLIENT_ID')
    client_secret = os.getenv('CLIENT_SECRET')
    url = "https://accounts.spotify.com/api/token"
    data = {
        "grant_type": "authorization_code",
        "code": code,
        "redirect_uri": "https://splitify-app-96607781f61f.herokuapp.com/callback",
        "client_id": client_id,
        "client_secret": client_secret
    }
    headers = {"Content-Type": "application/x-www-form-urlencoded"}
    
    response = requests.post(url, data=data, headers=headers)
    if response.status_code != 200:
        print(f"Error exchanging code for token: {response.status_code}, {response.text}")
        return None

    return response.json()

def get_user_id(auth_token):
    response = spotify_request("GET", "/me", auth_token)
    if response:
        return response.get('id')
    return None

def get_all_playlists(auth_token):
    user_id = get_user_id(auth_token)
    if not user_id:
        return None

    endpoint = f"/users/{user_id}/playlists"
    response = spotify_request("GET", endpoint, auth_token)
    return response

def get_playlist_length(playlist_id, auth_token):
    endpoint = f"/playlists/{playlist_id}/tracks"
    params = {"fields": "total"}
    response = spotify_request("GET", endpoint, auth_token, params=params)
    if response:
        return response.get('total', 0)
    return -1

def get_playlist_name(playlist_id, auth_token):
    endpoint = f"/playlists/{playlist_id}"
    response = spotify_request("GET", endpoint, auth_token)
    if response:
        return response.get("name", "")
    return ""

def get_playlist_children(start_index, playlist_id, auth_token):
    endpoint = f"/playlists/{playlist_id}/tracks"
    params = {
        "offset": start_index,
        "limit": 100,
        "fields": "items(track(name,id,artists(name)))"
    }
    response = spotify_request("GET", endpoint, auth_token, params=params)
    return response

def get_audio_features(track_ids:list[str], auth_token)->list[dict[str,float]]:
    endpoint = "/audio-features"
    params = {
        "ids": ",".join(track_ids)
    }
    response = spotify_request("GET", endpoint, auth_token, params=params)
    assert response
    return response['audio_features']

def create_playlist(user_id, auth_token, name, description):
    endpoint = f"/users/{user_id}/playlists"
    json_data = {
        "name": name,
        "description": description,
        "public": True
    }
    response = spotify_request("POST", endpoint, auth_token, json_data=json_data)
    if response:
        return response.get('id')
    return None

def add_songs(playlist_id, track_uris, auth_token, position):
    endpoint = f"/playlists/{playlist_id}/tracks"
    json_data = {
        "uris": track_uris,
        "position": position
    }
    response = spotify_request("POST", endpoint, auth_token, json_data=json_data)
    if response:
        return response
    return None
