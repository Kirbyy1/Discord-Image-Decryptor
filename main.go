package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	webview "github.com/webview/webview_go"
)

// Embed the HTML, CSS, and JavaScript files.
//
//go:embed index.html styles.css script.js
var content embed.FS

type ImageInfo struct {
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Size      string `json:"size"`
}

// getDiscordCachePath returns the path to the Discord cache directory
func getDiscordCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var cachePath string
	switch runtime.GOOS {
	case "windows":
		cachePath = filepath.Join(homeDir, "AppData", "Roaming", "discord", "Cache", "Cache_Data")
	case "darwin":
		cachePath = filepath.Join(homeDir, "Library", "Application Support", "discord", "Cache")
	case "linux":
		cachePath = filepath.Join(homeDir, ".config", "discord", "Cache")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cachePath, nil
}

func main() {
	discordCachePath, err := getDiscordCachePath()
	if err != nil {
		fmt.Println("Error detecting Discord cache path:", err)
		return
	}

	http.Handle("/", http.FileServer(http.FS(content)))
	http.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		handleImages(w, r, discordCachePath)
	})
	http.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		serveImage(w, r, discordCachePath)
	})
	http.HandleFunc("/delete-cache", func(w http.ResponseWriter, r *http.Request) {
		deleteAllCache(w, r, discordCachePath)
	})
	http.HandleFunc("/delete-image", func(w http.ResponseWriter, r *http.Request) {
		deleteImage(w, r, discordCachePath)
	})
	http.HandleFunc("/open-file", func(w http.ResponseWriter, r *http.Request) {
		openFileExplorer(w, r, discordCachePath)
	})

	go func() {
		fmt.Println("Starting server at http://localhost:8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println("Failed to start server:", err)
		}
	}()

	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Discord Cache Viewer")
	w.SetSize(1024, 768, webview.HintNone)
	w.Navigate("http://localhost:8080")
	w.Run()
}

func handleImages(w http.ResponseWriter, r *http.Request, cachePath string) {
	files, err := os.ReadDir(cachePath)
	if err != nil {
		http.Error(w, "Unable to read cache directory", http.StatusInternalServerError)
		fmt.Println("Error reading cache directory:", err)
		return
	}

	var images []ImageInfo
	for _, file := range files {
		if !file.IsDir() {
			originalPath := filepath.Join(cachePath, file.Name())
			newPath := originalPath
			if !strings.HasSuffix(originalPath, ".png") {
				newPath = originalPath + ".png"
				err := os.Rename(originalPath, newPath)
				if err != nil {
					fmt.Println("Error renaming file:", err)
					continue
				}
			}

			fileInfo, err := os.Stat(newPath)
			if err != nil {
				fmt.Println("Error getting file info:", err)
				continue
			}

			imageInfo := ImageInfo{
				Name:      fileInfo.Name(),
				Timestamp: fileInfo.ModTime().Format(time.RFC3339),
				Size:      fmt.Sprintf("%.2f MB", float64(fileInfo.Size())/(1024*1024)),
			}
			images = append(images, imageInfo)
		}
	}

	sort.Slice(images, func(i, j int) bool {
		return images[i].Timestamp > images[j].Timestamp
	})

	response := map[string]interface{}{
		"total":  len(images),
		"images": images,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
		fmt.Println("Error encoding response:", err)
	}
}

func serveImage(w http.ResponseWriter, r *http.Request, cachePath string) {
	imageName := strings.TrimPrefix(r.URL.Path, "/images/")
	imagePath := filepath.Join(cachePath, imageName)

	http.ServeFile(w, r, imagePath)
}

func deleteAllCache(w http.ResponseWriter, r *http.Request, cachePath string) {
	files, err := os.ReadDir(cachePath)
	if err != nil {
		http.Error(w, "Unable to read cache directory", http.StatusInternalServerError)
		fmt.Println("Error reading cache directory:", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(cachePath, file.Name())
			err := os.Remove(filePath)
			if err != nil {
				http.Error(w, "Unable to delete file", http.StatusInternalServerError)
				fmt.Println("Error deleting file:", err)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func deleteImage(w http.ResponseWriter, r *http.Request, cachePath string) {
	imagePath := r.URL.Query().Get("path")
	if imagePath == "" {
		http.Error(w, "Missing image path", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(cachePath, imagePath)
	err := os.Remove(filePath)
	if err != nil {
		http.Error(w, "Unable to delete file", http.StatusInternalServerError)
		fmt.Println("Error deleting file:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func openFileExplorer(w http.ResponseWriter, r *http.Request, cachePath string) {
	imagePath := r.URL.Query().Get("path")
	if imagePath == "" {
		http.Error(w, "Missing image path", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(cachePath, imagePath)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", filePath)
	case "darwin":
		cmd = exec.Command("open", "-R", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		http.Error(w, "Unsupported operating system", http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, "Unable to open file explorer", http.StatusInternalServerError)
		fmt.Println("Error opening file explorer:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
