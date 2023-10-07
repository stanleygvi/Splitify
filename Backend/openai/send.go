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
		"task":           fmt.Sprintf("Categorize the given songs into separate playlists based on their musical style and content, not just by artist. Provide a unique description and name for each playlist. REQUIRED: Make sure that every provided song is added to a playlist. Don't include any unrelated data in the response. Use the json format below:"),
		"requirements":   "Each song should be represented across these playlists. Each playlist should have a Description, Name, Public status (always true), and a list of ALL songs included in the playlist. Each song should be recorded as it's provided id. There should be no duplicate songs. OUTPUT MUST BE IN JSON FORMAT. DO NOT TYPE ANYTHING OUTSIDE OF THE JSON.",
		"num_playlists":  num_groups,
		"example_output": "{playlists:[{'name': 'Example Name', 'description': 'Example Description', 'public': true, song_ids:[ 1, 47, 216]}]}",
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
	fmt.Println(string(responseBody))
	var response Response
	jsonErr := json.Unmarshal(responseBody, &response)
	if jsonErr != nil {
		return fmt.Sprintf("Error:%s", err)
	}

	content := response.Choices[0].Message.Content
	// content := string(responseBody)
	return content
}
