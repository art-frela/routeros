package mockserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/art-frela/routeros/types"
)

func (s *Server) IPAddresses(w http.ResponseWriter, r *http.Request) {
	if !s.auth(w, r) {
		return
	}

	if !s.checkPathAndMethods(w, r, types.EndpointIPAddresses, []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}) {
		return
	}

	if r.Method == http.MethodGet {
		id := r.URL.Query().Get(".id")
		if id != "" {
			// Find specific address by ID
			for _, ipList := range s.ipAddresses {
				for _, ip := range ipList {
					if ip.ID == id {
						writeResponseJSON(w, http.StatusOK, types.IPAddressList{ip})
						return
					}
				}
			}
			// Return empty list if ID not found
			writeResponseJSON(w, http.StatusOK, types.IPAddressList{})
			return
		}
		// Return all addresses
		allAddresses := make(types.IPAddressList, 0)
		for _, ipList := range s.ipAddresses {
			allAddresses = append(allAddresses, ipList...)
		}
		writeResponseJSON(w, http.StatusOK, allAddresses)
		return
	}

	if r.Method == http.MethodPut {
		// Add new address
		var newAddr types.IPAddressAdd
		if err := json.NewDecoder(r.Body).Decode(&newAddr); err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, types.Error{
				Detail:  err.Error(),
				Error:   http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		addr, err := s.ipAddresses.add(newAddr)
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, types.Error{
				Detail:  err.Error(),
				Error:   http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		writeResponseJSON(w, http.StatusOK, addr)
		return
	}

	if r.Method == http.MethodDelete {
		// Remove address by ID
		id := r.URL.Query().Get(".id")
		if id == "" {
			writeResponseJSON(w, http.StatusBadRequest, types.Error{
				Detail:  "missing .id parameter",
				Error:   http.StatusBadRequest,
				Message: http.StatusText(http.StatusBadRequest),
			})
			return
		}

		err := s.ipAddresses.remove(id)
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, types.Error{
				Detail:  err.Error(),
				Error:   http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		writeResponseJSON(w, http.StatusOK, struct{}{})
		return
	}

	if r.Method == http.MethodPatch {
		// Update address by ID
		id := r.URL.Query().Get(".id")
		if id == "" {
			writeResponseJSON(w, http.StatusBadRequest, types.Error{
				Detail:  "missing .id parameter",
				Error:   http.StatusBadRequest,
				Message: http.StatusText(http.StatusBadRequest),
			})
			return
		}

		var updateAddr types.IPAddressAdd
		if err := json.NewDecoder(r.Body).Decode(&updateAddr); err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, types.Error{
				Detail:  err.Error(),
				Error:   http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		addr, err := s.ipAddresses.update(id, updateAddr)
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, types.Error{
				Detail:  err.Error(),
				Error:   http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		writeResponseJSON(w, http.StatusOK, addr)
		return
	}
}

type ipAddressesMap map[string]types.IPAddressList

func (m ipAddressesMap) add(item types.IPAddressAdd) (types.IPAddress, error) {
	newItem := types.IPAddress{
		ID:              newKey(),
		Address:         item.Address,
		Network:         getNetworkFromAddress(item.Address),
		Interface:       item.Interface,
		ActualInterface: item.Interface,
		Dynamic:         "false",
		Disabled:        item.Disabled,
	}

	if m == nil {
		m = ipAddressesMap{"default": types.IPAddressList{newItem}}
	} else {
		if m["default"] == nil {
			m["default"] = types.IPAddressList{newItem}
		} else {
			m["default"] = append(m["default"], newItem)
		}
	}

	return newItem, nil
}

func (m ipAddressesMap) remove(id string) error {
	if m["default"] == nil {
		return nil
	}

	// Find the item to remove
	idx := -1
	for i, addr := range m["default"] {
		if addr.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("address with id %s not found", id) // Item not found
	}

	// Remove the item
	m["default"] = append(m["default"][:idx], m["default"][idx+1:]...)
	return nil
}

func (m ipAddressesMap) update(id string, item types.IPAddressAdd) (types.IPAddress, error) {
	if m["default"] == nil {
		return types.IPAddress{}, fmt.Errorf("address with id %s not found", id)
	}

	// Find the item to update
	idx := -1
	for i, addr := range m["default"] {
		if addr.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return types.IPAddress{}, fmt.Errorf("address with id %s not found", id)
	}

	// Update the item
	updatedItem := types.IPAddress{
		ID:              id,
		Address:         item.Address,
		Network:         getNetworkFromAddress(item.Address),
		Interface:       item.Interface,
		ActualInterface: item.Interface,
		Dynamic:         "false",
		Disabled:        item.Disabled,
	}

	m["default"][idx] = updatedItem
	return updatedItem, nil
}

func getNetworkFromAddress(address string) string {
	// Simple implementation to extract network from address like "192.168.1.1/24"
	// This would return "192.168.1.0" for the network part
	parts := []string{}
	for _, c := range address {
		if c == '/' {
			break
		}
		parts = append(parts, string(c))
	}

	// For now, just return a placeholder network
	// A full implementation would calculate the actual network based on CIDR
	return address
}

func WithIPAddresses(addresses []types.IPAddress) Option {
	return func(s *Server) {
		if s.ipAddresses == nil {
			s.ipAddresses = make(ipAddressesMap)
		}
		s.ipAddresses["default"] = addresses
	}
}
