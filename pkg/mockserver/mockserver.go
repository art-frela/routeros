package mockserver

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/art-frela/routeros/types"
)

type Server struct {
	basicAuth      string
	ipFwList       ipFwList
	reachableHosts reachableHosts
}

func New(user, pass string) *Server {
	return &Server{basicAuth: base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))}
}

type Option func(*Server)

func WithIPFireWallAddressList(addressList map[string]types.FirewallAddressList) Option {
	return func(s *Server) {
		s.ipFwList = addressList
	}
}

func WithReachableHosts(hosts ...string) Option {
	return func(s *Server) {
		s.reachableHosts = make(reachableHosts)

		for _, host := range hosts {
			s.reachableHosts[host] = struct{}{}
		}
	}
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) bool {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Basic ")
	if s.basicAuth != token {
		writeResponseJSON(w, http.StatusUnauthorized, &types.Error{
			Error:   http.StatusUnauthorized,
			Message: http.StatusText(http.StatusUnauthorized),
		})

		return false
	}

	return true
}

func (s *Server) checkPathAndMethods(w http.ResponseWriter, r *http.Request, endpoint string, allowedMethods []string) bool {
	if r.URL.Path != path.Join(types.EndpointRest, endpoint) {
		writeResponseJSON(w, http.StatusBadRequest, types.Error{
			Detail:  "no such command or directory (...)",
			Error:   http.StatusBadRequest,
			Message: http.StatusText(http.StatusBadRequest),
		})

		return false
	}

	if !slices.Contains(allowedMethods, r.Method) {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, `<!doctype html>
<html lang=en>
<title>Error 503 : unknown method</title>
<h1>Error 503 : unknown method</h1>`)

		return false
	}

	return true
}

func writeResponseJSON(w http.ResponseWriter, code int, resp any) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}
