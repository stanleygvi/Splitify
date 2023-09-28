package main

import (
	"backend/spotify"
	"encoding/json"
	"fmt"
)

type Artist struct {
	Name string `json:"name"`
}

type Track struct {
	SpotifyID string   `json:"id"`
	Name      string   `json:"name"`
	Artists   []Artist `json:"artists"`
}

type PlaylistItem struct {
	Track Track `json:"track"`
}

var playlistData struct {
	Items []PlaylistItem `json:"items"`
}

// calculate how many slices of 100 goes into length
func calcSlices(length int) int {
	if length <= 0 {
		return 0
	}

	return (length + 99) / 100
}

func main() {
	id := "01NTza8NIKI7vBIz1jRJD6"
	authToken := "BQAYOSbvnug0qJV_wqQ98UUPc9EUyxkfQkeQMDMDI6fRnKXHNJGt3AUdSqkBEZhyveZ_G2_kASQv9BKxbaN0VDEaJLY54L-shvpQ1KMSRbcnoigi8r0"
	length := spotify.Get_playlist_length(id, authToken)
	slices := calcSlices(length)

	for i := 0; i < slices; i++ {
		resp, err := spotify.Get_playlist_children(i*100, id, authToken)
		if err != nil {
			fmt.Println(err.Error())
		}

		var playlistItems struct {
			Items []PlaylistItem `json:"items"`
		}

		if err := json.Unmarshal([]byte(resp), &playlistItems); err != nil {
			fmt.Println("Error:", err)
			return
		}

		playlistData.Items = append(playlistData.Items, playlistItems.Items...)

	}

	fmt.Println("Total items in playlistData:", len(playlistData.Items))
	fmt.Println(playlistData.Items[950])
}
