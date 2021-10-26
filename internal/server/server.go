package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

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
	server.httpServer = &http.Server{
		Handler: server,
	}

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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/update" || r.Method != "GET" {
		jsonResponse(w, r, 404, "{\"status\":\"error\",\"error\":\"not found\"}")
		return
	}

	host := r.URL.Query().Get("host")
	if host == "" {
		jsonResponse(w, r, 400, "{\"status\":\"error\",\"error\":\"missing parameter 'host'\"}")
		return
	}

	ip := r.URL.Query().Get("ip")
	if ip == "" {
		jsonResponse(w, r, 400, "{\"status\":\"error\",\"error\":\"missing parameter 'ip'\"}")
		return
	}

	domain, ok := s.config.Domains[host]
	if !ok {
		jsonResponse(w, r, 404, "{\"status\":\"error\",\"error\":\"not found\"}")
		return
	}

	username, password, err := decodeBasicAuth(r.Header.Get("Authorization"))
	if err != nil {
		jsonResponse(w, r, 401, "{\"status\":\"error\",\"error\":\"unauthorized\"}")
		return
	}

	if !s.verifyAuth(domain, username, password) {
		jsonResponse(w, r, 403, "{\"status\":\"error\",\"error\":\"forbidden\"}")
		return
	}

	if err := s.r53.UpdateDomain(domain, ip); err != nil {
		log.Printf("Failed to update %s: %s", domain.Name, err)
		jsonResponse(w, r, 500, "{\"status\":\"error\",\"error\":\"%s\"}", err)
		return
	}

	log.Printf("Updated %s to %s", domain.Name, ip)
	jsonResponse(w, r, 200, "{\"status\":\"ok\",\"ip\":\"%s\"}", ip)
	return
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

func decodeBasicAuth(header string) (string, string, error) {
	if header == "" {
		return "", "", errors.New("no auth")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", errors.New("requires basic auth")
	}

	auth, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", err
	}

	pair := strings.SplitN(string(auth), ":", 2)
	if len(pair) != 2 {
		return "", "", errors.New("requires username and password")
	}

	return pair[0], pair[1], nil
}

func jsonResponse(w http.ResponseWriter, r *http.Request, code int, format string, a ...interface{}) {
	log.Printf("%s %s %s %d", r.Method, r.URL, r.Header.Get("User-Agent"), code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(format, a...))
}
