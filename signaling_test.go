package dgwrtc

import (
	"fmt"
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-monitor"
	"github.com/darwinOrg/go-web/wrapper"
	dgws "github.com/darwinOrg/go-websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"testing"
)

func TestSignaling(t *testing.T) {
	monitor.Start("signaling", 19002)
	go StartStunServer(19302)
	engine := wrapper.DefaultEngine()
	dgws.InitWsConnLimit(10)

	RegisterSignaling(&SignalingConfig{
		RouterGroup:  engine.Group("/ws"),
		RelativePath: "",
		BizType:      "test",
		GetBizId: func(c *gin.Context, ctx *dgctx.DgContext) (int64, error) {
			return 1, nil
		},
		GetRoomId:     DefaultGetRoomId,
		GetRoomClient: DefaultGetRoomClient,
		StartSignalingCallback: func(ctx *dgctx.DgContext, conn *websocket.Conn) error {
			return nil
		},
		SignalingMessageCallback: func(ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[[]byte]) error {
			return nil
		},
		ClientLeaveRoomCallback: func(ctx *dgctx.DgContext, client *Client) error {
			return nil
		},
	})

	engine.Run(fmt.Sprintf(":%d", 8080))
}
