package dgwrtc

import (
	"fmt"
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-web/wrapper"
	dgws "github.com/darwinOrg/go-websocket"
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
		SignalingMessageCallback: func(ctx *dgctx.DgContext, wsm *dgws.WebSocketMessage[[]byte]) error {
			return nil
		},
	})

	engine.Run(fmt.Sprintf(":%d", 8080))
}
