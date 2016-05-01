package lobby

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/cenk/rpc2"
	"github.com/cenk/rpc2/jsonrpc"
	"github.com/kellydunn/golang-geo"
	"golang.org/x/net/websocket"
)

var (
	ErrAlreadyExists = errors.New("a lobby with that ID has already been created by someone else")
	ErrNotFound      = errors.New("can't find the requested lobby")
	ErrNotAuthorized = errors.New("invalid credentials")
)

const lobbiesKey = "lobbies"

type Location struct {
	Latitude, Longitude float64
}

// Point returns the geo.Point that represents the location.
func (l Location) Point() *geo.Point {
	return geo.NewPoint(l.Latitude, l.Longitude)
}

type Lobby struct {
	ID, Name, Creator        string
	Hidden, RequiresPassword bool
	Location                 *Location
	Distance                 float64
	People, Capacity         int

	client *rpc2.Client
}

type Server struct {
	rpc         *rpc2.Server
	lobbies     map[string]*Lobby
	lobbiesLock sync.RWMutex
	mux         *http.ServeMux
}

// NewServer creates a Server.
func NewServer() *Server {
	s := &Server{
		rpc:     rpc2.NewServer(),
		lobbies: make(map[string]*Lobby),
		mux:     http.NewServeMux(),
	}

	s.mux.Handle("/ws", websocket.Handler(s.Serve))

	s.rpc.Handle("lobby.new", s.newLobby)
	s.rpc.Handle("lobby.connect", s.connectLobby)
	s.rpc.Handle("lobby.list", s.listLobby)
	s.rpc.OnDisconnect(s.disconnect)

	return s
}

type NewLobbyResponse struct{}

func (s *Server) newLobby(client *rpc2.Client, req *Lobby, resp *NewLobbyResponse) error {
	log.Printf("lobby.new: %+v", req)
	s.lobbiesLock.Lock()
	defer s.lobbiesLock.Unlock()
	if lobby, ok := s.lobbies[req.ID]; ok && lobby.client != client {
		return ErrAlreadyExists
	}
	req.client = client
	s.lobbies[req.ID] = req
	var lobbies []*Lobby
	if lobbiesInterface, ok := client.State.Get(lobbiesKey); ok {
		lobbies = lobbiesInterface.([]*Lobby)
	}
	lobbies = append(lobbies, req)
	client.State.Set(lobbiesKey, lobbies)
	return nil
}

func (s *Server) disconnect(client *rpc2.Client) {
	s.lobbiesLock.Lock()
	defer s.lobbiesLock.Unlock()

	lobbiesInterface, ok := client.State.Get(lobbiesKey)
	if !ok {
		return
	}
	for _, lobby := range lobbiesInterface.([]*Lobby) {
		delete(s.lobbies, lobby.ID)
	}
}

// ConnectLobbyRequest represents a request to connect to a lobby.
type ConnectLobbyRequest struct {
	ID       string
	Offer    string
	Password string
}

// ConnectLobbyResponse represents the response to a connect request.
type ConnectLobbyResponse struct {
	Offer string
}

func (s *Server) connectLobby(client *rpc2.Client, req *ConnectLobbyRequest, resp *ConnectLobbyResponse) error {
	log.Printf("lobby.connect: %+v", req)
	s.lobbiesLock.RLock()
	lobby, ok := s.lobbies[req.ID]
	s.lobbiesLock.RUnlock()
	if !ok {
		return ErrNotFound
	}
	return lobby.client.Call("client.connect", req, resp)
}

// ListLobbyRequest represents a request to list all available lobbies by location.
type ListLobbyRequest struct {
	Location *Location
}

// ListLobbyResponse represents the results returned by a ListLobbyRequest sorted by distance.
type ListLobbyResponse struct {
	Lobbies []*Lobby
}

func (s *Server) listLobby(client *rpc2.Client, req *ListLobbyRequest, resp *ListLobbyResponse) error {
	log.Printf("lobby.list: %+v", req)
	s.lobbiesLock.RLock()
	defer s.lobbiesLock.RUnlock()
	var lobbies []*Lobby
	for _, lobby := range s.lobbies {
		if lobby.Hidden {
			continue
		}
		l := *lobby
		if req.Location != nil {
			p := req.Location.Point()
			p2 := l.Location.Point()
			l.Distance = p.GreatCircleDistance(p2)
		}
		l.Location = nil
		lobbies = append(lobbies, &l)
	}
	*resp = ListLobbyResponse{Lobbies: lobbies}
	return nil
}

func (s *Server) Serve(ws *websocket.Conn) {
	s.rpc.ServeCodec(jsonrpc.NewJSONCodec(ws))
}

// Listen listens and serves the http endpoint at the address specified.
func (s *Server) Listen(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}
	log.Printf("Listening on %s...", addr)
	return server.ListenAndServe()
}
