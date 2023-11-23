package dgwrtc

import (
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-common/result"
	"github.com/darwinOrg/go-web/wrapper"
	"github.com/gin-gonic/gin"
)

func RegisterCredentials(server *TurnServer, routerGroup *gin.RouterGroup, relativePath string) {
	if server == nil {
		return
	}

	wrapper.Get(&wrapper.RequestHolder[result.Void, *result.Result[*UserCredentials]]{
		RouterGroup:  routerGroup,
		RelativePath: relativePath,
		NonLogin:     true,
		BizHandler: func(c *gin.Context, dc *dgctx.DgContext, _ *result.Void) *result.Result[*UserCredentials] {
			credentials, err := server.GenerateLongTermCredentials()
			if err != nil {
				return result.FailByError[*UserCredentials](err)
			}

			return result.Success[*UserCredentials](credentials)
		},
	})
}
