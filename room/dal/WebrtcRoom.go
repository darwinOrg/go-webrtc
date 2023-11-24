package dal

import (
	"github.com/rolandhe/daog"
	"github.com/rolandhe/daog/ttypes"
)

var WebrtcRoomFields = struct {
	Id         string
	BizType    string
	BizId      string
	RoomId     string
	RoomName   string
	RoomStatus string
	CreatedBy  string
	CreatedAt  string
	ModifiedBy string
	ModifiedAt string
}{
	"id",
	"biz_type",
	"biz_id",
	"room_id",
	"room_name",
	"room_status",
	"created_by",
	"created_at",
	"modified_by",
	"modified_at",
}

var WebrtcRoomMeta = &daog.TableMeta[WebrtcRoom]{
	Table: "webrtc_room",
	Columns: []string{
		"id",
		"biz_type",
		"biz_id",
		"room_id",
		"room_name",
		"room_status",
		"created_by",
		"created_at",
		"modified_by",
		"modified_at",
	},
	AutoColumn: "id",
	LookupFieldFunc: func(columnName string, ins *WebrtcRoom, point bool) any {
		if "id" == columnName {
			if point {
				return &ins.Id
			}
			return ins.Id
		}
		if "biz_type" == columnName {
			if point {
				return &ins.BizType
			}
			return ins.BizType
		}
		if "biz_id" == columnName {
			if point {
				return &ins.BizId
			}
			return ins.BizId
		}
		if "room_id" == columnName {
			if point {
				return &ins.RoomId
			}
			return ins.RoomId
		}
		if "room_name" == columnName {
			if point {
				return &ins.RoomName
			}
			return ins.RoomName
		}
		if "room_status" == columnName {
			if point {
				return &ins.RoomStatus
			}
			return ins.RoomStatus
		}
		if "created_by" == columnName {
			if point {
				return &ins.CreatedBy
			}
			return ins.CreatedBy
		}
		if "created_at" == columnName {
			if point {
				return &ins.CreatedAt
			}
			return ins.CreatedAt
		}
		if "modified_by" == columnName {
			if point {
				return &ins.ModifiedBy
			}
			return ins.ModifiedBy
		}
		if "modified_at" == columnName {
			if point {
				return &ins.ModifiedAt
			}
			return ins.ModifiedAt
		}

		return nil
	},
}

var WebrtcRoomDao daog.QuickDao[WebrtcRoom] = &struct {
	daog.QuickDao[WebrtcRoom]
}{
	daog.NewBaseQuickDao(WebrtcRoomMeta),
}

type WebrtcRoom struct {
	Id         int64
	BizType    string
	BizId      int64
	RoomId     string
	RoomName   string
	RoomStatus int32
	CreatedBy  int64
	CreatedAt  ttypes.NormalDatetime
	ModifiedBy int64
	ModifiedAt ttypes.NormalDatetime
}
