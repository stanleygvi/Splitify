// LoginPage.js
import React from 'react';
import './LoginPage.css';  // Optional: if you want to have separate styles for this page

function LoginPage() {
    const initiateLogin = () => {
        // Redirect the user to the server's /login endpoint to start the OAuth process.
        window.location.href = "https://www.splitifytool.com/login";

    };

    return (
        <div className="login-page">
            <h2>Welcome to Splitify</h2>
            <p>To get started, please login with your Spotify account:</p>
            <button onClick={initiateLogin}>Login with Spotify</button>
        </div>
    );
}

export default LoginPage;
