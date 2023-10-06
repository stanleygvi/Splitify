// PlaylistInputPage.js
import React, { useState } from 'react';

function PlaylistInputPage() {
    const [playlistUrl, setPlaylistUrl] = useState('');

    const handleInputChange = (e) => {
        setPlaylistUrl(e.target.value);
    };

    const handleSubmit = () => {
        // Process the playlist URL as required
        console.log("Entered playlist URL:", playlistUrl);
        // Continue with your logic here...
    };

    return (
        <div>
            <h2>Enter Spotify Playlist URL</h2>
            <input
                type="text"
                value={playlistUrl}
                onChange={handleInputChange}
                placeholder="Paste your Spotify playlist URL here"
            />
            <button onClick={handleSubmit}>Process Playlist</button>
        </div>
    );
}

export default PlaylistInputPage;
