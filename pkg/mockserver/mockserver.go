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
	basicAuth string
	ipFwList  ipFwList
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

func (s *Server) auth(r *http.Request) *types.Error {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Basic ")
	if s.basicAuth != token {
		return &types.Error{
			Error:   http.StatusUnauthorized,
			Message: http.StatusText(http.StatusUnauthorized),
		}
	}

	return nil
}

func (s *Server) IPFirewallAddressList(w http.ResponseWriter, r *http.Request) {
	if er := s.auth(r); er != nil {
		writeResponseJSON(w, http.StatusUnauthorized, er)

		return
	}

	if r.URL.Path != path.Join(types.EndpointRest, types.EndpointIPFirewallAddresList) {
		writeResponseJSON(w, http.StatusBadRequest, types.Error{
			Detail:  "no such command or directory (...)",
			Error:   http.StatusBadRequest,
			Message: http.StatusText(http.StatusBadRequest),
		})

		return
	}

	var allowedMethods = []string{http.MethodGet, http.MethodPut}

	if !slices.Contains(allowedMethods, r.Method) {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, `<!doctype html>
<html lang=en>
<title>Error 503 : unknown method</title>
<h1>Error 503 : unknown method</h1>`)
		return
	}

	if r.Method == http.MethodGet {
		list := r.URL.Query().Get("list")
		addr := r.URL.Query().Get("address")
		writeResponseJSON(w, http.StatusOK, s.ipFwList.find(list, addr))

		return
	}

	// POST
	var newItem types.FirewallAddressListNewItem
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, types.Error{
			Detail:  err.Error(),
			Error:   http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		})

		return
	}

	item, er := s.ipFwList.add(newItem)
	if er != nil {
		writeResponseJSON(w, http.StatusInternalServerError, er)

		return
	}

	writeResponseJSON(w, http.StatusOK, item)
}

func writeResponseJSON(w http.ResponseWriter, code int, resp any) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}
