package agentmodel

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
)

var (
	FrontendSessionManager peer.SessionManager
	AgentSvcID string
)

func GetClientSession(sesid int64) cellnet.Session {
	return FrontendSessionManager.GetSession(sesid)
}

func GetUser(sesid int64) *User {
	return SessionToUser(GetClientSession(sesid))
}

//创建一个User并与session绑定
func CreateUser(clientSes cellnet.Session) *User {
	u := NewUser(clientSes)
	clientSes.(cellnet.ContextSet).SetContext("user", u)
	return u
}

//获取session的User
func SessionToUser(clientSes cellnet.Session) *User {
	if clientSes == nil {
		return nil
	}
	if raw, ok := clientSes.(cellnet.ContextSet).GetContext("user"); ok {
		return raw.(*User)
	}
	return nil
}

//遍历agent持有的所有用户，即所有的client
func VisitUser(callback func(*User) bool) {
	FrontendSessionManager.VisitSession(func(clientSes cellnet.Session) bool {
		if u := SessionToUser(clientSes); u != nil {
			return callback(u)
		}
		return true
	})
}