package login

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/util"
	"landlord_go/basic"
	"landlord_go/proto"
	"landlord_go/service"
	"landlord_go/srv/hub/status"
	"landlord_go/srv/login/database"
	"strings"
)

const (
	loginSuccess = 200
	loginWrongUserName = 201
	loginWrongPassword = 202
)

func init() {
	proto.Handle_Login_LoginREQ = func(ev cellnet.Event) {
		var ack  = &proto.LoginACK{}
		loginReq := ev.Message().(*proto.LoginREQ)
		agentSvcID := hubstatus.SelectServiceByLowUserCount("agent", "dev", false)
		if agentSvcID == "" {
			ack.Result = proto.ResultCode_AgentNotFound
			service.Reply(ev, &ack)
			return
		}
		agentWAN := basic.GetRemoteServiceWANAddress("agent", agentSvcID)
		host, port, err := util.SpliteAddress(agentWAN)
		if err != nil {
			log.Errorf("invalid address: '%s' %s\n", agentWAN, err.Error())
			ack.Result = proto.ResultCode_AgentAddressError
			service.Reply(ev, &ack)
			return
		}
		ack.Server = &proto.ServerInfo{
			IP:host,
			Port:int32(port),
		}
		ack.GameSvcID = hubstatus.SelectServiceByLowUserCount("game", "dev", false)
		if ack.GameSvcID == "" {
			ack.Result = proto.ResultCode_GameNotFound
			service.Reply(ev, &ack)
			return
		}
		ack.Result = proto.ResultCode_NoError
		var username string
		var password string
		token := string(loginReq.TokenReq)
		log.Tracef("receive LoginREQ, token: %s\n", token)
		if strings.Contains(token, ":") {
			tmp := strings.Split(token, ":")
			username = tmp[0]
			password = tmp[1]
		}
		userPassword := database.GetUserPassword(username)
		if userPassword == "" {
			ack.TokenAck = loginWrongUserName
			service.Reply(ev, &ack)
			return
		}
		if password == userPassword {
			ack.TokenAck = loginSuccess
		} else {
			ack.TokenAck = loginWrongPassword
		}
		service.Reply(ev, &ack)
	}
}