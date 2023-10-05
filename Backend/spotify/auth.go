package spotify

import (
	// ... other imports ...
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

var clientID string
var clientSecret string

const (
	tokenURL    = "https://accounts.spotify.com/api/token"
	redirectURI = "http://localhost:8080/callback"
)

func init() {
	clientIDBytes, err := os.ReadFile("spotify/CLIENT_ID.secret")
	if err != nil {
		log.Fatalf("Failed to read CLIENT_ID: %v", err)
	}
	clientID = string(clientIDBytes)

	clientSecretBytes, err := os.ReadFile("spotify/CLIENT_SECRET.secret")
	if err != nil {
		log.Fatalf("Failed to read CLIENT_SECRET: %v", err)
	}
	clientSecret = string(clientSecretBytes)
}

// GetToken function: exchanges the authorization code for an access token
func GetToken(code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tokenData := struct {
		AccessToken string `json:"access_token"`
	}{}

	if err := json.Unmarshal(body, &tokenData); err != nil {
		return "", err
	}

	return tokenData.AccessToken, nil
}
