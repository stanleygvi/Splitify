package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func GetAccessToken(code string) (string, error) {
	// Construct the token request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost:8888/callback") // Your Redirect URI

	// Make a POST request to Spotify's token endpoint
	resp, err := http.Post("https://accounts.spotify.com/api/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the response to get the access token
	var tokenResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("Access token not found in response")
	}
	fmt.Println(accessToken)
	return accessToken, nil
}
