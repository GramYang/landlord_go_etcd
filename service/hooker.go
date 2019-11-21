package service

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/tcp"
	"landlord_go/discovery"
)

// 服务互联消息处理
type SvcEventHooker struct {
}

func (SvcEventHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	switch msg := inputEvent.Message().(type) {
	case *ServiceIdentifyACK:
		log.Tracef("receive ServiceIdentifyACK, name: %s, id: %s\n")
		if pre := GetRemoteService(msg.SvcID); pre == nil {

			// 添加连接上来的对方服务
			AddRemoteService(inputEvent.Session(), msg.SvcID, msg.SvcName)
		}
	case *cellnet.SessionConnected:

		ctx := inputEvent.Session().Peer().(cellnet.ContextSet)

		var sd *discovery.ServiceDesc
		if ctx.FetchContext("sd", &sd) {

			// 用Connector的名称（一般是ProcName）让远程知道自己是什么服务，用于网关等需要反向发送消息的标识
			inputEvent.Session().Send(&ServiceIdentifyACK{
				SvcName: GetSvcName(),
				SvcID:   GetLocalSvcID(),
			})

			AddRemoteService(inputEvent.Session(), sd.ID, sd.Name)
		} else {

			log.Errorf("Make sure call multi.AddPeer before peer.Start, peer: %s", inputEvent.Session().Peer().TypeName())
		}

	case *cellnet.SessionClosed:

		RemoveRemoteService(inputEvent.Session())
	}

	return inputEvent

}

func (SvcEventHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	return inputEvent
}

func init() {

	// 服务器间通讯协议
	proc.RegisterProcessor("tcp.svc", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {

		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(new(SvcEventHooker), new(tcp.MsgHooker)))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

	// 与客户端通信的处理器
	proc.RegisterProcessor("tcp.client", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {

		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(new(tcp.MsgHooker)))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}