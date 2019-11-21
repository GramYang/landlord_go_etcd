package frontend

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	"landlord_go/basic/model"
	"landlord_go/discovery"
	"landlord_go/service"
	"landlord_go/srv/agent/model"
	"time"
)

//创建一个不用消息队列和usercallback的peer
func Start(param agentmodel.FrontendParameter) {
	clientListener := peer.NewGenericPeer(param.NetPeerType, param.SvcName, param.ListenAddr, nil)
	proc.BindProcessorHandler(clientListener, param.NetProcName, nil)
	if socketOpt, ok := clientListener.(cellnet.TCPSocketOption); ok {
		socketOpt.SetSocketBuffer(2048, 2048, true)
		socketOpt.SetSocketDeadline(time.Second * 40, time.Second * 20)
	}
	log.Tracef("create frontend acceptor without queue and usercallback, name: %s, peer: %s, proc: %s, addr: %s\n",
		param.SvcName, param.NetPeerType, param.NetProcName, param.ListenAddr)
	clientListener.Start()
	//保存acceptor的session
	agentmodel.FrontendSessionManager = clientListener.(peer.SessionManager)
	//注册acceptor
	service.Register(clientListener)
	bsmodel.AddLocalService(clientListener)
}

func Stop() {
	if agentmodel.FrontendSessionManager != nil {
		agentmodel.FrontendSessionManager.(cellnet.Peer).Stop()
		_ = discovery.Default.Deregister(agentmodel.AgentSvcID)
	}
}