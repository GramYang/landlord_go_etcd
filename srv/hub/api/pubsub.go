package hubapi

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/relay"
	"landlord_go/basic"
	"landlord_go/basic/model"
	"landlord_go/proto"
	"landlord_go/srv/hub/model"
)

func ConnectToHub(hubReady func()) cellnet.Peer{
	hubmodel.OnHubReady = hubReady
	return basic.CreateConnector(bsmodel.ServiceParameter{
		SvcName:"hub",
		NetProcName:"tcp.hub",
	})
}

func Subscribe(channel string) {
	if hubmodel.HubSession == nil {
		log.Errorf("hub session not ready, channel: %s", channel)
		return
	}
	hubmodel.HubSession.Send(&proto.SubscribeChannelREQ{
		Channel:channel,
	})
	log.Tracef("subscribe the channel: %s\n", channel)
}

func Publish(channel string, msg interface{}) {
	if hubmodel.HubSession == nil {
		log.Errorf("hub session not ready, channel: %s", channel)
		return
	}
	_ = relay.Relay(hubmodel.HubSession, msg, channel)
}