package server

import (
	"net/http"
)

type Server struct {
	addr string
	mux  *http.ServeMux
}

func New(addr string) *Server {
	s := &Server{addr: addr, mux: http.NewServeMux()}
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}
