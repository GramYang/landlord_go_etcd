package routerule

import (
	"encoding/json"
	"errors"
	"landlord_go/discovery"
	"landlord_go/srv/agent/model"
	"landlord_go/table"
)

//从服务发现下载路由规则
func Download() error {
	log.Traceln("download route rule from discovery...")
	var tab table.RouteTable
	res, err := discovery.Default.GetValue(agentmodel.ConfigPath)
	if err != nil {
		return errors.New("value not exists")
	}
	err1 := json.Unmarshal([]byte(res), &tab)
	if err1 != nil {
		return err
	}
	agentmodel.ClearRule()
	for _, r := range tab.Rule {
		agentmodel.AddRouteRule(r)
	}
	log.Tracef("Total %d rules added\n", len(tab.Rule))
	return nil
}

//不从服务发现下载路由规则，直接写到代码中去
func GetRouteRule() {
	log.Traceln("get route rule...")
	r1 := &table.RouteRule{MsgName:"JsonREQ", SvcName:"game", Mode:"auth", MsgID:20000}
	agentmodel.AddRouteRule(r1)
	r2 := &table.RouteRule{MsgName:"VerifyREQ", SvcName:"game", Mode:"pass", MsgID:13457}
	agentmodel.AddRouteRule(r2)
}