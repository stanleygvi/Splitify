import React, { useEffect, useState } from 'react';
import './PlaylistInputPage.css'; // Importing the CSS file for styling

function PlaylistInputPage() {
    const [playlists, setPlaylists] = useState([]);
    const [selectedPlaylists, setSelectedPlaylists] = useState([]);

    useEffect(() => {
        fetch("http://localhost:8888/user-playlists")
        .then(response => response.json())
        .then(data => {
            if (data && data.items) {
                setPlaylists(data.items);
            } else {
                console.error("Unexpected data structure from server:", data);
            }
        })
        .catch(error => {
            console.error("There was an error fetching the playlists:", error);
        });
    }, []);

    const handlePlaylistSelection = (id) => {
        if (selectedPlaylists.includes(id)) {
            setSelectedPlaylists(prev => prev.filter(playlistId => playlistId !== id));
        } else {
            setSelectedPlaylists(prev => [...prev, id]);
        }
    }

    const handleProcessPlaylists = () => {
        console.log("Selected Playlists:", selectedPlaylists);
        
        fetch("http://localhost:8888/process-playlist", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ playlistIds: selectedPlaylists })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error("Network response was not ok");
            }
            return response.json();
        })
        .then(data => {
            console.log("Response from server:", data);
        })
        .catch(error => {
            console.error("There was a problem with the fetch operation:", error);
        });
    }

    return (
        <div>
            <h2>Select Playlists to Process</h2>
            <ul className="playlist-list">
                {playlists.map(playlist => (
                    <li 
                      key={playlist.id} 
                      className={selectedPlaylists.includes(playlist.id) ? 'selected' : ''} 
                      onClick={() => handlePlaylistSelection(playlist.id)}
                    >
                        <img 
                            src={playlist.images[0]?.url || "https://www.fredsmithxmastrees.com/wp-content/uploads/2017/04/Square-500x500-dark-grey.png"} 
                            alt={playlist.name + " cover image"} 
                            width={300} 
                            height={300}
                        />
                        <p>{playlist.name}</p>
                    </li>
                ))}
            </ul>
            <button onClick={handleProcessPlaylists}>Process Selected Playlists</button>
        </div>
    );

}

export default PlaylistInputPage;
