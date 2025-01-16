from threading import Lock
from concurrent.futures import ThreadPoolExecutor
from collections import defaultdict, Counter
from Backend.spotify_api import (
    get_playlist_length,
    get_playlist_children,
    create_playlist,
    add_songs,
    get_user_id,
    get_playlist_name,
    get_artists,
)
from Backend.helpers import calc_slices
import asyncio
import time

def fetch_batch(start_index, playlist_id, auth_token):
    response = get_playlist_children(start_index, playlist_id, auth_token)
    if response and "items" in response:
        tracks = response["items"]
        track_to_artists = {
            track["track"]["id"]: [artist["id"] for artist in track["track"]["artists"]]
            for track in tracks
            if track["track"] and "artists" in track["track"]
        }
        return track_to_artists
    return {}

def fetch_genres_chunk(chunk, track_genres, genre_lock, auth_token):
    chunk_artist_ids = [artist for track, artist_list in chunk for artist in artist_list]
    artist_genres = asyncio.run(get_artist_details(chunk_artist_ids, auth_token))
    for track_id, artist_list in chunk:
        genres = set()
        for artist_id in artist_list:
            genres.update(artist_genres.get(artist_id, {}).get("genres", []))
        with genre_lock:
            track_genres[track_id] = list(genres)

def assign_genres_to_tracks(auth_token, playlist_id):
    """Assign genres to each track and return a mapping of track_id to genres."""
    slices = calc_slices(get_playlist_length(playlist_id, auth_token))
    track_genres = {}
    genre_lock = Lock()

    all_artists = {}
    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = [executor.submit(fetch_batch, i, playlist_id, auth_token) for i in range(0, slices * 100, 100)]
        for future in futures:
            all_artists.update(future.result())

    artist_chunks = list(all_artists.items())
    batch_size = 50

    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = [
            executor.submit(fetch_genres_chunk, artist_chunks[i:i + batch_size], track_genres, genre_lock, auth_token)
            for i in range(0, len(artist_chunks), batch_size)
        ]
        for future in futures:
            future.result()

    return track_genres

async def get_artist_details(artist_ids, auth_token):
    """Fetch artist details, specifically genres, for a list of artist IDs."""
    artist_details = {}
    chunk_size = 50

    async def fetch_chunk(artist_chunk):
        return get_artists(artist_chunk, auth_token)

    tasks = [
        fetch_chunk(artist_ids[i:i + chunk_size])
        for i in range(0, len(artist_ids), chunk_size)
    ]

    for response in await asyncio.gather(*tasks):
        if response and "artists" in response:
            for artist in response["artists"]:
                artist_details[artist["id"]] = {
                    "genres": artist.get("genres", [])
                }
        else:
            print(f"Failed to fetch details for some artist chunks.")

    return artist_details

def sort_genres_by_count(track_genres):
    """Return a sorted list of genres by frequency in ascending order."""
    genre_counter = Counter(genre for genres in track_genres.values() for genre in genres)
    return sorted(genre_counter.items(), key=lambda x: x[1])

def add_tracks_to_playlist(playlist_id, track_uris, position, auth_token, playlist_name, genre):
    status = add_songs(playlist_id, track_uris, auth_token, position)
    time.sleep(0.5)
    if not status or status.get("Error", None):
        print(
            f"Append Error: Playlist {playlist_name} - {genre}, status {status}, starting from index: {position}"
        )

def create_and_populate_subgenre_playlists(
    sorted_genres, track_genres, tracks_data, user_id, auth_token, playlist_name
):
    """Create playlists divided by subgenre, prioritizing lower count genres first."""
    used_tracks = set()

    with ThreadPoolExecutor(max_workers=5) as executor:
        for genre, _ in sorted_genres:
            genre_tracks = [
                track_id
                for track_id, genres in track_genres.items()
                if genre in genres and track_id not in used_tracks
            ]
            if not genre_tracks:
                continue

            playlist_id = create_playlist(
                user_id,
                auth_token,
                f"{playlist_name} - {genre}",
                f"Split by subgenre: {genre}. Made using Splitify: https://splitifytool.com/",
            )

            slices = calc_slices(len(genre_tracks))
            for position in range(0, slices * 100, 100):
                if (position + 100) > len(genre_tracks):
                    track_slice = genre_tracks[position:]
                else:
                    track_slice = genre_tracks[position : position + 100]

                track_uris = [tracks_data[track_id]["uri"] for track_id in track_slice]
                executor.submit(add_tracks_to_playlist, playlist_id, track_uris, position, auth_token, playlist_name, genre)

            used_tracks.update(genre_tracks)

def process_playlists(auth_token, playlist_ids):
    """Process multiple playlists by splitting them into subgenre playlists."""
    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = [executor.submit(process_single_playlist, auth_token, playlist_id) for playlist_id in playlist_ids]
        for future in futures:
            future.result()

def process_single_playlist(auth_token, playlist_id):
    """Process a single playlist and divide its tracks into subgenre playlists."""
    playlist_name = get_playlist_name(playlist_id, auth_token)
    user_id = get_user_id(auth_token)

    track_genres = assign_genres_to_tracks(auth_token, playlist_id)

    tracks_data = {
        track_id: {"uri": track_id}
        for track_id in track_genres.keys()
    }

    sorted_genres = sort_genres_by_count(track_genres)

    create_and_populate_subgenre_playlists(
        sorted_genres, track_genres, tracks_data, user_id, auth_token, playlist_name
    )
