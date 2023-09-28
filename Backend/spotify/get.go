package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ResponseData struct {
	Total int `json:"total"`
}
type track struct {
	Spotify_id string `json:"spotify_id"`
	Name       string `json:"name"`
}

type playlist struct {
	Spotify_id string  `json:"spotify_id"`
	Name       string  `json:"name"`
	Items      []track `json:"items"`
}

func Get_playlist_length(playlistID string, authToken string) int {
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=0&limit=100&fields=items(track(name,id,artists(name)))?limit=100", playlistID)
	// Send an HTTP GET request to the Spotify playlist URL
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	// Set the Authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Send the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return 0
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return 0
	}
	// Parse the JSON data into a struct
	var responseData ResponseData
	if err := json.Unmarshal(body, &responseData); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return 0
	}

	// Return the value of the "total" key
	return responseData.Total

}

func Get_playlist_children(index int, playlistID string, authToken string) (string, error) {
	// goes in limits of 100 songs

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
