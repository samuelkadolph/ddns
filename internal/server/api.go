package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type baseHandler struct {
	s *Server
}

type jsonAPI baseHandler
type noipAPI baseHandler

func (a *jsonAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		a.response(w, r, 404, `{"status":"error","error":"not found"}`)
		return
	}

	host := r.URL.Query().Get("host")
	if host == "" {
		a.response(w, r, 400, `{"status":"error","error":"missing parameter 'host'"}`)
		return
	}

	ip := r.URL.Query().Get("ip")
	if ip == "" {
		a.response(w, r, 400, `{"status":"error","error":"missing parameter 'ip'"}`)
		return
	}

	domain, ok := a.s.config.Domains[host]
	if !ok {
		a.response(w, r, 404, `{"status":"error","error":"not found"}`)
		return
	}

	username, password, err := decodeBasicAuth(r.Header.Get("Authorization"))
	if err != nil {
		a.response(w, r, 401, `{"status":"error","error":"unauthorized"}`)
		return
	}

	if !a.s.verifyAuth(domain, username, password) {
		a.response(w, r, 403, `{"status":"error","error":"forbidden"}`)
		return
	}

	if err := a.s.r53.UpdateDomain(domain, ip); err != nil {
		log.Printf("Failed to update %s: %s", domain.Name, err)
		a.response(w, r, 500, `{"status":"error","error":"%s"}`, err)
		return
	}

	log.Printf("Updated %s to %s", domain.Name, ip)
	a.response(w, r, 200, `{"status":"ok","ip":"%s"}`, ip)
	return
}

func (a *jsonAPI) response(w http.ResponseWriter, r *http.Request, code int, format string, v ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(format, v...))
}

func (a *noipAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		a.response(w, r, 200, "badagent")
		return
	}

	host := r.URL.Query().Get("hostname")
	if host == "" {
		a.response(w, r, 200, "badagent")
		return
	}

	ip := r.URL.Query().Get("myip")
	if ip == "" {
		a.response(w, r, 200, "badagent")
		return
	}

	domain, ok := a.s.config.Domains[host]
	if !ok {
		a.response(w, r, 200, "nohost")
		return
	}

	username, password, err := decodeBasicAuth(r.Header.Get("Authorization"))
	if err != nil {
		a.response(w, r, 401, "badauth")
		return
	}

	if !a.s.verifyAuth(domain, username, password) {
		a.response(w, r, 403, "badauth")
		return
	}

	if err := a.s.r53.UpdateDomain(domain, ip); err != nil {
		log.Printf("Failed to update %s: %s", domain.Name, err)
		a.response(w, r, 500, "911")
		return
	}

	log.Printf("Updated %s to %s", domain.Name, ip)
	a.response(w, r, 200, "good %s", ip)
	return
}

func (a *noipAPI) response(w http.ResponseWriter, r *http.Request, code int, format string, v ...interface{}) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(format, v...))
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
