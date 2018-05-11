package lobby

import (
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"github.com/d4l3k/messagediff"
)

func TestLobbyRPC(t *testing.T) {
	// Setup the server and client
	s := NewServer()
	ts := httptest.NewServer(s.mux)
	defer ts.Close()
	origin := ts.URL
	url := strings.Replace(ts.URL, "http", "ws", 1) + "/ws" // "ws://localhost:12345/ws"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}
	c := rpc2.NewClientWithCodec(jsonrpc.NewJSONCodec(ws))
	go c.Run()

	c.Handle("client.connect", func(client *rpc2.Client, req *ConnectLobbyRequest, resp *ConnectLobbyResponse) error {
		if req.Password != "test" {
			return ErrNotAuthorized
		}
		*resp = ConnectLobbyResponse{Answer: "webrtc offer"}
		return nil
	})

	testData := []struct {
		rpc      string
		req      interface{}
		resp     interface{}
		expected interface{}
		err      error
	}{
		{
			"lobby.list",
			&ListLobbyRequest{Service: "test"},
			&ListLobbyResponse{},
			&ListLobbyResponse{},
			nil,
		},
		{
			"lobby.connect",
			&ConnectLobbyRequest{},
			&ConnectLobbyResponse{},
			&ConnectLobbyResponse{},
			ErrNotFound,
		},
		{
			"lobby.new",
			&Lobby{
				ID:               "1",
				Name:             "Duck",
				Creator:          "d4l3k",
				RequiresPassword: true,
				Service:          "test",
				Location:         &Location{1, 1},
			},
			&NewLobbyResponse{},
			&NewLobbyResponse{},
			nil,
		},
		{
			"lobby.new",
			&Lobby{
				ID:      "2",
				Service: "test",
				Hidden:  true,
			},
			&NewLobbyResponse{},
			&NewLobbyResponse{},
			nil,
		},
		{
			"lobby.list",
			&ListLobbyRequest{
				Service: "test",
			},
			&ListLobbyResponse{},
			&ListLobbyResponse{
				Lobbies: []*Lobby{
					{
						ID:               "1",
						Name:             "Duck",
						Creator:          "d4l3k",
						RequiresPassword: true,
					},
				},
			},
			nil,
		},
		{
			"lobby.list",
			&ListLobbyRequest{
				Service:  "test",
				Location: &Location{1, 2},
			},
			&ListLobbyResponse{},
			&ListLobbyResponse{
				Lobbies: []*Lobby{
					{
						ID:               "1",
						Name:             "Duck",
						Creator:          "d4l3k",
						RequiresPassword: true,
						Distance:         111.17799068882648,
					},
				},
			},
			nil,
		},
		{
			"lobby.list",
			&ListLobbyRequest{Service: "other service"},
			&ListLobbyResponse{},
			&ListLobbyResponse{},
			nil,
		},
		{
			"lobby.connect",
			&ConnectLobbyRequest{ID: "1"},
			&ConnectLobbyResponse{},
			&ConnectLobbyResponse{},
			ErrNotAuthorized,
		},
		{
			"lobby.connect",
			&ConnectLobbyRequest{ID: "1", Password: "test"},
			&ConnectLobbyResponse{},
			&ConnectLobbyResponse{Answer: "webrtc offer"},
			ErrNotAuthorized,
		},
	}

	for i, td := range testData {
		if err := c.Call(td.rpc, td.req, td.resp); err != nil && (td.err == nil || err.Error() != td.err.Error()) {
			t.Error(err)
		}
		if diff, equal := messagediff.PrettyDiff(td.expected, td.resp); !equal {
			t.Errorf("%d. %s(%+v) = %+v; not %+v; diff %s", i, td.rpc, td.req, td.resp, td.expected, diff)
		}
	}

	// Close first connection and try second to make sure the lobby has been removed
	c.Close()

	ws2, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}
	c2 := rpc2.NewClientWithCodec(jsonrpc.NewJSONCodec(ws2))
	defer c2.Close()
	go c2.Run()

	testData2 := []struct {
		rpc      string
		req      interface{}
		resp     interface{}
		expected interface{}
		err      error
	}{
		{
			"lobby.list",
			&ListLobbyRequest{
				Service: "test",
			},
			&ListLobbyResponse{},
			&ListLobbyResponse{},
			nil,
		},
	}

	for i, td := range testData2 {
		if err := c2.Call(td.rpc, td.req, td.resp); err != nil && (td.err == nil || err.Error() != td.err.Error()) {
			t.Error(err)
		}
		if diff, equal := messagediff.PrettyDiff(td.expected, td.resp); !equal {
			t.Errorf("%d. %s(%+v) = %+v; not %+v; diff %s", i, td.rpc, td.req, td.resp, td.expected, diff)
		}
	}
}
