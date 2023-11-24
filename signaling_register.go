package dgwrtc

import (
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-common/utils"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-web/wrapper"
	dgws "github.com/darwinOrg/go-websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type GetBizIdFunc func(c *gin.Context, ctx *dgctx.DgContext) (int64, error)
type GetRoomIdFunc func(ctx *dgctx.DgContext, bizType string, bizId int64) (string, error)
type GetRoomClientFunc func(c *gin.Context, ctx *dgctx.DgContext) (*Client, error)
type StartSignalingCallbackFunc func(ctx *dgctx.DgContext, conn *websocket.Conn) error
type ClientLeaveRoomCallbackFunc func(ctx *dgctx.DgContext, client *Client) error

type SignalingConfig struct {
	RouterGroup              *gin.RouterGroup
	RelativePath             string
	BizType                  string
	GetBizId                 GetBizIdFunc
	GetRoomId                GetRoomIdFunc
	GetRoomClient            GetRoomClientFunc
	StartSignalingCallback   StartSignalingCallbackFunc
	SignalingMessageCallback dgws.WebSocketMessageCallback[[]byte]
	ClientLeaveRoomCallback  ClientLeaveRoomCallbackFunc
}

func RegisterSignaling(config *SignalingConfig) {
	server := newSignalingServer()

	dgws.GetBytes(&wrapper.RequestHolder[dgws.WebSocketMessage[[]byte], error]{
		RouterGroup:  config.RouterGroup,
		RelativePath: config.RelativePath,
		BizHandler: func(_ *gin.Context, ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[[]byte]) error {
			if wsm.MessageType == websocket.TextMessage {
				signalingMessage, err := utils.ConvertJsonBytesToBean[SignalingMessage](*wsm.MessageData)
				if err != nil {
					return err
				}

				server.handleSignalingMessage(ctx, signalingMessage, config.ClientLeaveRoomCallback)
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
		bizId, err := config.GetBizId(c, ctx)
		if err != nil {
			dglogger.Errorf(ctx, "GetBizId error: %v", err)
			return err
		}

		roomId, err := config.GetRoomId(ctx, config.BizType, bizId)
		if err != nil {
			dglogger.Errorf(ctx, "getRoomId error: %v", err)
			return err
		}
		setRoomId(ctx, roomId)

		client, err := config.GetRoomClient(c, ctx)
		if err != nil {
			dglogger.Errorf(ctx, "[bizId: %d, roomId: %s] GetRoomClient error: %v", bizId, roomId, err)
			return err
		}
		setClient(ctx, client)

		return nil
	}, func(ctx *dgctx.DgContext, conn *websocket.Conn) (*websocket.Conn, error) {
		client := getClient(ctx)
		client.conn = conn

		if config.StartSignalingCallback != nil {
			err := config.StartSignalingCallback(ctx, conn)
			if err != nil {
				dglogger.Errorf(ctx, "StartSignalingCallback error: %v", err)
				return nil, err
			}
		}

		return nil, nil
	}, dgws.DefaultIsEndFunc, func(ctx *dgctx.DgContext, conn *websocket.Conn, _ *websocket.Conn) error {
		// Clean up and close the connection
		server.leaveRoom(ctx, getClient(ctx), config.ClientLeaveRoomCallback)

		return nil
	})
}
