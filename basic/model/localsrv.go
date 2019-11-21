package bsmodel

import (
	"github.com/davyxu/cellnet"
	"sync"
)

//用于集中管理启动的本地srv
var (
	localServices []cellnet.Peer
	localServicesGuard sync.RWMutex
)

func AddLocalService(p cellnet.Peer) {
	localServicesGuard.Lock()
	defer localServicesGuard.Unlock()
	localServices = append(localServices, p)
}

func RemoveLocalService(p cellnet.Peer) {
	localServicesGuard.Lock()
	defer localServicesGuard.Unlock()
	for index, srv := range localServices {
		if srv == p {
			localServices = append(localServices[:index], localServices[index + 1:]...)
			break
		}
	}
}

func GetLocalService(srvName string) cellnet.Peer {
	localServicesGuard.RLock()
	defer localServicesGuard.RUnlock()
	for _, srv := range localServices {
		if prop, ok := srv.(cellnet.PeerProperty); ok && prop.Name() == srvName {
			return srv
		}
	}
	return nil
}

func VisitLocalService(callback func(cellnet.Peer) bool) {
	localServicesGuard.RLock()
	defer localServicesGuard.RUnlock()
	for _, srv := range localServices {
		if !callback(srv) {
			break
		}
	}
}

func StopAllService() {
	log.Traceln("close all peers...")
	localServicesGuard.RLock()
	defer localServicesGuard.RUnlock()
	for i := len(localServices) - 1; i >= 0; i-- {
		srv := localServices[i]
		srv.Stop()
	}
}