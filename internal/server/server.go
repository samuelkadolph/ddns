package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/samuelkadolph/ddns/internal/config"
	"github.com/samuelkadolph/ddns/internal/route53"
)

type Server struct {
	config     *config.Config
	httpServer *http.Server
	r53        *route53.Manager
}

func NewServer(cfg *config.Config, listens Listens) (*Server, error) {
	server := &Server{config: cfg}

	mux := http.NewServeMux()
	mux.Handle("/update", &jsonAPI{s: server})
	mux.Handle("/nic/update", &noipAPI{s: server})
	server.httpServer = &http.Server{Handler: RequestLogger(mux)}

	r53, err := route53.NewManager(cfg)
	if err != nil {
		return nil, err
	}
	server.r53 = r53

	for _, listen := range listens {
		l, err := net.Listen("tcp", listen)
		if err != nil {
			log.Printf("Failed to listen on '%s': %s\n", listen, err)
			return nil, err
		}
		go server.httpServer.Serve(l)
		log.Printf("Listening on %s\n", listen)
	}

	return server, nil
}

func (s *Server) Shutdown() error {
	return s.httpServer.Shutdown(context.Background())
}

func (s *Server) verifyAuth(domain *config.Domain, username string, password string) bool {
	user, ok := s.config.AuthenticateUser(username, password)
	if !ok {
		return false
	}

	return s.config.AuthorizeUser(user, domain)
}
