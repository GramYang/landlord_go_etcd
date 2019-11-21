package main

import (
	_ "github.com/davyxu/cellnet/codec/gogopb"
	_ "github.com/davyxu/cellnet/codec/protoplus"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"landlord_go/basic"
	"landlord_go/basic/model"
	"landlord_go/proto"
	"landlord_go/srv/hub/api"
	"landlord_go/srv/hub/status"
	_ "landlord_go/srv/login/login"
)

func main() {
	basic.Init("login")
	basic.CreateAcceptor(bsmodel.ServiceParameter{
		SvcName:"login",
		NetPeerType:"tcp.Acceptor",
		NetProcName:"tcp.client",
		ListenAddr:":6790",
	})
	hubapi.ConnectToHub(func() {
		hubstatus.StartRecvStatus([]string{"game_status", "agent_status"}, &proto.Handle_Login_SvcStatusACK)
	})
	basic.StartLoop(nil)
	basic.Exit()
}