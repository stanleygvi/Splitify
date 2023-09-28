package main

import (
	"backend/spotify"
	"encoding/json"
	"fmt"
	"sync"
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

func append_to_playlistData(startIndex int, id, authToken string, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := spotify.Get_playlist_children(startIndex, id, authToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var playlistItems struct {
		Items []PlaylistItem `json:"items"`
	}

	if err := json.Unmarshal([]byte(resp), &playlistItems); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Append the playlistItems.Items to playlistData.Items
	playlistData.Items = append(playlistData.Items, playlistItems.Items...)
	fmt.Printf("Appended %d items to playlistData from startIndex %d\n", len(playlistItems.Items), startIndex)
}

func main() {
	id := "01NTza8NIKI7vBIz1jRJD6"
	authToken := "BQAYOSbvnug0qJV_wqQ98UUPc9EUyxkfQkeQMDMDI6fRnKXHNJGt3AUdSqkBEZhyveZ_G2_kASQv9BKxbaN0VDEaJLY54L-shvpQ1KMSRbcnoigi8r0"
	length := spotify.Get_playlist_length(id, authToken)
	slices := calcSlices(length)

	var wg sync.WaitGroup

	for i := 0; i < slices; i++ {
		wg.Add(1)
		startIndex := i * 100
		go append_to_playlistData(startIndex, id, authToken, &wg)
	}

	wg.Wait()

	fmt.Printf("Total items in playlistData: %d\n", len(playlistData.Items))
	fmt.Println(playlistData.Items[0])
}
