package backend

import (
	"github.com/davyxu/cellnet"
	"landlord_go/proto"
	agentmodel "landlord_go/srv/agent/model"
)

func init() {
	proto.Handle_Agent_CloseClientACK = func(ev cellnet.Event) {
		msg := ev.Message().(*proto.CloseClientACK)
		if len(msg.ID) == 0 {
			agentmodel.VisitUser(func(user *agentmodel.User) bool {
				user.ClientSession.Close()
				return true
			})
		} else {
			for _, sesid := range msg.ID {
				if u := agentmodel.GetUser(sesid); u != nil {
					u.ClientSession.Close()
				}
			}
		}
	}

	proto.Handle_Agent_Default = func(ev cellnet.Event) {

	}
}