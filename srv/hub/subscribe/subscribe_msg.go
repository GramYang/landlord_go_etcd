package subscribe

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/relay"
	"landlord_go/proto"
	"landlord_go/srv/hub/model"
)

func init() {
	relay.SetBroadcaster(func(event *relay.RecvMsgEvent) {
		if channelName := event.PassThroughAsString(); channelName != "" {
			hubmodel.VisitSubscriber(channelName, func(ses cellnet.Session) bool {
				_ = relay.Relay(ses, event.Message(), channelName)
				return true
			})
		}
	})

	proto.Handle_Hub_SubscribeChannelREQ = func(ev cellnet.Event) {
		msg := ev.Message().(*proto.SubscribeChannelREQ)
		hubmodel.AddSubscriber(msg.Channel, ev.Session())
		log.Infof("channel add: '%s', sesid: %d", msg.Channel, ev.Session().ID())
		ev.Session().Send(&proto.SubscribeChannelACK{
			Channel:msg.Channel,
		})
	}

	proto.Handle_Hub_Default = func(ev cellnet.Event) {
		switch ev.Message().(type) {
		case *cellnet.SessionClosed:
			hubmodel.RemoveSubscriber(ev.Session(), func(chanName string) {
				log.Infof("channel remove: '%s', sesid: %d", chanName, ev.Session().ID())
			})
		}
	}
}