package server

import (
	"fmt"
	"log/slog"
	"net/http"
)

type Server struct {
	srv *http.Server
	mux *http.ServeMux
	log *slog.Logger
}

func NewServer(addr string) *Server {
	mux := http.NewServeMux()
	return &Server{
		srv: &http.Server{Addr: addr, Handler: mux},
		mux: mux,
		log: slog.With("service", "server"),
	}
}

func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Start() error {
	s.log.Info(fmt.Sprintf("starting http server on %s", s.srv.Addr))

	return s.srv.ListenAndServe()
}
