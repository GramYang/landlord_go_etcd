package agentmodel

import (
	"github.com/davyxu/cellnet"
	"github.com/pkg/errors"
	"landlord_go/proto"
	"landlord_go/service"
	"time"
)

type Backend struct {
	SvcName string
	SvcID string
}

type User struct {
	ClientSession cellnet.Session
	Targets []*Backend
	LastPingTime time.Time
	CID proto.ClientID
}

// 广播到这个用户绑定的所有后台
func (self *User) BroadcastToBackends(msg interface{}) {
	for _, t := range self.Targets {
		backendSes := service.GetRemoteService(t.SvcID)
		if backendSes != nil {
			backendSes.Send(msg)
		}
	}
}

//转发到特定后台服务
func (self *User) TransmitToBackend(backendSvcid string, msgID int, msgData []byte) error {
	backendSes := service.GetRemoteService(backendSvcid)
	if backendSes == nil {
		return errors.New("backend not found")
	}
	backendSes.Send(&proto.TransmitACK{
		MsgID:uint32(msgID),
		MsgData:msgData,
		ClientID:self.CID.ID,
	})
	return nil
}

//用户绑定后台
func (self *User) SetBackend(svcName, svcID string) {
	for _, t := range self.Targets {
		if t.SvcName == svcName {
			t.SvcID = svcID
			return
		}
	}
	self.CID = proto.ClientID {
		ID:self.ClientSession.ID(),
		SvcID:AgentSvcID,
	}
	self.Targets = append(self.Targets,&Backend{
		SvcName:svcName,
		SvcID:svcID,
	})
}

//获取用户绑定后台svcid
func (self *User) GetBackend(svcName string) string {
	for _, t := range self.Targets {
		if t.SvcName == svcName {
			return t.SvcID
		}
	}
	return ""
}

func NewUser(clientSes cellnet.Session) *User {
	return &User{
		ClientSession:clientSes,
	}
}