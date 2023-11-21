package dgwrtc

import (
	dgctx "github.com/darwinOrg/go-common/context"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-web/wrapper"
	dgws "github.com/darwinOrg/go-websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GetRoomIdFunc func(c *gin.Context, ctx *dgctx.DgContext) (string, error)
type SignalingMessageCallbackFunc func(ctx *dgctx.DgContext, message *SignalingMessage) error

type SignalingConfig struct {
	RouterGroup              *gin.RouterGroup `binding:"required"`
	RelativePath             string
	Server                   *Server       `binding:"required"`
	GetRoomIdFunc            GetRoomIdFunc `binding:"required"`
	SignalingMessageCallback SignalingMessageCallbackFunc
}

func RegisterSignaling(config *SignalingConfig) {
	dgws.GetJson(&wrapper.RequestHolder[dgws.WebSocketMessage[SignalingMessage], error]{
		RouterGroup:  config.RouterGroup,
		RelativePath: config.RelativePath,
		BizHandler: func(_ *gin.Context, ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[SignalingMessage]) error {
			if wsm.MessageType == websocket.TextMessage {
				client := getClient(ctx)
				config.Server.handleSignalingMessage(ctx, client, wsm.MessageData)
				if config.SignalingMessageCallback != nil {
					err := config.SignalingMessageCallback(ctx, wsm.MessageData)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	}, func(c *gin.Context, ctx *dgctx.DgContext) error {
		roomID, err := config.GetRoomIdFunc(c, ctx)
		if err != nil {
			dglogger.Errorf(ctx, "GetRoomId error: %v", err)
			return err
		}
		SetRoomId(ctx, roomID)

		return nil
	}, func(ctx *dgctx.DgContext, conn *websocket.Conn) (*websocket.Conn, error) {
		client := &Client{
			id:     uuid.NewString(),
			conn:   conn,
			server: config.Server,
		}

		config.Server.mutex.Lock()
		config.Server.clients[client.id] = client
		setClient(ctx, client)
		config.Server.mutex.Unlock()

		return nil, nil
	}, dgws.DefaultIsEndFunc, func(ctx *dgctx.DgContext, conn *websocket.Conn, _ *websocket.Conn) error {
		client := getClient(ctx)

		// Clean up and close the connection
		config.Server.leaveRoom(ctx, client)

		config.Server.mutex.Lock()
		delete(config.Server.clients, client.id)
		config.Server.mutex.Unlock()

		return nil
	})
}
