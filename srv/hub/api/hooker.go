package hubapi

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/tcp"
	"landlord_go/basic/model"
	"landlord_go/service"
	"landlord_go/srv/hub/model"
)

type subscriberHooker struct {}

func (subscriberHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	switch inputEvent.Message().(type) {
	case *cellnet.SessionConnected:
		hubmodel.HubSession = inputEvent.Session()
		Subscribe(service.GetSvcName())
		Subscribe(service.GetLocalSvcID())
		if hubmodel.OnHubReady != nil {
			cellnet.QueuedCall(bsmodel.Queue, hubmodel.OnHubReady)
		}
	}
	return inputEvent
}

func (subscriberHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	return inputEvent
}

func init() {
	proc.RegisterProcessor("tcp.hub", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(new(subscriberHooker), new(tcp.MsgHooker)))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}