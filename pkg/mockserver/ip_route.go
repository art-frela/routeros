package mockserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/art-frela/routeros/types"
)

func (s *Server) IPRoutes(w http.ResponseWriter, r *http.Request) {
	if !s.auth(w, r) {
		return
	}

	// RouterOS REST accepts the resource id as a trailing path segment
	// (/rest/ip/route/*5); PUT always targets the collection itself.
	pathID, okPath := resourceIDFromPath(r.URL.Path, types.EndpointIPRoutes)
	if !okPath || (pathID != "" && r.Method == http.MethodPut) {
		writeResponseJSON(w, http.StatusBadRequest, types.Error{
			Detail:  "no such command or directory (...)",
			Error:   http.StatusBadRequest,
			Message: http.StatusText(http.StatusBadRequest),
		})

		return
	}

	// Normalize path-form URLs onto the collection path so the shared
	// path/method gate treats both forms identically.
	r.URL.Path = path.Join(types.EndpointRest, types.EndpointIPRoutes)

	if !s.checkPathAndMethods(w, r, types.EndpointIPRoutes, []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}) {
		return
	}

	if r.Method == http.MethodGet {
		id := r.URL.Query().Get(".id")
		if id == "" {
			id = pathID
		}

		if id != "" {
			// Find specific route by ID
			for _, routeList := range s.ipRoutes {
				for _, route := range routeList {
					if route.ID == id {
						writeResponseJSON(w, http.StatusOK, types.IPRouteList{route})
						return
					}
				}
			}
			// Return empty list if ID not found
			writeResponseJSON(w, http.StatusOK, types.IPRouteList{})
			return
		}
		// Return all routes
		allRoutes := make(types.IPRouteList, 0)
		for _, routeList := range s.ipRoutes {
			allRoutes = append(allRoutes, routeList...)
		}
		writeResponseJSON(w, http.StatusOK, allRoutes)
		return
	}

	if r.Method == http.MethodPut {
		// Add new route
		var newRoute types.IPRouteAdd
		if err := json.NewDecoder(r.Body).Decode(&newRoute); err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, types.Error{
				Detail:  err.Error(),
				Error:   http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		route, err := s.ipRoutes.add(newRoute)
		if err != nil {
			writeResourceError(w, err)
			return
		}

		writeResponseJSON(w, http.StatusOK, route)
		return
	}

	if r.Method == http.MethodDelete {
		// Remove route by ID
		id := r.URL.Query().Get(".id")
		if id == "" {
			id = pathID
		}

		if id == "" {
			writeResponseJSON(w, http.StatusBadRequest, types.Error{
				Detail:  "missing .id parameter",
				Error:   http.StatusBadRequest,
				Message: http.StatusText(http.StatusBadRequest),
			})
			return
		}

		if err := s.ipRoutes.remove(id); err != nil {
			writeResourceError(w, err)
			return
		}

		writeResponseJSON(w, http.StatusOK, struct{}{})
		return
	}

	if r.Method == http.MethodPatch {
		// Update route by ID
		id := r.URL.Query().Get(".id")
		if id == "" {
			id = pathID
		}

		if id == "" {
			writeResponseJSON(w, http.StatusBadRequest, types.Error{
				Detail:  "missing .id parameter",
				Error:   http.StatusBadRequest,
				Message: http.StatusText(http.StatusBadRequest),
			})
			return
		}

		var updateRoute types.IPRouteAdd
		if err := json.NewDecoder(r.Body).Decode(&updateRoute); err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, types.Error{
				Detail:  err.Error(),
				Error:   http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		route, err := s.ipRoutes.update(id, updateRoute)
		if err != nil {
			writeResourceError(w, err)
			return
		}

		writeResponseJSON(w, http.StatusOK, route)
		return
	}
}

type ipRoutesMap map[string]types.IPRouteList

func (m ipRoutesMap) add(item types.IPRouteAdd) (types.IPRoute, error) {
	newItem := types.IPRoute{
		ID:          newKey(),
		DstAddress:  item.DstAddress,
		Gateway:     item.Gateway,
		Distance:    item.Distance,
		Scope:       item.Scope,
		TargetScope: item.TargetScope,
		Static:      "true",
		Active:      "true",
		Dynamic:     "false",
		Disabled:    item.Disabled,
		Comment:     item.Comment,
	}

	if m["default"] == nil {
		m["default"] = types.IPRouteList{newItem}
	} else {
		m["default"] = append(m["default"], newItem)
	}

	return newItem, nil
}

func (m ipRoutesMap) remove(id string) error {
	if m["default"] == nil {
		return nil
	}

	idx := -1
	for i, route := range m["default"] {
		if route.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("route with id %s: %w", id, errResourceNotFound)
	}

	m["default"] = append(m["default"][:idx], m["default"][idx+1:]...)
	return nil
}

func (m ipRoutesMap) update(id string, item types.IPRouteAdd) (types.IPRoute, error) {
	if m["default"] == nil {
		return types.IPRoute{}, fmt.Errorf("route with id %s: %w", id, errResourceNotFound)
	}

	idx := -1
	for i, route := range m["default"] {
		if route.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return types.IPRoute{}, fmt.Errorf("route with id %s: %w", id, errResourceNotFound)
	}

	// Preserve existing fields and update only the provided ones
	existing := m["default"][idx]
	updatedItem := types.IPRoute{
		ID:          id,
		DstAddress:  item.DstAddress,
		Gateway:     item.Gateway,
		Distance:    item.Distance,
		Scope:       item.Scope,
		TargetScope: item.TargetScope,
		Static:      existing.Static,
		Active:      existing.Active,
		Dynamic:     existing.Dynamic,
		Disabled:    item.Disabled,
		Comment:     item.Comment,
	}

	if updatedItem.Distance == "" {
		updatedItem.Distance = existing.Distance
	}
	if updatedItem.Scope == "" {
		updatedItem.Scope = existing.Scope
	}
	if updatedItem.TargetScope == "" {
		updatedItem.TargetScope = existing.TargetScope
	}
	if updatedItem.Comment == "" {
		updatedItem.Comment = existing.Comment
	}
	if updatedItem.Disabled == "" {
		updatedItem.Disabled = existing.Disabled
	}

	m["default"][idx] = updatedItem
	return updatedItem, nil
}

// WithIPRoutes configures the mock server with pre-populated IP routes.
func WithIPRoutes(routes []types.IPRoute) Option {
	return func(s *Server) {
		if s.ipRoutes == nil {
			s.ipRoutes = make(ipRoutesMap)
		}
		s.ipRoutes["default"] = routes
	}
}
