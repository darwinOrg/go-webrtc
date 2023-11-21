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
type StartSignalingCallbackFunc func(ctx *dgctx.DgContext, conn *websocket.Conn) error

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
	GetRoomIdFunc            GetRoomIdFunc
	StartSignalingCallback   StartSignalingCallbackFunc
	SignalingMessageCallback dgws.WebSocketMessageCallback[[]byte]
}

func RegisterSignaling(config *SignalingConfig) {
	server := NewServer()

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
				server.handleSignalingMessage(ctx, client, signalingMessage)
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
			dglogger.Errorf(ctx, "GetRoomId error: %v", err)
			return err
		}
		setRoomId(ctx, roomID)

		return nil
	}, func(ctx *dgctx.DgContext, conn *websocket.Conn) (*websocket.Conn, error) {
		client := &Client{
			id:     uuid.NewString(),
			conn:   conn,
			server: server,
		}

		server.mutex.Lock()
		server.clients[client.id] = client
		setClient(ctx, client)
		server.mutex.Unlock()

		if config.StartSignalingCallback != nil {
			err := config.StartSignalingCallback(ctx, conn)
			if err != nil {
				dglogger.Errorf(ctx, "StartSignalingCallback error: %v", err)
				return nil, err
			}
		}

		return nil, nil
	}, dgws.DefaultIsEndFunc, func(ctx *dgctx.DgContext, conn *websocket.Conn, _ *websocket.Conn) error {
		client := getClient(ctx)

		// Clean up and close the connection
		server.leaveRoom(ctx, client)

		server.mutex.Lock()
		delete(server.clients, client.id)
		server.mutex.Unlock()

		return nil
	})
}
