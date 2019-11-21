package hubstatus

import (
	"github.com/davyxu/cellnet/timer"
	"landlord_go/basic/model"
	"landlord_go/proto"
	"landlord_go/service"
	"landlord_go/srv/hub/api"
	"time"
)

func StartSendStatus(channelName string, updateInterval time.Duration, statusGetter func() int) {
	log.Tracef("start send status: %s\n", channelName)
	timer.NewLoop(bsmodel.Queue, updateInterval, func(loop *timer.Loop) {
		var ack proto.SvcStatusACK
		ack.SvcID = service.GetLocalSvcID()
		ack.UserCount = int32(statusGetter())
		hubapi.Publish(channelName, &ack)
	}, nil).Notify().Start()
}