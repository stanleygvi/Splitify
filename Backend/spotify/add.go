package spotify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func Create_playlist(user_id string, auth string, name string, description string) (string, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", user_id)
	payload := []byte(fmt.Sprintf(`{
		"name": "%s",
		"description": " %s - created using https://splitify-fac76.web.app/",
		"public": true
	}`, name, description))

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

func Add_songs(playlist_id string, songs []string, auth string, index int) {
	// Define the API endpoint
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlist_id)

	// Create a struct for our request body
	type requestBody struct {
		URIs     []string `json:"uris"`
		Position int      `json:"position"`
	}

	// Fill the struct with our data
	bodyData := requestBody{
		URIs:     songs,
		Position: index,
	}

	// Convert our struct to JSON
	jsonData, err := json.Marshal(bodyData)
	if err != nil {
		fmt.Println("Error marshalling the JSON:", err)
		return
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	defer resp.Body.Close()
	// Read the response body
	if resp.StatusCode != 201 {
		fmt.Println(resp.Status)
	}

}
