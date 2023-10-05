package main

import (
	"backend/openai"
	"backend/spotify"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// Your data structures
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

// Your helper functions
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

func addTrackIDsToPlaylist(gptPlaylists *GPT_Playlists, playlistItems []PlaylistItem) {
	for i := range gptPlaylists.Playlists {
		track_ids := ""
		for _, index := range gptPlaylists.Playlists[i].Indexes {
			track_ids += fmt.Sprintf("spotify:%s,", playlistItems[index].Track.SpotifyID)
		}
		gptPlaylists.Playlists[i].Track_ids = track_ids
	}
}

func generateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = letters[rand.Intn(len(letters))]
	}
	return string(bytes)
}

const (
	redirectURI = "http://localhost:8888/callback"
	authURL     = "https://accounts.spotify.com/authorize"
)

// HTTP Handlers
func loginHandler(w http.ResponseWriter, r *http.Request) {
	clientID, err := os.ReadFile("spotify/CLIENT_ID.secret")
	if err != nil {
		return
	}
	state := generateRandomString(16)
	scope := "user-read-private user-read-email playlist-modify-public playlist-modify-private"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", string(clientID))
	params.Add("scope", scope)
	params.Add("redirect_uri", redirectURI)
	params.Add("state", state)

	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

var authStr string

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	authToken, err := spotify.GetToken(code)
	if err != nil {
		http.Error(w, "Failed to get auth token", http.StatusInternalServerError)
		return
	}

	// TODO: Save this authToken for further requests (in-memory for simplicity, but use DB for production).
	w.Write([]byte("Authentication successful! You can now process the playlist."))

	os.WriteFile("TOKEN.secret", []byte(authToken), 0600)
}

func processPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	// For the example, I'll assume you're fetching the authToken directly.
	// In practice, you'd want to retrieve this from where you stored it after the /callback
	authToken, err := os.ReadFile("spotify/TOKEN.secret") // Use whatever method you use to retrieve the token here.
	if err != nil {
		http.Error(w, "Failed to get auth token", http.StatusInternalServerError)
		return
	}
	authStr := string(authToken)
	playlist_id := "2vCOpYeOF6cgvCEwacU3bC" // You might want this to be dynamic in the future
	length := spotify.Get_playlist_length(playlist_id, authStr)
	slices := calcSlices(length)

	var wg sync.WaitGroup

	for i := 0; i < slices; i++ {
		wg.Add(1)
		startIndex := i * 100
		go append_to_playlistData(startIndex, playlist_id, authStr, &wg)
	}

	wg.Wait()
	songs := ""

	for index, obj := range playlistData.Items {
		objStr := convertToString(index, obj.Track)
		songs += objStr
	}

	gpt_resp := openai.Send("5", songs)
	var gptPlaylists GPT_Playlists
	if err := json.Unmarshal([]byte(gpt_resp), &gptPlaylists); err != nil {
		http.Error(w, "Error unmarshalling main response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	addTrackIDsToPlaylist(&gptPlaylists, playlistData.Items)
	for _, playlist := range gptPlaylists.Playlists {
		spotify.Add_songs(playlist.Spotify_ID, playlist.Track_ids, authStr)
	}

	w.Write([]byte("Playlists updated successfully!"))
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/process-playlist", processPlaylistHandler)

	log.Fatal(http.ListenAndServe(":8888", nil))
}
