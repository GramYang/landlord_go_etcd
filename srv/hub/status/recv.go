package hubstatus

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/timer"
	"landlord_go/basic/model"
	"landlord_go/proto"
	"landlord_go/srv/hub/api"
	"landlord_go/srv/hub/model"
	"time"
)

var (
	recvLoop *timer.Loop
)

const (
	statusUpdateTimeout = time.Second * 3
)

func StartRecvStatus(channelNames []string, srvStatusHandler *func(ev cellnet.Event)) {
	log.Tracef("start receive status: %s\n", channelNames)
	for _, channelName := range channelNames {
		hubapi.Subscribe(channelName)
	}
	*srvStatusHandler = func(ev cellnet.Event) {
		msg := ev.Message().(*proto.SvcStatusACK)
		hubmodel.UpdateStatus(&hubmodel.Status{
			UserCount:msg.UserCount,
			SvcID:msg.SvcID,
		})
	}
	if recvLoop == nil {
		recvLoop = timer.NewLoop(bsmodel.Queue, statusUpdateTimeout, func(loop *timer.Loop) {
			now := time.Now()
			var timeoutSvcID []string
			hubmodel.VisitStatus(func(status *hubmodel.Status) bool {
				if now.Sub(status.LastUpdate) > statusUpdateTimeout {
					timeoutSvcID = append(timeoutSvcID, status.SvcID)
				}
				return true
			})
			for _, svcid := range timeoutSvcID {
				hubmodel.RemoveStatus(svcid)
			}
		}, nil)
		recvLoop.Notify().Start()
	}
}