from threading import Thread
from spotify_api import get_playlist_length, get_playlist_children, create_playlist, add_songs, get_user_id
from helpers import calc_slices, convert_to_string

def process_playlists(auth_token, playlist_ids):
    threads = []
    for playlist_id in playlist_ids:
        length = get_playlist_length(playlist_id, auth_token)
        if length == -1:
            print(f"Error fetching playlist length for {playlist_id}")
            continue

        for i in range(0, length, 400):
            thread = Thread(target=process_single_playlist, args=(auth_token, playlist_id, length, i))
            thread.start()
            threads.append(thread)

    for thread in threads:
        thread.join()

def process_single_playlist(auth_token, playlist_id, total_length, start_index):
    slices = calc_slices(total_length)
    playlist_data_store = {'items': []}

    for i in range(start_index, slices * 100, 100):
        append_to_playlist_data(i, playlist_id, auth_token, playlist_data_store)

    user_id = get_user_id(auth_token)

def append_to_playlist_data(start_index, playlist_id, auth_token, data_store):
    response = get_playlist_children(start_index, playlist_id, auth_token)
    if response and 'items' in response:
        data_store['items'].extend(response['items'])
        print(f"Appended {len(response['items'])} items from playlist starting at index {start_index}")
    else:
        print(f"Failed to append playlist data from index {start_index}")

def add_track_ids_to_playlist(gpt_playlists, playlist_items):
    for playlist in gpt_playlists['playlists']:
        track_uris = []
        for index in playlist['song_ids']:
            if index >= len(playlist_items):
                print(f"Index out of range: {index} for playlist items of length {len(playlist_items)}")
                continue
            track_id = playlist_items[index]['track']['id']
            if track_id:
                track_uris.append(f"spotify:track:{track_id}")
        playlist['track_ids'] = ','.join(track_uris)
