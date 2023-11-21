package dgwrtc

import (
	dgctx "github.com/darwinOrg/go-common/context"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/gorilla/websocket"
	"sync"
)

var (
	RoomIdKey = "roomID"
	clientKey = "client"
)

type CommandType string

const (
	CommandJoin      CommandType = "join"
	CommandOffer     CommandType = "offer"
	CommandAnswer    CommandType = "answer"
	CommandCandidate CommandType = "candidate"
	CommandLeave     CommandType = "leave"
)

type SignalingMessage struct {
	Command CommandType            `json:"command"`
	Payload map[string]interface{} `json:"payload"`
}

// ICECandidate represents an ICE candidate
type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

// Room represents a room
type Room struct {
	id      string             // Room ID
	clients map[string]*Client // Clients in the room
	mutex   sync.RWMutex
}

// Client represents a connected client
type Client struct {
	id     string          // Client ID
	conn   *websocket.Conn // WebSocket connection for the client
	server *Server         // Reference to the signaling server
	room   *Room           // Room that the client belongs to
}

// Server represents the signaling server
type Server struct {
	clients map[string]*Client // All connected clients
	rooms   map[string]*Room   // All rooms
	mutex   sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[string]*Client),
		rooms:   make(map[string]*Room),
	}
}

// createRoom creates a new room with the given room ID
func (s *Server) createRoom(roomID string) *Room {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room := &Room{
		id:      roomID,
		clients: make(map[string]*Client),
	}
	s.rooms[roomID] = room
	return room
}

// joinRoom joins a client to the specified room
func (s *Server) joinRoom(roomID string, client *Client) *Room {
	s.mutex.Lock()
	room, ok := s.rooms[roomID]
	if !ok {
		s.mutex.Unlock()
		room = s.createRoom(roomID)
		s.mutex.Lock()
	}
	room.mutex.Lock()
	room.clients[client.id] = client
	client.room = room
	room.mutex.Unlock()
	s.mutex.Unlock()
	return room
}

// leaveRoom removes a client from its current room
func (s *Server) leaveRoom(ctx *dgctx.DgContext, client *Client) {
	s.mutex.Lock()
	if client.room != nil {
		room := client.room
		room.mutex.Lock()
		s.sendSignalingMessageToRoom(ctx, client.room, client, &SignalingMessage{
			Command: CommandLeave,
			Payload: map[string]interface{}{},
		})
		delete(room.clients, client.id)
		if len(room.clients) == 0 {
			delete(s.rooms, room.id)
		}
		room.mutex.Unlock()
		client.room = nil
	}
	s.mutex.Unlock()
}

// 发送信令消息给房间内的其他客户端
func (s *Server) sendSignalingMessageToRoom(ctx *dgctx.DgContext, room *Room, sender *Client, message *SignalingMessage) {
	for id, client := range room.clients {
		if client != sender {
			err := client.conn.WriteJSON(message)
			if err != nil {
				dglogger.Errorf(ctx, "Failed to send message to client, id:%v, err:%v", id, err)
				return
			}
		}
	}
}

// handleSignalingMessage handles signaling messages received from clients
func (s *Server) handleSignalingMessage(ctx *dgctx.DgContext, client *Client, message *SignalingMessage) {
	switch message.Command {
	case CommandJoin:
		roomID := getRoomId(ctx)
		if roomID == "" {
			dglogger.Error(ctx, "has no room ID")
			return
		}
		// Handle logic for client joining a room
		s.joinRoom(roomID, client)
	case CommandOffer, CommandAnswer, CommandCandidate:
		// Handle RTC signaling messages
		room := client.room
		if room != nil {
			s.sendSignalingMessageToRoom(ctx, room, client, message)
		}
	case CommandLeave:
		s.leaveRoom(ctx, client)
		return
	}
}

func setRoomId(ctx *dgctx.DgContext, roomID string) {
	ctx.SetExtraKeyValue(RoomIdKey, roomID)
}

func getRoomId(ctx *dgctx.DgContext) string {
	sessionId := ctx.GetExtraValue(RoomIdKey)
	if sessionId == nil {
		return ""
	}

	return sessionId.(string)
}

func setClient(ctx *dgctx.DgContext, client *Client) {
	ctx.SetExtraKeyValue(clientKey, client)
}

func getClient(ctx *dgctx.DgContext) *Client {
	server := ctx.GetExtraValue(clientKey)
	if server == nil {
		return nil
	}

	return server.(*Client)
}
