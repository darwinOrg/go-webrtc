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
		GetRoomIdFunc: func(c *gin.Context, ctx *dgctx.DgContext) (string, error) {
			return c.Query(RoomIdKey), nil
		},
		StartSignalingCallback: func(ctx *dgctx.DgContext, conn *websocket.Conn) error {
			return nil
		},
		SignalingMessageCallback: func(ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[[]byte]) error {
			return nil
		},
	})

	engine.Run(fmt.Sprintf(":%d", 8080))
}
