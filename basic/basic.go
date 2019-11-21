package basic

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	model "landlord_go/basic/model"
	"landlord_go/discovery"
	"landlord_go/proto"
	"landlord_go/service"
	"time"
)

//初始化框架
func Init(svcName string) {
	model.Queue = cellnet.NewEventQueue()
	model.Queue.StartLoop()
	service.Init(svcName)
	service.ConnectDiscovery()
}

func CreateAcceptor(param model.ServiceParameter) cellnet.Peer {
	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Acceptor"
	}
	var q cellnet.EventQueue
	if !param.NoQueue { //所有的peer都共用一个队列
		q = model.Queue
	}
	p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, param.ListenAddr, q)
	msgFunc := proto.GetMessageHandler(param.SvcName)
	//tcp.svc
	proc.BindProcessorHandler(p, param.NetProcName, func(ev cellnet.Event) {
		if msgFunc != nil {
			msgFunc(ev)
		}
	})
	if opt, ok := p.(cellnet.TCPSocketOption); ok {
		opt.SetSocketBuffer(2048, 2048, true)
	}
	log.Tracef("create acceptor, name: %s, peer: %s, proc: %s, addr: %s\n",
		param.SvcName, param.NetPeerType, param.NetProcName, param.ListenAddr)
	model.AddLocalService(p)
	p.Start()
	service.Register(p)
	return p
}

func CreateConnector(param model.ServiceParameter) cellnet.Peer{
	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Connector"
	}
	msgFunc := proto.GetMessageHandler(service.GetSvcName()) //从service获取svcName，确保service的Init已调用
	opt := service.DiscoveryOption{
		MaxCount:param.MaxConnCount,
	}
	var q cellnet.EventQueue
	if !param.NoQueue { //所有的peer都共用一个队列
		q = model.Queue
	}
	mp := service.DiscoveryService(param.SvcName, opt, func(multiPeer service.MultiPeer, sd *discovery.ServiceDesc) {
		p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, sd.Address(), q)
		proc.BindProcessorHandler(p, param.NetProcName, func(ev cellnet.Event) {
			if msgFunc != nil {
				msgFunc(ev)
			}
		})
		if opt, ok := p.(cellnet.TCPSocketOption); ok {
			opt.SetSocketBuffer(2048, 2048, true)
		}
		p.(cellnet.TCPConnector).SetReconnectDuration(time.Second * 3)
		log.Tracef("create connector, name: %s, peer: %s, proc: %s\n",
			param.SvcName, param.NetPeerType, param.NetProcName)
		multiPeer.AddPeer(sd, p)
		p.Start()
	})
	mp.(service.MultiPeer).SetContext("multi", param)
	model.AddLocalService(mp)
	return mp
}

//等待退出信号
func StartLoop(onReady func()) {
	model.CheckReady()
	if onReady != nil {
		cellnet.QueuedCall(model.Queue, onReady)
	}
	service.WaitExitSignal()
}

func GetRemoteServiceWANAddress(svcName, svcid string) string {
	result := service.QueryService(svcName, service.Filter_MatchSvcID(svcid))
	if result == nil {
		return ""
	}
	desc := result.(*discovery.ServiceDesc)
	wanAddr := desc.GetMeta("WANAddress")
	if wanAddr != "" {
		return wanAddr
	}
	//没有外网地址就返回内网ip
	return desc.Address()
}

//退出处理
func Exit() {
	_ = service.Close()
	model.StopAllService()
}