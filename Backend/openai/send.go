package openai

import (
	"bytes"
	"io"
	"net/http"
)

func Send() string {
	// Replace with your OpenAI API key
	apiKey := "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"

	// Define the API endpoint URL
	apiUrl := "https://api.openai.com/v1/chat/completions"

	// Define the request payload as a JSON string
	requestBody := `{
        "model": "gpt-3.5-turbo",
        "messages": [{"role": "user", "content": "Say this is a test!"}],
        "temperature": 0.7
    }`

	// Create a new HTTP request
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		panic(err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Create an HTTP client
	client := &http.Client{}

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Print the response
	return string(responseBody)
}
