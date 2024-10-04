from sklearn.cluster import KMeans
import pandas as pd
import json


def calc_clusters():
    return 3


def scale_audio_features(
    dataframe: pd.DataFrame, keys: list[str], weights: list[float]
):
    assert len(keys) == len(weights)
    for index in range(0, len(weights)):
        dataframe[keys[index]] *= weights[index]


def cluster_df(playlist_data):
    data = pd.DataFrame(playlist_data["tracks"])
    uri = pd.Series(data["uri"])
    training_data = data.drop("uri", axis=1)
    # scale_audio_features(training_data, ["time_signature","tempo", "loudness", "key", "valence", "danceability", "energy", "mode"], [0.1, 0.01, 1/90,0.05, 1.5, 1.2, 2.0, 0.6])
    with open("training.json", "w") as outfile:
        json.dump(playlist_data, outfile)

    kmeans = KMeans(n_clusters=calc_clusters(), random_state=0, n_init="auto")
    kmeans.fit(training_data)
    clusters = kmeans.labels_
    training_data["cluster"] = clusters
    training_data["uri"] = uri
    return training_data
