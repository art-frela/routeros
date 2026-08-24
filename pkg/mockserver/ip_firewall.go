package mockserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/art-frela/routeros/types"
)

func (s *Server) IPFirewallAddressList(w http.ResponseWriter, r *http.Request) {
	if !s.auth(w, r) {
		return
	}

	if !s.checkPathAndMethods(w, r, types.EndpointIPFirewallAddresList, []string{http.MethodGet, http.MethodPut}) {
		return
	}

	if r.Method == http.MethodGet {
		list := r.URL.Query().Get("list")
		addr := r.URL.Query().Get("address")
		writeResponseJSON(w, http.StatusOK, s.ipFwList.find(list, addr))

		return
	}

	// PUT (add new entry)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, types.Error{
			Detail:  err.Error(),
			Error:   http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		})

		return
	}

	var newItem types.FirewallAddressListNewItem
	if err := json.Unmarshal(body, &newItem); err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, types.Error{
			Detail:  err.Error(),
			Error:   http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		})

		return
	}

	if invalid := validateOptionalBooleans(body, "disabled", "dynamic"); invalid != nil {
		writeResponseJSON(w, http.StatusBadRequest, *invalid)

		return
	}

	item, er := s.ipFwList.add(newItem)
	if er != nil {
		writeResponseJSON(w, http.StatusInternalServerError, er)

		return
	}

	writeResponseJSON(w, http.StatusOK, item)
}

// validateOptionalBooleans mirrors RouterOS: optional boolean fields may be
// absent from the payload, but when present their value must be one of
// yes/no/true/false — anything else (including the empty string) draws
// 400 "invalid value of <field>, must be either yes or no".
func validateOptionalBooleans(payload []byte, keys ...string) *types.Error {
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil // malformed JSON is reported by the struct decode
	}

	for _, key := range keys {
		value, present := raw[key]
		if !present {
			continue
		}

		str, isString := value.(string)
		if !isString || !slices.Contains([]string{"yes", "no", "true", "false"}, str) {
			return &types.Error{
				Detail:  fmt.Sprintf("invalid value of %s, must be either yes or no", key),
				Error:   http.StatusBadRequest,
				Message: http.StatusText(http.StatusBadRequest),
			}
		}
	}

	return nil
}

type ipFwList map[string]types.FirewallAddressList

func (lst ipFwList) find(list, address string) types.FirewallAddressList {
	if len(lst) == 0 {
		return types.FirewallAddressList{}
	}

	if list == "" {
		res := make(types.FirewallAddressList, 0)
		for _, addresses := range lst {
			res = append(res, addresses.Find(address)...)
		}

		return res
	}

	return lst[list].Find(address)
}

func (lst ipFwList) add(item types.FirewallAddressListNewItem) (types.FirewallAddressListItem, *types.Error) {
	if len(item.List) == 0 || len(item.Address) == 0 {
		return types.FirewallAddressListItem{}, &types.Error{
			Detail:  "empty list or address",
			Error:   http.StatusServiceUnavailable,
			Message: http.StatusText(http.StatusServiceUnavailable),
		}
	}

	newItem := types.FirewallAddressListItem{
		ID:           newKey(),
		Address:      item.Address,
		CreationTime: types.DateTime{Time: time.Now().Truncate(time.Minute)},
		Disabled:     "false",
		Dynamic:      "false",
		List:         item.List,
	}

	if lst == nil {
		lst = ipFwList{item.List: types.FirewallAddressList{
			newItem,
		}}

		return newItem, nil
	}

	exists, ok := lst[item.List]
	if !ok {
		lst[item.List] = types.FirewallAddressList{
			newItem,
		}

		return newItem, nil
	}

	if slices.ContainsFunc(exists, func(item types.FirewallAddressListItem) bool {
		return item.Address == newItem.Address
	}) {
		return types.FirewallAddressListItem{}, &types.Error{
			Detail:  "failure: already have such entry",
			Error:   http.StatusBadRequest,
			Message: http.StatusText(http.StatusBadRequest),
		}
	}

	lst[item.List] = append(exists, newItem)

	return newItem, nil
}
