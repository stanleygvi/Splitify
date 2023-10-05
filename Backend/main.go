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
	Spotify_ID  string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Indexes     []int  `json:"song_ids"`
	Track_ids   string
}

type GPT_Playlists struct {
	Playlists []Playlist
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
func addTrackIDsToPlaylist(gptPlaylists *GPT_Playlists, playlistItems []PlaylistItem) {
	for i := range gptPlaylists.Playlists {
		track_ids := ""
		for _, index := range gptPlaylists.Playlists[i].Indexes {
			track_ids += fmt.Sprintf("spotify:%s,", playlistItems[index].Track.SpotifyID)
			//gptPlaylists.Playlists[i].Track_ids = append(gptPlaylists.Playlists[i].Track_ids, playlistItems[index].Track.SpotifyID)
		}
		gptPlaylists.Playlists[i].Track_ids = track_ids
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
	var gptPlaylists GPT_Playlists
	if err := json.Unmarshal([]byte(gpt_resp), &gptPlaylists); err != nil {
		fmt.Println("Error unmarshalling main response:", err)
		return
	}
	// function to add the track ids of each index to the 'Track_ids' value of the playlist
	// fmt.Println(gptPlaylists.Playlists)

	addTrackIDsToPlaylist(&gptPlaylists, playlistData.Items)
	fmt.Println(gptPlaylists.Playlists[0].Track_ids)
	for _, playlist := range gptPlaylists.Playlists {
		spotify.Add_songs(playlist.Spotify_ID, playlist.Track_ids, authTokenstr)
	}
}
