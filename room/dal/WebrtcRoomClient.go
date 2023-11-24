package dal

import (
	"github.com/rolandhe/daog"
	"github.com/rolandhe/daog/ttypes"
)

var WebrtcRoomClientFields = struct {
	Id           string
	RoomId       string
	ClientId     string
	ClientType   string
	ClientStatus string
	UserId       string
	CreatedBy    string
	CreatedAt    string
	ModifiedBy   string
	ModifiedAt   string
}{
	"id",
	"room_id",
	"client_id",
	"client_type",
	"client_status",
	"user_id",
	"created_by",
	"created_at",
	"modified_by",
	"modified_at",
}

var WebrtcRoomClientMeta = &daog.TableMeta[WebrtcRoomClient]{
	Table: "webrtc_room_client",
	Columns: []string{
		"id",
		"room_id",
		"client_id",
		"client_type",
		"client_status",
		"user_id",
		"created_by",
		"created_at",
		"modified_by",
		"modified_at",
	},
	AutoColumn: "id",
	LookupFieldFunc: func(columnName string, ins *WebrtcRoomClient, point bool) any {
		if "id" == columnName {
			if point {
				return &ins.Id
			}
			return ins.Id
		}
		if "room_id" == columnName {
			if point {
				return &ins.RoomId
			}
			return ins.RoomId
		}
		if "client_id" == columnName {
			if point {
				return &ins.ClientId
			}
			return ins.ClientId
		}
		if "client_type" == columnName {
			if point {
				return &ins.ClientType
			}
			return ins.ClientType
		}
		if "client_status" == columnName {
			if point {
				return &ins.ClientStatus
			}
			return ins.ClientStatus
		}
		if "user_id" == columnName {
			if point {
				return &ins.UserId
			}
			return ins.UserId
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

var WebrtcRoomClientDao daog.QuickDao[WebrtcRoomClient] = &struct {
	daog.QuickDao[WebrtcRoomClient]
}{
	daog.NewBaseQuickDao(WebrtcRoomClientMeta),
}

type WebrtcRoomClient struct {
	Id           int64
	RoomId       string
	ClientId     string
	ClientType   string
	ClientStatus int32
	UserId       int64
	CreatedBy    int64
	CreatedAt    ttypes.NormalDatetime
	ModifiedBy   int64
	ModifiedAt   ttypes.NormalDatetime
}
