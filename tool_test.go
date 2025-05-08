package routeros

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/art-frela/routeros/pkg/mockserver"
	"github.com/art-frela/routeros/types"
	"github.com/stretchr/testify/assert"
)

func TestToolService_Ping(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dummy := mockserver.New("root", "master")
	mockserver.WithReachableHosts("127.0.0.1")(dummy)

	ts := httptest.NewServer(http.HandlerFunc(dummy.Ping))
	defer ts.Close()

	t.Setenv("ROS_TEST_FIND_BASE_URL", ts.URL)
	t.Setenv("ROS_TEST_FIND_USER", "root")

	cfg, err := NewClientConfigFromEnv("ROS_TEST_FIND")
	if err != nil {
		t.Errorf("NewClientConfigFromEnv error: %s", err)

		return
	}

	c, err := NewClient(*cfg)
	if err != nil {
		t.Errorf("NewClient error: %s", err)

		return
	}

	tests := []struct {
		name    string
		req     types.EchoRequest
		want    types.EchoResponse
		wantErr bool
	}{
		{
			name: "test.1 ok reachable",
			req: types.EchoRequest{
				Address: "127.0.0.1",
				Count:   3,
			},
			want: types.EchoResponse{
				{
					Host:       "127.0.0.1",
					PacketLoss: "0",
					Received:   "1",
					Sent:       "1",
					Seq:        "0",
					Size:       mockserver.Ptr(mockserver.EchoSize),
					TTL:        mockserver.Ptr(mockserver.EchoTTL),
				},
				{
					Host:       "127.0.0.1",
					PacketLoss: "0",
					Received:   "2",
					Sent:       "2",
					Seq:        "1",
					Size:       mockserver.Ptr(mockserver.EchoSize),
					TTL:        mockserver.Ptr(mockserver.EchoTTL),
				},
				{
					Host:       "127.0.0.1",
					PacketLoss: "0",
					Received:   "3",
					Sent:       "3",
					Seq:        "2",
					Size:       mockserver.Ptr(mockserver.EchoSize),
					TTL:        mockserver.Ptr(mockserver.EchoTTL),
				},
			},
			wantErr: false,
		},
		{
			name: "test.2 unreachable",
			req: types.EchoRequest{
				Address: "ya.ru",
				Count:   2,
			},
			want: types.EchoResponse{
				{
					Host:       "ya.ru",
					PacketLoss: "100",
					Received:   "0",
					Sent:       "1",
					Seq:        "0",
					Status:     mockserver.Ptr(mockserver.EchoStatus),
				},
				{
					Host:       "ya.ru",
					PacketLoss: "100",
					Received:   "0",
					Sent:       "2",
					Seq:        "1",
					Status:     mockserver.Ptr(mockserver.EchoStatus),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &ToolService{c: c}

			got, err := ts.Ping(ctx, tt.req)

			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			if tt.wantErr {
				assert.Error(t, err, errMsg)
			} else {
				assert.NoError(t, err, errMsg)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
