package dgwrtc

import (
	dgctx "github.com/darwinOrg/go-common/context"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	"github.com/darwinOrg/go-common/utils"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-web/wrapper"
	"github.com/darwinOrg/go-webrtc/room"
	dgws "github.com/darwinOrg/go-websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	ClientIdKey   = "clientId"
	ClientTypeKey = "clientType"
)

type GetBizIdFunc func(c *gin.Context, ctx *dgctx.DgContext) (int64, error)
type GetRoomIdFunc func(ctx *dgctx.DgContext, bizType string, bizId int64) (string, error)
type GetRoomClientFunc func(c *gin.Context, ctx *dgctx.DgContext, roomId string) (*Client, error)
type StartSignalingCallbackFunc func(ctx *dgctx.DgContext, conn *websocket.Conn) error

func DefaultGetRoomId(ctx *dgctx.DgContext, bizType string, bizId int64) (string, error) {
	rm, err := room.GetOrCreateRoom(ctx, bizType, bizId)
	if err != nil {
		return "", err
	}

	return rm.RoomId, nil
}

func DefaultGetRoomClient(c *gin.Context, ctx *dgctx.DgContext, roomId string) (*Client, error) {
	clientId := c.Query(ClientIdKey)
	clientType := c.Query(ClientTypeKey)
	if clientId == "" || clientType == "" {
		return nil, dgerr.ARGUMENT_NOT_VALID
	}

	rc, err := room.GetOrCreateRoomClient(ctx, roomId, clientId, clientType)
	if err != nil {
		return nil, err
	}

	return &Client{
		id:  rc.ClientId,
		tye: rc.ClientType,
	}, nil
}

func DefaultClientLeaveRoomCallback(ctx *dgctx.DgContext, client *Client) error {
	return room.ClientLeaveRoom(ctx, getRoomId(ctx), client.id)
}

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

	if config.GetRoomId == nil {
		config.GetRoomId = DefaultGetRoomId
	}
	if config.GetRoomClient == nil {
		config.GetRoomClient = DefaultGetRoomClient
	}
	if config.ClientLeaveRoomCallback == nil {
		config.ClientLeaveRoomCallback = DefaultClientLeaveRoomCallback
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

		client, err := config.GetRoomClient(c, ctx, roomId)
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
