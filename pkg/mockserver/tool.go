package mockserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/art-frela/routeros/types"
)

const (
	EchoSize   string = "56"
	EchoTTL    string = "128"
	EchoStatus string = "timeout"
)

func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	if !s.auth(w, r) {
		return
	}

	if !s.checkPathAndMethods(w, r, types.EndpointToolPing, []string{http.MethodPost}) {
		return
	}

	var echo types.EchoRequest
	if err := json.NewDecoder(r.Body).Decode(&echo); err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, types.Error{
			Detail:  err.Error(),
			Error:   http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		})

		return
	}

	writeResponseJSON(w, http.StatusOK, s.reachableHosts.echo(echo))
}

type reachableHosts map[string]struct{}

func (rh reachableHosts) echo(echo types.EchoRequest) types.EchoResponse {
	if len(rh) == 0 {
		return makeEchoResponse(echo, false)
	}

	_, ok := rh[echo.Address]

	return makeEchoResponse(echo, ok)
}

func makeEchoResponse(echo types.EchoRequest, ok bool) types.EchoResponse {
	res := make(types.EchoResponse, echo.Count)

	for i := range echo.Count {
		pong := types.EchoResponseElement{
			Host:       echo.Address,
			PacketLoss: "100",
			Received:   "0",
			Sent:       strconv.Itoa(int(i + 1)),
			Seq:        strconv.Itoa(int(i)),
			Status:     Ptr(EchoStatus),
		}

		if ok {
			pong.PacketLoss = "0"
			pong.Received = strconv.Itoa(int(i + 1))
			pong.Size = Ptr(EchoSize)
			pong.TTL = Ptr(EchoTTL)
			pong.Status = nil
		}

		res[i] = pong
	}

	return res
}
