package dgwrtc

import (
	"fmt"
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-monitor"
	"github.com/darwinOrg/go-web/wrapper"
	dgws "github.com/darwinOrg/go-websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"testing"
)

func TestSignaling(t *testing.T) {
	monitor.Start("wrtc", 19002)
	go StartStunServer(3478)
	engine := wrapper.DefaultEngine()
	dgws.InitWsConnLimit(10)

	s := &Server{
		clients: make(map[string]*Client),
		rooms:   make(map[string]*Room),
	}

	dgws.GetJson(&wrapper.RequestHolder[dgws.WebSocketMessage[SignalingMessage], error]{
		RouterGroup: engine.Group("/ws"),
		BizHandler: func(_ *gin.Context, ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[SignalingMessage]) error {
			if wsm.MessageType == websocket.TextMessage {
				client := ctx.GetExtraValue("client").(*Client)
				s.handleSignalingMessage(ctx, client, wsm.MessageData)
			}

			return nil
		},
	}, nil, func(ctx *dgctx.DgContext, conn *websocket.Conn) (*websocket.Conn, error) {
		client := &Client{
			id:     uuid.NewString(),
			conn:   conn,
			server: s,
		}

		s.mutex.Lock()
		s.clients[client.id] = client
		ctx.SetExtraKeyValue("client", client)
		s.mutex.Unlock()

		return nil, nil
	}, dgws.DefaultIsEndFunc, func(ctx *dgctx.DgContext, conn *websocket.Conn, _ *websocket.Conn) error {
		client := ctx.GetExtraValue("client").(*Client)

		// Clean up and close the connection
		s.leaveRoom(ctx, client)

		s.mutex.Lock()
		delete(s.clients, client.id)
		s.mutex.Unlock()

		return nil
	})

	engine.Run(fmt.Sprintf(":%d", 8080))
}
