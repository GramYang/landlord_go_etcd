package backend

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/tcp"
	"landlord_go/proto"
	"landlord_go/service"
	"landlord_go/srv/agent/model"
)

type BackendMsgHooker struct {
}

func (BackendMsgHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	switch incomingMsg := inputEvent.Message().(type) {
	case *proto.TransmitACK:
		log.Tracef("receive TransmitACK from client, msgId: %s, clientId: %s\n", incomingMsg.MsgID, incomingMsg.ClientID)
		userMsg, _, err := codec.DecodeMessage(int(incomingMsg.MsgID), incomingMsg.MsgData)
		if err != nil {
			log.Warnf("Backend msg decode failed, %s, msgid: %d", err.Error(), incomingMsg.MsgID)
			return nil
		}
		ev := &RecvMsgEvent{
			Ses:      inputEvent.Session(),
			Msg:      userMsg,
			ClientID: incomingMsg.ClientID,
		}

		outputEvent = ev
	default:
		outputEvent = inputEvent
	}
	return
}

func (BackendMsgHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	return inputEvent
}

type broadcasterHooker struct {
}

// 来自后台服务器的消息
func (broadcasterHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	switch incomingMsg := inputEvent.Message().(type) {
	case *proto.TransmitACK:
		log.Tracef("receive TransmitACK from backend, msgid: %s\n", incomingMsg.MsgID)
		rawPkt := &cellnet.RawPacket{
			MsgData: incomingMsg.MsgData,
			MsgID:   int(incomingMsg.MsgID),
		}
		// 单发
		if incomingMsg.ClientID != 0 {
			clientSes := agentmodel.GetClientSession(incomingMsg.ClientID)
			if clientSes != nil {
				clientSes.Send(rawPkt)
			}
			// 广播
		} else if incomingMsg.ClientIDList != nil {
			for _, cid := range incomingMsg.ClientIDList {
				clientSes := agentmodel.GetClientSession(cid)
				if clientSes != nil {
					clientSes.Send(rawPkt)
				}
			}
		} else if incomingMsg.All {
			agentmodel.FrontendSessionManager.VisitSession(func(clientSes cellnet.Session) bool {
				clientSes.Send(rawPkt)
				return true
			})
		}
		// 本事件已经处理, 不再后传
		return nil
	}
	return inputEvent
}

// 发送给后台服务器
func (broadcasterHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	return inputEvent
}

func init() {
	proc.RegisterProcessor("svc.backend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(service.SvcEventHooker), // 服务互联处理
			new(BackendMsgHooker),       // 网关消息处理
			new(tcp.MsgHooker)))         // tcp基础消息处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

	proc.RegisterProcessor("agent.backend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(service.SvcEventHooker), // 服务互联处理
			new(broadcasterHooker),      // 网关消息处理
			new(tcp.MsgHooker)))         // tcp基础消息处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}