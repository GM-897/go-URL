package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type URL struct {
	ID           string    `json:"id"`
	OriginalURL  string    `json:"original_url"`
	ShortURL     string    `json:"short_url"`
	CreationDate time.Time `json:"creation_date"`
}

var urlDB = make(map[string]URL)

// Frontend HTML/CSS/JS as string literals
const (
	indexHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>URL Shortener</title>
    <style>
        %s
    </style>
</head>
<body>
    <div class="container">
        <h1>URL Shortener</h1>
        <form id="shortenForm">
            <input type="url" id="originalUrl" placeholder="Enter your URL" required>
            <button type="submit">Shorten</button>
        </form>
        <div id="result" class="hidden">
            <p>Short URL: <a id="shortUrl" target="_blank"></a></p>
        </div>
    </div>
    <script>
        %s
    </script>
</body>
</html>`
	
	cssStyles = `
body {
    font-family: Arial, sans-serif;
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
}

.container {
    text-align: center;
}

input[type="url"] {
    width: 300px;
    padding: 8px;
    margin-right: 10px;
}

button {
    padding: 8px 20px;
    background: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.hidden {
    display: none;
}

#shortUrl {
    color: #007bff;
    text-decoration: none;
}`

	javascriptCode = `
document.getElementById('shortenForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const urlInput = document.getElementById('originalUrl');
    const resultDiv = document.getElementById('result');
    const shortUrlLink = document.getElementById('shortUrl');

    try {
        const response = await fetch('/shorten', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ url: urlInput.value })
        });

        const data = await response.json();
        shortUrlLink.href = '/r/' + data.short_url;
        shortUrlLink.textContent = window.location.host + '/r/' + data.short_url;
        resultDiv.classList.remove('hidden');
        urlInput.value = '';
    } catch (error) {
        alert('Error shortening URL');
    }
});`
)

func generateShortURL(originalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalURL))
	data := hasher.Sum(nil)
	return hex.EncodeToString(data)[:8]
}

func createURL(originalURL string) string {
	shortURL := generateShortURL(originalURL)
	urlDB[shortURL] = URL{
		ID:           shortURL,
		OriginalURL:  originalURL,
		ShortURL:     shortURL,
		CreationDate: time.Now(),
	}
	return shortURL
}

func getURL(id string) (URL, error) {
	url, ok := urlDB[id]
	if !ok {
		return URL{}, errors.New("URL not found")
	}
	return url, nil
}

func RootPageURL(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, indexHTML, cssStyles, javascriptCode)
}

func ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortURL := createURL(data.URL)
	json.NewEncoder(w).Encode(map[string]string{"short_url": shortURL})
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/r/"):]
	url, err := getURL(id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

func main() {
	http.HandleFunc("/", RootPageURL)
	http.HandleFunc("/shorten", ShortURLHandler)
	http.HandleFunc("/r/", redirectURLHandler)

	fmt.Println("Server running on :3000")
	http.ListenAndServe(":3000", nil)
}