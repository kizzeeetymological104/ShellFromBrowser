package server

import (
	"io/fs"
	"net/http"

	"github.com/valorisa/ShellFromBrowser/web"
)

type Server struct {
	addr string
	mux  *http.ServeMux
}

func New(addr string) *Server {
	s := &Server{addr: addr, mux: http.NewServeMux()}
	s.mux.HandleFunc("/ws", s.handleWebSocket)

	staticFS, _ := fs.Sub(web.StaticFiles, "static")
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, _ := web.StaticFiles.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}
