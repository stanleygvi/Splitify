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
	apiKeyBytes := os.Getenv("OPENAI_SECRET")

	apiKey := string(apiKeyBytes)

	instruction := map[string]interface{}{
		"task":            fmt.Sprintf("Categorize the given songs into %s playlists based on their musical style and content, not just by artist. Make sure each playlist is unique from the others. Provide a unique description and name for each playlist. REQUIRED: Make sure that every provided song is added to a playlist.", num_groups),
		"requirements":    "Each song should be represented across these playlists. Each playlist should have a Description, Name, Public status (always true), and a list of ALL songs included in the playlist. Each song should be recorded as its provided id. There should be no duplicate songs. OUTPUT MUST BE IN JSON FORMAT. DO NOT TYPE ANYTHING OUTSIDE OF THE JSON.",
		"num_playlists":   num_groups,
		"response_format": "JSON",
		"example_output": `{
			"playlists": [
				{
					"id": "playlist1ID",
					"name": "Example Name",
					"description": "Example Description",
					"song_ids": [1, 47, 216]
				},
				{
					"id": "playlist2ID",
					"name": "Example Name 2",
					"description": "Example Description 2",
					"song_ids": [2, 48, 100, 5, 43, 99]
				}
			]
		}`,
		"songs": songs,
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
		"model":       "gpt-4-1106-preview",
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
	fmt.Println("\n\nRESP:\n", responseBody)

	var response Response
	jsonErr := json.Unmarshal(responseBody, &response)
	if jsonErr != nil {
		return fmt.Sprintf("Error:%s", err)
	}

	content := response.Choices[0].Message.Content

	return string(content)
}
