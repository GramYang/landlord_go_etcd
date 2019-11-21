package service

import (
	"github.com/davyxu/cellnet"
	"sync"
)

type RemoteServiceContext struct {
	Name string
	SvcID string
}

type NotifyFunc func(ctx *RemoteServiceContext, ses cellnet.Session)

var (
	//每个服务都持有一个connBySvcID，SvcID标识对面的服务
	connBySvcID = map[string]cellnet.Session{}
	connBySvcNameGuard sync.RWMutex
	//对connBySvcID进行删除操作时调用
	removeNotify NotifyFunc
)

func AddRemoteService(ses cellnet.Session, svcid, name string) {
	connBySvcNameGuard.Lock()
	ses.(cellnet.ContextSet).SetContext("ctx", &RemoteServiceContext{Name:name, SvcID:svcid})
	connBySvcID[svcid] = ses
	connBySvcNameGuard.Unlock()
	log.Infof("remote service added: '%s' sid: %d", svcid, ses.ID())
}

func RemoveRemoteService(ses cellnet.Session) {
	if ses == nil {
		return
	}
	ctx := SessionToContext(ses)
	if ctx != nil {
		if removeNotify != nil {
			removeNotify(ctx, ses)
		}
		connBySvcNameGuard.Lock()
		delete(connBySvcID, ctx.SvcID)
		connBySvcNameGuard.Unlock()
		log.Infof("remote service removed '%s' sid: %d", ctx.SvcID, ses.ID())
	} else {
		log.Infof("remote service removed sid: %d, context lost", ses.ID())
	}
}

func SetRemoteServiceNotify(mode string, callback NotifyFunc) {
	switch mode {
	case "remove":
		removeNotify = callback
	default:
		panic("unknown notify mode")
	}
}

func SessionToContext(ses cellnet.Session) *RemoteServiceContext {
	if ses == nil {
		return nil
	}
	if raw, ok := ses.(cellnet.ContextSet).GetContext("ctx"); ok {
		return raw.(*RemoteServiceContext)
	}
	return nil
}

func GetRemoteService(svcid string) cellnet.Session {
	connBySvcNameGuard.RLock()
	defer connBySvcNameGuard.RUnlock()
	if ses, ok := connBySvcID[svcid]; ok {
		return ses
	}
	return nil
}

func VisitRemoteService(callback func(ses cellnet.Session, ctx *RemoteServiceContext) bool) {
	connBySvcNameGuard.RLock()
	defer connBySvcNameGuard.RUnlock()
	for _, ses := range connBySvcID {
		if !callback(ses, SessionToContext(ses)) {
			break
		}
	}
}