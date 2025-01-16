from threading import Thread, Lock
from concurrent.futures import ThreadPoolExecutor
from collections import Counter
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
import time

def assign_genres_to_tracks(auth_token, playlist_id):
    """Assign genres to each track and return a mapping of track_id to genres."""
    slices = calc_slices(get_playlist_length(playlist_id, auth_token))
    track_genres = {}
    genre_lock = Lock()

    for i in range(0, slices * 100, 100):
        response = get_playlist_children(i, playlist_id, auth_token)
        if response and "items" in response:
            tracks = response["items"]
            track_to_artists = {
                track["track"]["id"]: [artist["id"] for artist in track["track"]["artists"]]
                for track in tracks
                if track["track"] and "artists" in track["track"]
            }
            print(track_to_artists)
            with ThreadPoolExecutor(max_workers=10) as executor:
                futures = [
                    executor.submit(
                        fetch_genres,
                        artist_ids,
                        track_id,
                        auth_token,
                        track_genres,
                        genre_lock,
                    )
                    for track_id, artist_ids in track_to_artists.items()
                ]
                for future in futures:
                    try:
                        future.result()
                    except Exception as e:
                        print(f"Error fetching genres: {e}")
    return track_genres

def fetch_genres(artist_ids, track_id, auth_token, track_genres, genre_lock):
    """Fetch genres for a given list of artist IDs and assign them to a track."""
    artist_genres = get_artists(artist_ids, auth_token)
    genres = set()
    for artist_id in artist_ids:
        genres.update(artist_genres[artist_id])
    with genre_lock:
        track_genres[track_id] = list(genres)

def sort_genres_by_count(track_genres):
    """Return a sorted list of genres by frequency in ascending order."""
    genre_counter = Counter(genre for genres in track_genres.values() for genre in genres)
    return sorted(genre_counter.items(), key=lambda x: x[1])

def create_and_populate_subgenre_playlists(
    sorted_genres, track_genres, tracks_data, user_id, auth_token, playlist_name
):
    """Create playlists divided by subgenre, prioritizing lower count genres first."""
    used_tracks = set()

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
            status = add_songs(playlist_id, track_uris, auth_token, position)
            time.sleep(0.5)

            if not status or status.get("Error", None):
                print(
                    f"Append Error: Playlist {playlist_name} - {genre}, status {status}, starting from index: {position}"
                )

        used_tracks.update(genre_tracks)

def process_playlists(auth_token, playlist_ids):
    """Process multiple playlists by splitting them into subgenre playlists."""
    for playlist_id in playlist_ids:
        process_single_playlist(auth_token, playlist_id)

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
