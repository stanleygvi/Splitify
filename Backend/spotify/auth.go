package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var clientID = os.Getenv("CLIENT_ID")
var clientSecret = os.Getenv("CLIENT_SECRET")

const (
	tokenURL    = "https://accounts.spotify.com/api/token"
	redirectURI = "http://localhost:8080/callback"
)

func GetToken(code string) (string, error) {
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("client ID or client secret is not set")
	}

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

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}

	// Set the Authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return false
	}
	defer resp.Body.Close()
	// fmt.Println(resp)

	// If the status code is 401, the token is invalid or expired
	return resp.StatusCode != 401
}

func RefreshAccessToken(refreshToken string) (string, error) {
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("client ID or client secret is not set")
	}

	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
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
		return "", fmt.Errorf("failed to refresh token: %s", resp.Status)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	newToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return newToken, nil
}
