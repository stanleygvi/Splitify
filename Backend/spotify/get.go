package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type User struct {
	ID string `json:"id"`
}
type PlaylistResponse struct {
	Items []PlaylistItem `json:"items"`
	Total int            `json:"total"`
}

type PlaylistItem struct {
	Track struct {
		ID string `json:"id"`
	} `json:"track"`
}

type Image struct {
	Url string `json:"url"`
}

type PlaylistInfo struct {
	Name   string  `json:"name"`
	ID     string  `json:"id"`
	Images []Image `json:"images"`
}
type AllPlaylists struct {
	Items []PlaylistInfo `json:"items"`
}

func Get_user_id(authToken string) string {

	url := "https://api.spotify.com/v1/me"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)

	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)

	}
	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		fmt.Println("Error parsing JSON:", err)

	}
	//fmt.Println("User ID response:", string(body))

	return user.ID

}

func Get_all_playlists(authToken string) AllPlaylists {
	// fmt.Println("on get playlist")
	id := Get_user_id(authToken)
	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", id)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	// fmt.Println(resp)
	if err != nil {
		fmt.Println("Error sending request:", err)

	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)

	}

	var all_playlists AllPlaylists
	if err := json.Unmarshal(body, &all_playlists); err != nil {
		fmt.Println("Error parsing JSON:", err)

	}
	// fmt.Println(all_playlists)
	return all_playlists

}

func Get_playlist_length(playlistID, authToken string) int {
	totalTracks := 0
	offset := 0
	limit := 100 // Max limit as per Spotify API

	for {
		url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=%d&limit=%d&fields=total,items(track(id))", playlistID, offset, limit)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v", err)
			return -1
		}

		req.Header.Set("Authorization", "Bearer "+authToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v", err)
			return -1
		}

		var response PlaylistResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Printf("Error decoding response: %v", err)
			resp.Body.Close()
			return -1
		}
		resp.Body.Close()

		totalTracks += len(response.Items)
		if totalTracks >= response.Total {
			break
		}
		offset += limit
	}

	return totalTracks
}

func Get_playlist_children(index int, playlistID string, authToken string) (string, error) {

	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=%d&limit=100&fields=items(track(name,id,artists(name)))", playlistID, index)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	fmt.Println("CHILDREN: \n", resp)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}

	return string(body), nil
}
