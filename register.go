package dgwrtc

import (
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-common/utils"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-web/wrapper"
	dgws "github.com/darwinOrg/go-websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GetRoomIdFunc func(c *gin.Context, ctx *dgctx.DgContext) (string, error)

var DefaultGetRoomIdFunc = func(c *gin.Context, ctx *dgctx.DgContext) (string, error) {
	roomID := c.Query(RoomIdKey)
	if roomID == "" {
		roomID = uuid.NewString()
	}

	return roomID, nil
}

type SignalingConfig struct {
	RouterGroup              *gin.RouterGroup
	RelativePath             string
	Server                   *Server
	GetRoomIdFunc            GetRoomIdFunc
	SignalingMessageCallback dgws.WebSocketMessageCallback[[]byte]
}

func RegisterSignaling(config *SignalingConfig) {
	if config.Server == nil {
		config.Server = NewServer()
	}

	if config.GetRoomIdFunc == nil {
		config.GetRoomIdFunc = DefaultGetRoomIdFunc
	}

	dgws.GetBytes(&wrapper.RequestHolder[dgws.WebSocketMessage[[]byte], error]{
		RouterGroup:  config.RouterGroup,
		RelativePath: config.RelativePath,
		BizHandler: func(_ *gin.Context, ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[[]byte]) error {
			if wsm.MessageType == websocket.TextMessage {
				signalingMessage, err := utils.ConvertJsonBytesToBean[SignalingMessage](*wsm.MessageData)
				if err != nil {
					return err
				}

				client := getClient(ctx)
				config.Server.handleSignalingMessage(ctx, client, signalingMessage)
			}

			if config.SignalingMessageCallback != nil {
				err := config.SignalingMessageCallback(ctx, wsm)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}, func(c *gin.Context, ctx *dgctx.DgContext) error {
		roomID, err := config.GetRoomIdFunc(c, ctx)
		if err != nil {
			dglogger.Errorf(ctx, "getRoomId error: %v", err)
			return err
		}
		setRoomId(ctx, roomID)

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
