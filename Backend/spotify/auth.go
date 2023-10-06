package spotify

import (
	// ... other imports ...
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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
func IsAccessTokenValid(accessToken string) bool {
	// Try to fetch user profile as a test. Adjust as necessary.
	resp, err := http.Get("https://api.spotify.com/v1/me")
	if err != nil {

		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {

		return false
	}

	return true
}
func RefreshAccessToken(refreshToken string) (string, error) {
	endpoint := "https://accounts.spotify.com/api/token"

	clientID, err := os.ReadFile("CLIENT_ID.secret")
	if err != nil {
		return "", err
	}

	clientSecret, err := os.ReadFile("CLIENT_SECRET.secret")
	if err != nil {
		return "", err
	}

	// Base64 encoding the clientID and clientSecret
	auth := base64.StdEncoding.EncodeToString([]byte(string(clientID) + ":" + string(clientSecret)))

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to refresh token: %s", resp.Status)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	newToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("Token not found in response")
	}

	return newToken, nil
}
