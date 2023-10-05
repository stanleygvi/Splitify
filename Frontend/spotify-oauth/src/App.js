import React from 'react';
import './App.css';

function App() {
    const authenticateWithSpotify = () => {
        // Implement your Spotify authentication logic here.
        // For simplicity, let's just redirect to Spotify's OAuth page.
        const clientId = "YOUR_SPOTIFY_CLIENT_ID";
        const redirectUri = encodeURIComponent(window.location.href); // Or your specific redirect URL
        window.location.href = `https://accounts.spotify.com/authorize?client_id=${clientId}&response_type=token&redirect_uri=${redirectUri}`;
    };

    return (
        <div className="app">
            <header className="app-header">
            <h1 className="app-title">Splitify</h1>
            <button className="spotify-auth-btn" onClick={authenticateWithSpotify}>
                Authenticate with Spotify
            </button>
        </header>
  
        </div>
    );
}

export default App;
