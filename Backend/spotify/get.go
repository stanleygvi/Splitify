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
	Items []struct {
		Track struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"track"`
	} `json:"items"`
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

func Get_playlist_length(playlistID string, authToken string) int {
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=0&limit=100&fields=items(track(name,id,artists(name)))", playlistID)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return -1
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return -1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return -1
	}

	var responseData PlaylistResponse
	if err := json.Unmarshal(body, &responseData); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return -1
	}

	return len(responseData.Items)
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
