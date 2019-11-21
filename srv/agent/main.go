package main

import (
	_ "github.com/davyxu/cellnet/codec/gogopb"
	_ "github.com/davyxu/cellnet/codec/protoplus"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"landlord_go/basic"
	"landlord_go/basic/model"
	_ "landlord_go/proto"
	"landlord_go/service"
	_ "landlord_go/srv/agent/backend"
	"landlord_go/srv/agent/frontend"
	"landlord_go/srv/agent/heartbeat"
	"landlord_go/srv/agent/model"
	"landlord_go/srv/agent/routerule"
	"landlord_go/srv/hub/api"
	"landlord_go/srv/hub/status"
	"time"
)

func main() {
	basic.Init("agent")
	//_ = routerule.Download()
	routerule.GetRouteRule()
	heartbeat.StartCheck()
	agentmodel.AgentSvcID = service.GetLocalSvcID()
	basic.CreateConnector(bsmodel.ServiceParameter{
		SvcName:"game",
		NetProcName:"agent.backend",
	})
	frontend.Start(agentmodel.FrontendParameter{
		SvcName:     "agent",
		ListenAddr:  ":6791",
		NetPeerType: "tcp.Acceptor",
		NetProcName: "tcp.frontend",
	})
	hubapi.ConnectToHub(func() {
		hubstatus.StartSendStatus("agent_status", time.Second * 3, func() int {
			return agentmodel.FrontendSessionManager.SessionCount()
		})
	})
	basic.StartLoop(nil)
	frontend.Stop()
	basic.Exit()
}