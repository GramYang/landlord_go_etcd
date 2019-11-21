package hubmodel

import "time"
//现阶段，先用srvname来当成srvid使用，后期有需要再扩展
type Status struct {
	UserCount int32
	SvcID string
	LastUpdate time.Time
}

var (
	statusBySrvID = map[string]*Status{}
)

func UpdateStatus(nowStatus *Status) *Status {
	status, _ := statusBySrvID[nowStatus.SvcID]
	if status == nil {
		status = nowStatus
		statusBySrvID[nowStatus.SvcID] = status
	}
	status.UserCount = nowStatus.UserCount
	status.LastUpdate = time.Now()
	return status
}

func RemoveStatus(srvID string) {
	delete(statusBySrvID, srvID)
}

func VisitStatus(callback func(status *Status) bool) {
	for _, s := range statusBySrvID {
		if !callback(s) {
			break
		}
	}
}