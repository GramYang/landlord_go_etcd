package heartbeat

import (
	"github.com/davyxu/cellnet/timer"
	"landlord_go/discovery"
	"landlord_go/srv/agent/model"
	"strconv"
	"time"
)

const (
	heartBeatDuration_key = "config/agent/heartbeat_sec"
)

func StartCheck() {
	res, err := discovery.Default.GetValue(heartBeatDuration_key)
	if err != nil {
		return
	}
	heartBeatDuration, err1 := strconv.Atoi(res)
	if err1 != nil {
		return
	}
	if heartBeatDuration != 0 {
		//超时检查比心跳多5秒
		timeOutDur := time.Duration(heartBeatDuration + 5) * time.Second
		log.Tracef("Heartbeat duration: '%ds' \n", heartBeatDuration)
		timer.NewLoop(nil, timeOutDur, func(loop *timer.Loop) {
			now := time.Now()
			agentmodel.VisitUser(func(u *agentmodel.User) bool {
				if now.Sub(u.LastPingTime) > timeOutDur {
					log.Warnf("Close client due to heartbeat time out, id: %d", u.ClientSession.ID())
					u.ClientSession.Close()
				}
				return true
			})
		}, nil).Start()
	} else {
		//默认心跳时间为5秒
		timeOutDur := time.Duration(5) * time.Second
		log.Tracef("Heartbeat duration: '%ds' \n", 5)
		timer.NewLoop(nil, timeOutDur, func(loop *timer.Loop) {
			now := time.Now()
			agentmodel.VisitUser(func(u *agentmodel.User) bool {
				if now.Sub(u.LastPingTime) > timeOutDur {
					log.Warnf("Close client due to heartbeat time out, id: %d", u.ClientSession.ID())
					u.ClientSession.Close()
				}
				return true
			})
		}, nil).Start()
	}
}