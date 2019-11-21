package frontend

import (
	"errors"
	"landlord_go/service"
	"landlord_go/srv/agent/model"
)

var (
	ErrAlreadyBind = errors.New("already bind user")
	ErrBackendServerNotFound = errors.New("backend svc not found")
	ErrBackendSDNotFound = errors.New("backend sd not found")
)

//客户端绑定后台服务，其实是向user录入信息顺便确认一下agent是否持有后台服务的session
func bindClientToBackend(backendSvcID string, clientSesID int64) (*agentmodel.User, error) {
	backendSes := service.GetRemoteService(backendSvcID)
	if backendSes == nil {
		return nil, ErrBackendServerNotFound
	}
	sd := service.SessionToContext(backendSes)
	if sd == nil {
		return nil, ErrBackendSDNotFound
	}
	clientSes := agentmodel.GetClientSession(clientSesID)
	u := agentmodel.SessionToUser(clientSes)
	if u != nil {
		return nil, ErrAlreadyBind
	}
	u = agentmodel.CreateUser(clientSes)
	u.SetBackend(sd.Name, sd.SvcID)
	return u, nil
}