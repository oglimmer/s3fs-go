package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Storage root directory - configurable via command line
var storageRootDir string

// sanitizePath takes a bucket name and a key (possibly containing slashes),
// and returns the absolute path where that object should live.
// It also verifies no “../” path-traversal escapes the root.
func sanitizePath(bucket, key string) (string, error) {
	// Join bucket and key under storageRootDir
	joined := filepath.Join(storageRootDir, bucket, key)
	// Clean the path (e.g. remove “..” segments)
	cleaned := filepath.Clean(joined)

	// Make both absolute
	absRoot, err := filepath.Abs(storageRootDir)
	if err != nil {
		return "", err
	}
	absTarget, err := filepath.Abs(cleaned)
	if err != nil {
		return "", err
	}

	// Ensure absTarget is under absRoot
	if !strings.HasPrefix(absTarget, absRoot+string(os.PathSeparator)) && absTarget != absRoot {
		return "", errors.New("invalid path: path traversal detected")
	}
	return absTarget, nil
}

// uploadHandler handles PUT /<bucket>/<key...>
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept PUT
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Trim leading slash, then split off the bucket name
	// e.g. /my-bucket/folder/file.txt → ["my-bucket", "folder/file.txt"]
	trimmed := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Bad Request: missing bucket", http.StatusBadRequest)
		return
	}
	bucket := parts[0]
	var key string
	if len(parts) == 2 {
		key = parts[1]
	} else {
		// If no key provided (e.g. “PUT /my-bucket/”), treat as empty key,
		// but we don’t allow empty keys. Return 400.
		http.Error(w, "Bad Request: missing key", http.StatusBadRequest)
		return
	}

	log.Printf("Debug: %s request received for bucket=%s, key=%s", r.Method, bucket, key)

	// Resolve and sanitize filesystem path
	targetPath, err := sanitizePath(bucket, key)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Ensure the parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		log.Printf("Error creating directories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create/truncate the file and stream the request body into it
	f, err := os.Create(targetPath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Copy body to file (streaming)
	if _, err := io.Copy(f, r.Body); err != nil {
		log.Printf("Error writing file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Debug: Successfully processed %s request for bucket=%s, key=%s", r.Method, bucket, key)

	// Respond with 204 No Content (same as S3 when no ETag/key metadata is returned)
	w.WriteHeader(http.StatusNoContent)
}

// downloadHandler handles GET /<bucket>/<key...>
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Bad Request: missing bucket", http.StatusBadRequest)
		return
	}
	bucket := parts[0]
	var key string
	if len(parts) == 2 {
		key = parts[1]
	} else {
		http.Error(w, "Bad Request: missing key", http.StatusBadRequest)
		return
	}

	log.Printf("Debug: %s request received for bucket=%s, key=%s", r.Method, bucket, key)

	targetPath, err := sanitizePath(bucket, key)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Open the file
	f, err := os.Open(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			log.Printf("Error opening file: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	defer f.Close()

	// Optionally, set Content-Type based on file extension,
	// but here we default to application/octet-stream for simplicity.
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(key)+"\"")
	w.WriteHeader(http.StatusOK)

	// Stream the file back
	if _, err := io.Copy(w, f); err != nil {
		log.Printf("Error streaming file: %v", err)
	}
	log.Printf("Debug: Successfully processed %s request for bucket=%s, key=%s", r.Method, bucket, key)
}

// deleteHandler handles DELETE /<bucket>/<key...>
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept DELETE
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Bad Request: missing bucket", http.StatusBadRequest)
		return
	}
	bucket := parts[0]
	var key string
	if len(parts) == 2 {
		key = parts[1]
	} else {
		http.Error(w, "Bad Request: missing key", http.StatusBadRequest)
		return
	}

	log.Printf("Debug: %s request received for bucket=%s, key=%s", r.Method, bucket, key)

	targetPath, err := sanitizePath(bucket, key)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Delete the file
	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			// S3 returns 204 even if object doesn't exist
			w.WriteHeader(http.StatusNoContent)
		} else {
			log.Printf("Error deleting file: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Debug: Successfully processed %s request for bucket=%s, key=%s", r.Method, bucket, key)

	// Successfully deleted - return 204 No Content (S3 compatible)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <storage-root-path>\n", os.Args[0])
		os.Exit(1)
	}
	storageRootDir = os.Args[1]

	// Ensure storage root exists
	if err := os.MkdirAll(storageRootDir, 0o755); err != nil {
		log.Fatalf("Unable to create storage root '%s': %v", storageRootDir, err)
	}

	// Use DefaultServeMux; register a single catch-all handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			uploadHandler(w, r)
		case http.MethodGet:
			downloadHandler(w, r)
		case http.MethodDelete:
			deleteHandler(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := ":8080"
	log.Printf("Starting S3-FS-Go on %s, storing at %s", addr, storageRootDir)
	log.Printf("Debug: Server configured with handlers for PUT, GET, DELETE methods")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}