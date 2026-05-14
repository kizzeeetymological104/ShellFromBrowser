package transfer

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Handler struct {
	baseDir string
	maxSize int64
}

func NewHandler(baseDir string, maxSize int64) *Handler {
	os.MkdirAll(baseDir, 0755)
	return &Handler{baseDir: baseDir, maxSize: maxSize}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	if filename == "." || filename == ".." {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	destPath := filepath.Join(h.baseDir, filename)
	dest, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer dest.Close()

	written, err := io.Copy(dest, file)
	if err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"filename": filename,
		"size":     written,
	})
}
