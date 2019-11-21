package frontend

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/gorillaws"
	"github.com/davyxu/cellnet/proc/tcp"
	"landlord_go/proto"
	"landlord_go/srv/agent/model"
	"time"
)

var (
	PingACKMsgID = cellnet.MessageMetaByFullName("proto.PingACK").ID
	VerifyREQMsgID = cellnet.MessageMetaByFullName("proto.VerifyREQ").ID
)

func ProcFrontendPacket(msgID int, msgData []byte, ses cellnet.Session) (msg interface{}, err error) {
	switch int(msgID) {
	case PingACKMsgID, VerifyREQMsgID:
		msg, _, err = codec.DecodeMessage(msgID, msgData)
		if err != nil {
			return nil, err
		}
		switch userMsg := msg.(type) {
		//心跳
		case *proto.PingACK:
			log.Traceln("receive PingACK")
			u := agentmodel.SessionToUser(ses)
			if u != nil {
				u.LastPingTime = time.Now()
				ses.Send(&proto.PingACK{})
			} else {
				ses.Close()
			}
		// 第一个到网关的消息
		case *proto.VerifyREQ:
			log.Tracef("receive VerifyREQ, svcName: %s\n", userMsg.GameSvcID)
			u, err := bindClientToBackend(userMsg.GameSvcID, ses.ID())
			if err == nil {
				_ = u.TransmitToBackend(userMsg.GameSvcID, msgID, msgData)
			} else {
				ses.Close()
				log.Errorln("bindClientToBackend", err)
			}
		}
	default:
		rule := agentmodel.GetRuleByMsgID(msgID)
		if rule == nil {
			return nil, fmt.Errorf("Message not in route table, msgid: %d", msgID)
		}
		u:= agentmodel.SessionToUser(ses)
		if u != nil {
			if err = u.TransmitToBackend(u.GetBackend(rule.SvcName), msgID, msgData); err != nil {
				log.Warnf("TransmitToBackend %s, msg: '%s' svc: %s", err, rule.MsgName, rule.SvcName)
			}
		} else {

		}
	}
	return
}

type FrontendEventHooker struct {
}

func (FrontendEventHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	switch inputEvent.Message().(type) {
	case *cellnet.SessionAccepted:
	case *cellnet.SessionClosed:
		u := agentmodel.SessionToUser(inputEvent.Session())
		if u != nil {
			u.BroadcastToBackends(&proto.ClientClosedACK{
				ID:&proto.ClientID{
					ID:inputEvent.Session().ID(),
					SvcID:agentmodel.AgentSvcID,
				},
			})
		}
	}
	return inputEvent
}

func (FrontendEventHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	return inputEvent
}

func init() {
	proc.RegisterProcessor("tcp.frontend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(directTCPTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(tcp.MsgHooker),
			new(FrontendEventHooker),
		))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

	proc.RegisterProcessor("ws.frontend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(directWSMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(gorillaws.MsgHooker),
			new(FrontendEventHooker),
		))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}