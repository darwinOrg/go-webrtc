package room

import (
	daogext "github.com/darwinOrg/daog-ext"
	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-webrtc/room/dal"
	"github.com/google/uuid"
	"github.com/rolandhe/daog"
	"github.com/rolandhe/daog/ttypes"
	"time"
)

func GetOrCreateRoom(ctx *dgctx.DgContext, bizType string, bizId int64) (*dal.WebrtcRoom, error) {
	return daogext.WriteWithResult(ctx, func(tc *daog.TransContext) (*dal.WebrtcRoom, error) {
		matcher := daog.NewMatcher().Eq(dal.WebrtcRoomFields.BizType, bizType).Eq(dal.WebrtcRoomFields.BizId, bizId)
		rm, err := dal.WebrtcRoomDao.QueryOneMatcher(tc, matcher)
		if err != nil {
			return nil, err
		}

		if rm != nil {
			return rm, nil
		}

		now := ttypes.NormalDatetime(time.Now())
		rm = &dal.WebrtcRoom{
			BizType:    bizType,
			BizId:      bizId,
			RoomId:     uuid.NewString(),
			RoomName:   "",
			RoomStatus: 0,
			CreatedBy:  ctx.UserId,
			CreatedAt:  now,
			ModifiedBy: 0,
			ModifiedAt: now,
		}

		_, err = dal.WebrtcRoomDao.Insert(tc, rm)
		if err != nil {
			return nil, err
		}

		return rm, nil
	})
}

func GetOrCreateRoomClient(ctx *dgctx.DgContext, roomId string, clientId string, clientType string) (*dal.WebrtcRoomClient, error) {
	return daogext.WriteWithResult(ctx, func(tc *daog.TransContext) (*dal.WebrtcRoomClient, error) {
		matcher := daog.NewMatcher().Eq(dal.WebrtcRoomClientFields.RoomId, roomId).Eq(dal.WebrtcRoomClientFields.ClientId, clientId)
		rc, err := dal.WebrtcRoomClientDao.QueryOneMatcher(tc, matcher)
		if err != nil {
			return nil, err
		}

		if rc != nil {
			return rc, nil
		}

		now := ttypes.NormalDatetime(time.Now())
		rc = &dal.WebrtcRoomClient{
			RoomId:       roomId,
			ClientId:     clientId,
			ClientType:   clientType,
			ClientStatus: 0,
			UserId:       ctx.UserId,
			CreatedBy:    ctx.UserId,
			CreatedAt:    now,
			ModifiedBy:   ctx.UserId,
			ModifiedAt:   now,
		}

		_, err = dal.WebrtcRoomClientDao.Insert(tc, rc)
		if err != nil {
			return nil, err
		}

		return rc, nil
	})
}

func ClientLeaveRoom(ctx *dgctx.DgContext, roomId string, clientId string) error {
	return daogext.Write(ctx, func(tc *daog.TransContext) error {
		matcher := daog.NewMatcher().Eq(dal.WebrtcRoomClientFields.RoomId, roomId).Eq(dal.WebrtcRoomClientFields.ClientId, clientId)
		rc, err := dal.WebrtcRoomClientDao.QueryOneMatcher(tc, matcher)
		if err != nil {
			return err
		}

		if rc == nil {
			return nil
		}

		rc.ClientStatus = 1
		rc.ModifiedBy = ctx.UserId
		rc.ModifiedAt = ttypes.NormalDatetime(time.Now())

		_, err = dal.WebrtcRoomClientDao.Update(tc, rc)

		return err
	})
}
