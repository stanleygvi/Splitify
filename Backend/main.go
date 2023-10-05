package main

import (
	"backend/openai"
	"backend/spotify"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type Artist struct {
	Name string `json:"name"`
}

type Track struct {
	SpotifyID string   `json:"id"`
	Name      string   `json:"name"`
	Artists   []Artist `json:"artists"`
}

type PlaylistItem struct {
	Track Track `json:"track"`
}

type Playlist struct {
	Spotify_ID  string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Items       []Track `json:"items"`
}

type GPT_Playlists struct {
	Playlists []Playlist
}

type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var playlistData struct {
	Items []PlaylistItem `json:"items"`
}

func convertToString(index int, track Track) string {
	var artists []string
	for _, artist := range track.Artists {
		artists = append(artists, artist.Name)
	}
	artistsStr := strings.Join(artists, ", ")
	return fmt.Sprintf("{%d: %s,%s},", index, track.Name, artistsStr)
}

// calculate how many slices of 100 goes into length
func calcSlices(length int) int {
	if length <= 0 {
		return 0
	}

	return (length + 99) / 100
}

func append_to_playlistData(startIndex int, id, authToken string, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := spotify.Get_playlist_children(startIndex, id, authToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var playlistItems struct {
		Items []PlaylistItem `json:"items"`
	}

	if err := json.Unmarshal([]byte(resp), &playlistItems); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Append the playlistItems.Items to playlistData.Items
	playlistData.Items = append(playlistData.Items, playlistItems.Items...)
	fmt.Printf("Appended %d items to playlistData from startIndex %d\n", len(playlistItems.Items), startIndex)
}

func add_playlist_to_spotify(user_id string, songs string, auth string, wg *sync.WaitGroup) {
	defer wg.Done()

	playlist_id, err := spotify.Create_playlist(user_id, auth)
	if err != nil {
		return
	}
	spotify.Add_songs(playlist_id, songs, auth)

}

func ExtractJSONFromContent(content string) (string, error) {
	startIndex := strings.Index(content, "{")
	if startIndex == -1 {
		return "", fmt.Errorf("no valid JSON object found")
	}
	return content[startIndex:], nil
}

func AddPlaylistsFromResponse(resp string, gptPlaylists *GPT_Playlists) {
	var response Response
	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		fmt.Println("Error unmarshalling main response:", err)
		return
	}

	for _, choice := range response.Choices {
		extractedJSON, err := ExtractJSONFromContent(choice.Message.Content)
		if err != nil {
			fmt.Println("Error extracting valid JSON:", err)
			continue
		}

		var playlistData map[string]interface{}
		if err := json.Unmarshal([]byte(extractedJSON), &playlistData); err != nil {
			fmt.Println("Error unmarshalling inner JSON:", err)
			continue
		}

		playlistInfo := playlistData["playlist"].(map[string]interface{})
		uriStrings := playlistData["uri_string"].(string)

		trackURIs := strings.Split(uriStrings, ",")
		var tracks []Track
		for _, uri := range trackURIs {
			tracks = append(tracks, Track{SpotifyID: strings.TrimPrefix(uri, "spotify:track:")})
		}

		playlist := Playlist{
			Name:        playlistInfo["name"].(string),
			Description: playlistInfo["description"].(string),
			Items:       tracks,
		}
		gptPlaylists.Playlists = append(gptPlaylists.Playlists, playlist)
	}
}
func main() {
	playlist_id := "2vCOpYeOF6cgvCEwacU3bC"
	// user_id := "user_id"
	authToken, err := os.ReadFile("spotify/TOKEN.secret")
	if err != nil {
		log.Fatalf("Failed to read API key: %v", err)
	}
	authTokenstr := string(authToken)
	length := spotify.Get_playlist_length(playlist_id, authTokenstr)
	slices := calcSlices(length)

	var wg sync.WaitGroup

	for i := 0; i < slices; i++ {
		wg.Add(1)
		startIndex := i * 100
		go append_to_playlistData(startIndex, playlist_id, authTokenstr, &wg)
	}

	wg.Wait()
	songs := ""

	for index, obj := range playlistData.Items {
		objStr := convertToString(index, obj.Track)
		songs += objStr
	}

	gpt_resp := openai.Send("5", songs)
	fmt.Println(gpt_resp)
	// var gptPlaylists GPT_Playlists
	// AddPlaylistsFromResponse(gpt_resp, &gptPlaylists)
	// fmt.Println(gptPlaylists.Playlists)

}
