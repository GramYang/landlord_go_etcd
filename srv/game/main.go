package main

import (
	_ "github.com/davyxu/cellnet/codec/gogopb"
	_ "github.com/davyxu/cellnet/codec/protoplus"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"landlord_go/basic"
	"landlord_go/basic/model"
	_ "landlord_go/srv/game/json"
	_ "landlord_go/srv/game/verify"
	"landlord_go/srv/hub/api"
	"landlord_go/srv/hub/status"
	"time"
)

func main() {
	basic.Init("game")
	basic.CreateAcceptor(bsmodel.ServiceParameter{
		SvcName:"game",
		NetProcName:"svc.backend",
		ListenAddr:":6792",
	})
	hubapi.ConnectToHub(func() {
		hubstatus.StartSendStatus("game_status", time.Second * 3, func() int {
			return 100
		})
	})
	basic.StartLoop(nil)
	basic.Exit()
}