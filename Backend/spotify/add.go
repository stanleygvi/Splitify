package spotify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func Create_playlist(user_id string, auth string) (string, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", user_id)
	payload := []byte(`{
		"name": "New Playlist",
		"description": "New playlist description",
		"public": true
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Failed to create playlist, status code: %s", resp.Status)
	}

	// Parse the JSON response
	var responseData map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&responseData); err != nil {
		fmt.Println("Error decoding response:", err)
		return "", err
	}

	// Extract the "id" field from the response
	id, ok := responseData["id"].(string)
	if !ok {
		fmt.Println("Error extracting playlist ID from response")
		return "", errors.New("Failed to extract playlist ID from response")
	}

	return id, nil
}

func Add_songs(playlist_id string, songs string, auth string) {

	// URL for the POST request
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlist_id)

	// Request payload data
	payload := []byte(`{
		"uris": [
			` + songs + `
		],
		"position": 0
	}`)

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.Status == "200 OK" {
		fmt.Println("Request was successful!")
	} else {
		fmt.Println("Request failed with status:", resp.Status)
	}
}
