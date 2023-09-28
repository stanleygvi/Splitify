package spotify

import (
	"fmt"
	"io"
	"net/http"
)

type track struct {
	Spotify_id string `json:"spotify_id"`
	Name       string `json:"name"`
}

type playlist struct {
	Spotify_id string  `json:"spotify_id"`
	Name       string  `json:"name"`
	Items      []track `json:"items"`
}

func Get_playlist(index int, playlistID string, authToken string) (string, error) {
	// goes in limits of 100 songs

	//url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=100&limit=100fields=items(track(name,id,artists(name)))?limit=100", playlistID)
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=%d&limit=100&fields=items(track(name,id,artists(name)))?limit=100", playlistID, index)
	// Send an HTTP GET request to the Spotify playlist URL
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	// Set the Authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Send the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}

	return string(body), nil
}
