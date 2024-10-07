# Read me

### Disclaimer

Hi! Thank you for checking out Splitify! I am currently trying to get an extension from Spotify in order to have this fully deployed. If you want to try it out feel free to shoot me an email (stanleygvi@gmail.com) with your Spotify email and I can add you to the authorized users.

## Description

Splitify is a free to use tool that will split any selected playlist in your library into smaller musically similar playlists.

Playlists will be split according to Spotify's audio features provided by their API: https://developer.spotify.com/documentation/web-api/reference/get-audio-features

# Privacy Policy

At Splitify, we take your privacy and data security seriously. This document explains the specific permissions (scopes) we request from Spotify and why they are necessary to provide the functionality of the Splitify tool.

## Spotify Scopes We Use

### `user-read-private`

**Why we need it**:  
This scope is required to authenticate you with Spotify using OAuth (Spotify's secure login mechanism). It allows us to access basic information about your Spotify account, such as your Spotify user ID, which is necessary for actions like retrieving your playlists or creating new ones. 

**What we access**:  
- Basic account details like your user ID, country, and display name.  
- **We do not access** your subscription type, listening habits, or any other private data beyond what is necessary to identify you as the user.

**What we don't do**:  
- We do not store or use any private information for purposes outside of our service.  
- We do not share your data with third parties.

### `playlist-modify-public`

**Why we need it**:  
This scope is required to create new public playlists on your behalf when we split your playlists using the Splitify tool. It allows us to add tracks to these newly created playlists and manage them for you.

**What we access**:  
- We only modify public playlists that we create through Splitify.  
- **We do not modify** any existing playlists

**What we don't do**:  
- We do not add, remove, or change any tracks in your public playlists without your explicit instruction.
- We do not create or modify any private playlists unless requested in future features.

### `playlist-read-collaborative`

**Why we need it**:  
This scope allows us to access collaborative playlists, where multiple users can add and edit tracks. If you want to split or process any collaborative playlists with Splitify, this permission is necessary to retrieve the tracks and playlists you collaborate on.

**What we access**:  
- We only access collaborative playlists to read their contents.  
- **We do not modify** or change anything in the playlist.

**What we don't do**:  
- We do not make any unauthorized modifications to your collaborative playlists or interfere with other users’ contributions.

## Your Data, Your Control

We request only the permissions that are absolutely necessary to provide the core functionality of Splitify. We do not request access to your email address, personal playlists, or any data unrelated to managing your public or collaborative playlists through our service.

We are committed to being transparent about what data we access and ensuring that your data is protected. If you ever have concerns about your privacy or the permissions requested, feel free to reach out to us at [support@splitifytool.com].

Thank you for trusting Splitify with your playlist management!

## Contact Us

For any questions or concerns about your data and privacy, please contact us:

- Email: [support@splitifytool.com]
- Website: [https://splitifytool.com]
