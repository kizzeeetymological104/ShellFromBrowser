package transfer

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "file parameter required", http.StatusBadRequest)
		return
	}

	cleaned := filepath.Clean(filename)
	if strings.Contains(cleaned, "..") || filepath.IsAbs(cleaned) {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(h.baseDir, cleaned)

	absBase, _ := filepath.Abs(h.baseDir)
	absPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absPath, absBase) {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	_, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(cleaned)+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, fullPath)
}
