package verify

import (
	"github.com/davyxu/cellnet"
	"landlord_go/proto"
	"landlord_go/service"
	"landlord_go/srv/agent/api"
)

func init() {
	proto.Handle_Game_VerifyREQ = agentapi.HandleBackendMessage(func(ev cellnet.Event, cid proto.ClientID) {
		msg := ev.Message().(*proto.VerifyREQ)
		log.Infof("verify: %+v \n", msg.GameToken)
		service.Reply(ev, &proto.VerifyACK{Result:proto.ResultCode_NoError})
	})
}