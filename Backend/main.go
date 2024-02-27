package main

import (
	"backend/openai"
	"backend/spotify"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/rs/cors"
)

// Data structures
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

type PlaylistDataStore struct {
	Items []PlaylistItem `json:"items"`
}

type PlaylistIDS struct {
	PlaylistIDS []string `json:"playlistIds"`
}

// Helper functions
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

func append_to_playlistData(startIndex int, id, authToken string, wg *sync.WaitGroup, dataStore *PlaylistDataStore) {
	defer wg.Done()

	resp, err := spotify.Get_playlist_children(startIndex, id, authToken)
	fmt.Println("CHILDREN:\n", resp)
	if err != nil {
		log.Println("Error fetching playlist children:", err.Error())
		return
	}

	var playlistItems struct {
		Items []PlaylistItem `json:"items"`
	}

	if err := json.Unmarshal([]byte(resp), &playlistItems); err != nil {
		log.Println("Error unmarshalling playlist items:", err)
		return
	}
	log.Println("\nPlayList Items:\n", playlistItems)
	dataStore.Items = append(dataStore.Items, playlistItems.Items...)
	log.Printf("Appended %d items to playlistData from startIndex %d\n", len(playlistItems.Items), startIndex)
}

func addTrackIDsToPlaylist(gptPlaylists *GPT_Playlists, playlistItems []PlaylistItem) {
	for i := range gptPlaylists.Playlists {
		var trackUris []string
		for _, index := range gptPlaylists.Playlists[i].Indexes {
			if index >= len(playlistItems) {
				log.Printf("Index out of range: %d for playlistItems of length %d", index, len(playlistItems))
				continue
			}
			spotifyId := playlistItems[index].Track.SpotifyID
			if len(spotifyId) > 5 && spotifyId != "" {
				trackUris = append(trackUris, fmt.Sprintf("spotify:track:%s", spotifyId))
			}
		}
		gptPlaylists.Playlists[i].Track_ids = strings.Join(trackUris, ",")
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
	redirectURI = "http://localhost:8080/callback"
	authURL     = "https://accounts.spotify.com/authorize"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Attempt to retrieve the user's access token directly from the environment variable.
	authToken := os.Getenv("TOKEN")
	if !spotify.IsAccessTokenValid(authToken) {
		// If the access token is not valid or doesn't exist, try to refresh it.
		refreshToken := os.Getenv("REFRESH_TOKEN")
		if refreshToken != "" {
			newAccessToken, err := spotify.RefreshAccessToken(refreshToken)
			if err != nil || newAccessToken == "" {
				// If refreshing the token fails, redirect to Spotify login to re-authenticate.
				redirectToSpotifyLogin(w, r)
				return
			}

			// Update the access token in the environment variable.
			os.Setenv("TOKEN", newAccessToken)
		} else {
			// No refresh token available, redirect to Spotify login.
			redirectToSpotifyLogin(w, r)
			return
		}
	}
}

func redirectToSpotifyLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("CLIENT_ID")
	state := generateRandomString(16)
	scope := "user-read-private user-read-email playlist-modify-public playlist-modify-private playlist-read-collaborative"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", clientID)
	params.Add("scope", scope)
	params.Add("show_dialog", "true")
	params.Add("redirect_uri", redirectURI)
	params.Add("state", state)

	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code present in callback", http.StatusBadRequest)
		return
	}

	token, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %s", err)
		http.Error(w, "Error exchanging code for token", http.StatusInternalServerError)
		return
	}
	os.Setenv("TOKEN", token)
	http.Redirect(w, r, "https://splitify-fac76.web.app/input-playlist", http.StatusFound)
}

func exchangeCodeForToken(code string) (string, error) {
	clientID := os.Getenv("CLIENT_ID")

	clientSecret := os.Getenv("CLIENT_SECRET")

	endpoint := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", string(clientID))
	data.Set("client_secret", string(clientSecret))
	// fmt.Println(data)
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
	// Retrieve the access token directly from the environment variable.
	authToken := os.Getenv("TOKEN")
	// fmt.Println(authToken)
	if !spotify.IsAccessTokenValid(authToken) {
		// This means the access token is not valid or not found.
		http.Error(w, "Authorization required: no access token available", http.StatusUnauthorized)
		return
	}

	playlists := spotify.Get_all_playlists(authToken)

	jsonData, err := json.Marshal(playlists)
	if err != nil {
		http.Error(w, "Failed to generate JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func processPlaylist(authToken string, playlist_id string, wg *sync.WaitGroup, dataStore *PlaylistDataStore) {
	defer wg.Done()
	length := spotify.Get_playlist_length(playlist_id, authToken)
	fmt.Println(length)
	if length == -1 {
		log.Println("Error fetching playlist length for:", playlist_id)
		return
	}

	slices := calcSlices(length)

	var wg_append sync.WaitGroup

	for i := 0; i < slices && i < 5; i++ {
		wg_append.Add(1)
		startIndex := i * 100

		go append_to_playlistData(startIndex, playlist_id, authToken, &wg_append, dataStore)
	}

	wg_append.Wait()
	songs := ""

	for index, obj := range dataStore.Items {
		objStr := convertToString(index, obj.Track)
		songs += objStr
	}

	gpt_resp := openai.Send("5", songs)
	var gptPlaylists GPT_Playlists
	// Find the start and end indices of the actual JSON content.
	startIndex := strings.Index(gpt_resp, "{")
	endIndex := strings.LastIndex(gpt_resp, "}") + 1 // +1 to include the closing brace.
	// Extract the JSON content.
	if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
		jsonContent := gpt_resp[startIndex:endIndex]

		// Unmarshal the JSON content into the struct.
		err := json.Unmarshal([]byte(jsonContent), &gptPlaylists)
		if err != nil {
			log.Printf("Error unmarshalling GPT response: %v", err)
			return
		}

	} else {
		log.Println("Valid JSON content not found in the response")
	}
	log.Println("Finished unmarshalling: \n", gptPlaylists)
	addTrackIDsToPlaylist(&gptPlaylists, dataStore.Items)
	// log.Println("added tracks to playlists:\n", res)
	user_id := spotify.Get_user_id(string(authToken))
	for _, playlist := range gptPlaylists.Playlists {
		playlist_id, err := spotify.Create_playlist(user_id, string(authToken), playlist.Name, playlist.Description)
		log.Println("Playlist ID: ", playlist_id)
		if err != nil {
			log.Println("Error creating playlist")
			return
		}

		songs := strings.Split(playlist.Track_ids, ",")

		for i := 0; i < len(songs); i += 100 {
			end := i + 100
			if end > len(songs) {
				end = len(songs)
			}

			spotify.Add_songs(playlist_id, songs[i:end], authToken, i)

		}

	}
}

func processPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	authToken := os.Getenv("TOKEN")
	if !spotify.IsAccessTokenValid(authToken) {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var playlistIDS PlaylistIDS
	if err := json.Unmarshal(body, &playlistIDS); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	for _, id := range playlistIDS.PlaylistIDS {
		wg.Add(1)
		var playlistDataStore PlaylistDataStore
		go processPlaylist(authToken, id, &wg, &playlistDataStore)
	}

	wg.Wait()

	w.Write([]byte("Playlists updated successfully!"))
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/callback", callbackHandler)
	mux.HandleFunc("/process-playlist", processPlaylistHandler)
	mux.HandleFunc("/user-playlists", getPlaylistHandler)
	handler := cors.Default().Handler(mux)
	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}

}
