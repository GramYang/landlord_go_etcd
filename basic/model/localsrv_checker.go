package bsmodel

import (
	"github.com/davyxu/cellnet"
)

//调用cellnet的PeerReadyChecker接口，检查一下cellnet是否运行
func IsAllReady() (ret bool) {
	ret = true
	VisitLocalService(func(srv cellnet.Peer) bool {
		if !srv.(cellnet.PeerReadyChecker).IsReady() {
			ret = false
			return false
		}
		return true
	})
	return
}

//在启动服务时，检查前面启动的服务是否运行正常，不正常的话log一下
func CheckReady() {
	if IsAllReady() {
		log.Infoln("all peers ready!")
	} else {
		log.Info("peers not all ready...")
	}
}