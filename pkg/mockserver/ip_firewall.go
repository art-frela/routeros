package mockserver

import (
	"encoding/json"
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
