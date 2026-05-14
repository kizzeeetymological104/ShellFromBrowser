package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/valorisa/ShellFromBrowser/internal/auth"
	"github.com/valorisa/ShellFromBrowser/internal/config"
	"github.com/valorisa/ShellFromBrowser/internal/terminal"
	"github.com/valorisa/ShellFromBrowser/web"
)

type Server struct {
	addr     string
	mux      *http.ServeMux
	cfg      *config.Config
	authProv auth.Provider
	sessions *terminal.Manager
}

func New(addr string, cfg *config.Config) *Server {
	s := &Server{addr: addr, mux: http.NewServeMux(), cfg: cfg}

	if cfg.Auth.Enabled {
		s.authProv = auth.NewLocalProvider(&cfg.Auth)
	}

	s.sessions = terminal.NewManager(cfg.Sessions.MaxPerUser)

	s.mux.HandleFunc("/api/login", s.handleLogin)
	s.mux.HandleFunc("/api/sessions", s.authMiddleware(s.handleSessions))
	s.mux.HandleFunc("/ws", s.authMiddleware(s.handleWebSocket))

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
	if s.cfg.Server.TLS.Enabled {
		return http.ListenAndServeTLS(s.addr, s.cfg.Server.TLS.Cert, s.cfg.Server.TLS.Key, s.mux)
	}
	return http.ListenAndServe(s.addr, s.mux)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Error string `json:"error,omitempty"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, loginResponse{Error: "invalid request"})
		return
	}

	if s.authProv == nil {
		writeJSON(w, http.StatusOK, loginResponse{Token: "no-auth"})
		return
	}

	token, err := s.authProv.Authenticate(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, loginResponse{Error: "invalid credentials"})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authProv == nil {
			next(w, r)
			return
		}

		token := r.URL.Query().Get("token")
		if token == "" {
			authHeader := r.Header.Get("Authorization")
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		_, err := s.authProv.ValidateToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	username := "anonymous"
	if s.authProv != nil {
		token := r.URL.Query().Get("token")
		if token == "" {
			h := r.Header.Get("Authorization")
			if len(h) > 7 {
				token = h[7:]
			}
		}
		claims, err := s.authProv.ValidateToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		username = claims.Username
	}

	switch r.Method {
	case http.MethodGet:
		sessions := s.sessions.ListByUser(username)
		writeJSON(w, http.StatusOK, sessions)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id != "" {
			s.sessions.Destroy(id)
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
