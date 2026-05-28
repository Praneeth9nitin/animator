package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

type SSEEvent struct {
	Type     string `json:"type"`
	Message  string `json:"message,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func sendEvent(w http.ResponseWriter, event SSEEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Prompt) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sendEvent(w, SSEEvent{Type: "progress", Message: "Analyzing your prompt..."})
	code := Route(req.Prompt)

	sendEvent(w, SSEEvent{Type: "progress", Message: "Manim code generated"})

	const maxRetries = 5
	var lastStderr string
	success := false

	for i := 0; i < maxRetries; i++ {
		if i == 0 {
			sendEvent(w, SSEEvent{Type: "progress", Message: "Spinning up Docker container..."})
		} else {
			sendEvent(w, SSEEvent{Type: "progress", Message: fmt.Sprintf("Error detected — fixing and retrying (attempt %d/%d)...", i+1, maxRetries)})
			code = FixCode(code, lastStderr)
		}

		stderr, _ := Container(code)
		if stderr == "" {
			success = true
			break
		}
		lastStderr = stderr
	}

	if !success {
		sendEvent(w, SSEEvent{Type: "error", Message: "Generation failed after retries: " + lastStderr})
		return
	}

	videoPath := findLatestVideo("files/media")
	if videoPath == "" {
		sendEvent(w, SSEEvent{Type: "error", Message: "Docker ran but no video was produced"})
		return
	}

	relPath := strings.TrimPrefix(videoPath, "files/")
	sendEvent(w, SSEEvent{Type: "done", VideoURL: "/video/" + relPath})
}

func handleVideo(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	path := strings.TrimPrefix(r.URL.Path, "/video/")
	fullPath := filepath.Join("files/media", path)
	http.ServeFile(w, r, fullPath)
}

func findLatestVideo(dir string) string {
	var latestPath string
	var latestMod time.Time

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".mp4") && info.ModTime().After(latestMod) {
			latestPath = path
			latestMod = info.ModTime()
		}
		return nil
	})

	return latestPath
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/generate", handleGenerate)
	mux.HandleFunc("/video/", handleVideo)

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
