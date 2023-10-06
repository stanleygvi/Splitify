import React, { useEffect, useState } from 'react';

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
        // Handle the processing logic here (e.g., send selected playlists to the server for processing).
        console.log("Selected Playlists:", selectedPlaylists);
    }

    return (
        <div>
            <h2>Select Playlists to Process</h2>
            <ul>
                {playlists.map(playlist => (
                    <li key={playlist.id}>
                        <label>
                            <input 
                                type="checkbox"
                                value={playlist.id}
                                checked={selectedPlaylists.includes(playlist.id)}
                                onChange={() => handlePlaylistSelection(playlist.id)}
                            />
                            <img 
                                src={playlist.images[0]?.url || "https://www.fredsmithxmastrees.com/wp-content/uploads/2017/04/Square-500x500-dark-grey.png"} 
                                alt={playlist.name + " cover image"} 
                                width={60} 
                                height={60}
                            />
                            {playlist.name}
                        </label>
                    </li>
                ))}
            </ul>
            <button onClick={handleProcessPlaylists}>Process Selected Playlists</button>
        </div>
    );
}

export default PlaylistInputPage;
