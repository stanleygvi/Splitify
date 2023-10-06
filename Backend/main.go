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

	"github.com/rs/cors"
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
func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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

	authToken, err := os.ReadFile("spotify/TOKEN.secret") // Use whatever method you use to retrieve the token here.
	if err != nil || !spotify.IsAccessTokenValid(string(authToken)) {
		redirectToSpotifyLogin(w, r)
		return
	}

	// If the token is not valid, try to use the refresh token
	if !spotify.IsAccessTokenValid(string(authToken)) {
		refreshToken, err := os.ReadFile("spotify/REFRESH_TOKEN.secret")
		if err != nil {
			// If there's an error reading the refresh token, it probably doesn't exist. Redirect to re-login.
			redirectToSpotifyLogin(w, r)
			return
		}

		// Use the refresh token to get a new access token
		newAccessToken, err := spotify.RefreshAccessToken(string(refreshToken))
		if err != nil {
			http.Error(w, "Failed to refresh access token", http.StatusInternalServerError)
			return
		}

		// Store the new access token for future use
		os.WriteFile("spotify/TOKEN.secret", []byte(newAccessToken), 0600)
	}

	redirectToSpotifyLogin(w, r)
}

func redirectToSpotifyLogin(w http.ResponseWriter, r *http.Request) {
	clientID, err := os.ReadFile("spotify/CLIENT_ID.secret")
	if err != nil {
		http.Error(w, "Failed to read client ID", http.StatusInternalServerError)
		return
	}

	state := generateRandomString(16)
	scope := "user-read-private user-read-email playlist-modify-public playlist-modify-private playlist-read-collaborative"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", string(clientID))
	params.Add("scope", scope)
	params.Add("show_dialog", "true")
	params.Add("redirect_uri", redirectURI)
	params.Add("state", state)

	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

var authStr string

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code present in callback", http.StatusBadRequest)
		return
	}

	token, err := exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Error exchanging code for token", http.StatusInternalServerError)
		return
	}

	// Save token for future use
	err = os.WriteFile("spotify/TOKEN.secret", []byte(token), 0600)
	if err != nil {
		http.Error(w, "Error saving token", http.StatusInternalServerError)
		return
	}

	// Redirect to the desired page
	http.Redirect(w, r, "https://splitify-fac76.web.app/input-playlist", http.StatusFound)
}

func exchangeCodeForToken(code string) (string, error) {
	clientID, err := os.ReadFile("spotify/CLIENT_ID.secret")
	if err != nil {
		return "", err
	}

	clientSecret, err := os.ReadFile("spotify/CLIENT_SECRET.secret")
	if err != nil {
		return "", err
	}

	endpoint := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", string(clientID))
	data.Set("client_secret", string(clientSecret))

	client := &http.Client{}
	r, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func getPlaylistHandler(w http.ResponseWriter, r *http.Request) {

	authToken, err := os.ReadFile("spotify/TOKEN.secret") // Use whatever method you use to retrieve the token here.
	if err != nil {
		http.Error(w, "Failed to get auth token", http.StatusInternalServerError)
		return
	}
	playlists := spotify.Get_all_playlists(string(authToken))
	fmt.Println("Playlists fetched:", playlists)
	jsonData, err := json.Marshal(playlists)
	if err != nil {
		http.Error(w, "Failed to generate JSON", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
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
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/callback", callbackHandler)
	mux.HandleFunc("/process-playlist", processPlaylistHandler)
	mux.HandleFunc("/user-playlists", getPlaylistHandler)
	handler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(":8888", handler))
}
