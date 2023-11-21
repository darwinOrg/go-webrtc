package dgwrtc

import (
	"fmt"
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-web/wrapper"
	dgws "github.com/darwinOrg/go-websocket"
	"github.com/gorilla/websocket"
	"testing"
)

func TestSignaling(t *testing.T) {
	go StartStunServer(3478)
	engine := wrapper.DefaultEngine()
	dgws.InitWsConnLimit(10)

	RegisterSignaling(&SignalingConfig{
		RouterGroup:   engine.Group("/ws"),
		RelativePath:  "",
		GetRoomIdFunc: DefaultGetRoomIdFunc,
		StartSignalingCallback: func(ctx *dgctx.DgContext, conn *websocket.Conn) error {
			return nil
		},
		SignalingMessageCallback: func(ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[[]byte]) error {
			return nil
		},
	})

	engine.Run(fmt.Sprintf(":%d", 8080))
}
