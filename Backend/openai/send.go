package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Message struct {
	Content string `json:"content"`
}

func Send(num_groups string, songs string) string {
	apiKeyBytes, err := os.ReadFile("openai/OPENAI_SECRET.secret")
	if err != nil {
		log.Fatalf("Failed to read API key: %v", err)
	}
	apiKey := string(apiKeyBytes)

	instruction := map[string]interface{}{
		"task":           "Categorize the given songs into separate playlists based on their musical style and content. Provide a unique description and name for each playlist. Use the format below:",
		"requirements":   fmt.Sprintf("MUST: Ensure every track is listed with a valid Spotify track ID under 'tracks_uri' Spotify track IDs are given by each song with the 'id' key. Tracks should be comma-separated. The total number of track_uris across the playlists should be equal to the number of songs Do not include unrelated content. REQUIRED: Create exactly %s playlists. Each song should be represented across these playlists. Each playlist should have a Description, Name, Public status (always true), and a 'tracks_uri' list of songs. Include a count field that verifies how many track_ids are in the playlist", num_groups),
		"num_playlists":  num_groups,
		"example_output": "Playlist 1:\n- Description: Example Description\n- Name: Example Name\n- Public: true\n- tracks_uri: spotify:track:ExampleTrackID1,spotify:track:ExampleTrackID2",
		"songs":          songs,
	}

	messageContent, err := json.Marshal(instruction)
	if err != nil {
		log.Fatal(err)
	}

	message := "Here's your instruction: " + string(messageContent)

	apiUrl := "https://api.openai.com/v1/chat/completions"
	err = os.WriteFile("message.txt", []byte(message), 0644)
	if err != nil {
		log.Fatal(err)
	}

	requestBody := map[string]interface{}{
		"model":       "gpt-3.5-turbo-16k",
		"messages":    []map[string]interface{}{{"role": "user", "content": message}},
		"temperature": 0.6,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(requestJSON))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var response Response
	jsonErr := json.Unmarshal(responseBody, &response)
	if jsonErr != nil {
		return fmt.Sprintf("Error:%s", err)
	}

	content := response.Choices[0].Message.Content
	return content
}
