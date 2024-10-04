from threading import Thread
from Backend.spotify_api import get_playlist_length, get_playlist_children, create_playlist, add_songs, get_user_id, get_audio_features, get_playlist_name
from Backend.helpers import calc_slices
from Backend.grouping import cluster_df
import time

def extract_ids(playlist_data):
    track_ids = []
    for track in playlist_data:
        if track["track"] and track["track"]["id"]:
            track_ids.append(track["track"]["id"])
    return track_ids

def clean_audio_features(audio_features:list[dict[str,float] ],remove_keys:list[str]):
    index = 0
    index_remove = []
    for feature in audio_features:
        if feature:
            for key in remove_keys:
                feature.pop(key, None)
        else:
            index_remove.append(index)
        index+=1
    for i in index_remove:
        audio_features.pop(i)



def process_playlists(auth_token, playlist_ids):
    threads = []
    for playlist_id in playlist_ids:
        length = get_playlist_length(playlist_id, auth_token)
        if length == -1:
            print(f"Error fetching playlist length for {playlist_id}")
            continue

        thread = Thread(target=process_single_playlist, args=(auth_token, playlist_id, length))
        thread.start()
        threads.append(thread)

    for thread in threads:
        thread.join()

def process_single_playlist(auth_token, playlist_id, total_length):
    name = get_playlist_name(playlist_id, auth_token)
    slices = calc_slices(total_length)
    playlist_data_store = {"id":playlist_id,"tracks": []}

    for i in range(0, slices * 100, 100):
        append_to_playlist_data(i, playlist_id, auth_token, playlist_data_store)
    if len(playlist_data_store["tracks"]) < 1:
        print(f"failed to process playlist: {playlist_id}")
        return
    user_id = get_user_id(auth_token)
    grouped = cluster_df(playlist_data_store)
    num_playlists = len(grouped["cluster"].value_counts())

    threads = []
    for num in range(0,num_playlists):
        cluster = grouped[grouped["cluster"] == num]
        thread = Thread(target=created_and_populate, args=(cluster, user_id, auth_token, name))
        thread.start()
        threads.append(thread)
    for thread in threads:
        thread.join()
        time.sleep(1)

def created_and_populate(cluster_df, user_id, auth_token, name):

    slices = calc_slices(len(cluster_df))
    if slices < 1:
        return
    
    playlist_id = create_playlist(user_id, auth_token, f"Split playlist from {name} ", "Made using Splitify: https://splitifytool.com/")
    for position in range(0, slices * 100, 100):
        if (position + 100) > len(cluster_df):
            cluster_slice = cluster_df.iloc[position:]
        else:
            cluster_slice = cluster_df.iloc[position:position+100]
        track_uris = cluster_slice["uri"].tolist()
        
        status = add_songs(playlist_id, track_uris, auth_token, position)
        print(f"{name} split append status {status} starting from index: {position}")

def append_to_playlist_data(start_index, playlist_id, auth_token, data_store):
    response = get_playlist_children(start_index, playlist_id, auth_token)
    if response and "items" in response:
        
        track_ids = extract_ids(response["items"])
        audio_features = get_audio_features(track_ids, auth_token)
        clean_audio_features(audio_features, ["type", "id", "track_href", "analysis_url", "duration_ms"])
        data_store["tracks"].extend(audio_features)
        print(f"Appended {len(response["items"])} tracks from playlist starting at index {start_index}")
    else:
        print(f"Failed to append playlist data from index {start_index}")

