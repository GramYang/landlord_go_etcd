package main

import (
	_ "github.com/davyxu/cellnet/codec/gogopb"
	_ "github.com/davyxu/cellnet/codec/protoplus"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"landlord_go/basic"
	"landlord_go/basic/model"
	_ "landlord_go/srv/hub/subscribe"
)

func main() {
	basic.Init("hub")
	basic.CreateAcceptor(bsmodel.ServiceParameter{
		SvcName:"hub",
		NetProcName:"tcp.svc",
		ListenAddr:":6789",
	})
	basic.StartLoop(nil)
	basic.Exit()
}