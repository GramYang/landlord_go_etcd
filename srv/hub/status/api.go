package hubstatus

import (
	"landlord_go/service"
	"landlord_go/srv/hub/model"
	"sort"
)

func SelectServiceByLowUserCount(svcName, svcGroup string, mustConnected bool) (finalSvcID string) {
	var statusList []*hubmodel.Status
	hubmodel.VisitStatus(func(status *hubmodel.Status) bool {
		name, _, group, err := service.ParseSvcID(status.SvcID)
		if err != nil {
			return true
		}
		if name != svcName {
			return true
		}
		if svcGroup == "" || svcGroup == group {
			if !mustConnected || service.GetRemoteService(status.SvcID) != nil {
				statusList = append(statusList, status)
			}
		}
		return true
	})
	total := len(statusList)
	switch total {
	case 0:
		return ""
	case 1:
		return statusList[0].SvcID
	default:
		sort.Slice(statusList, func(i, j int) bool {
			a := statusList[i]
			b := statusList[j]
			return a.UserCount < b.UserCount
		})
		finalSvcID = statusList[0].SvcID
	}
	return
}