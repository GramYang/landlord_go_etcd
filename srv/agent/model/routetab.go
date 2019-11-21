package agentmodel

import (
	"landlord_go/table"
	"sync"
)

const (
	ConfigPath = "config_demo/route_rule"
)

var (
	// 消息名映射路由规则
	ruleByMsgName      = map[string]*table.RouteRule{}
	ruleByMsgNameGuard sync.RWMutex

	ruleByMsgID = map[int]*table.RouteRule{}
)

// 消息名取路由规则
func GetTargetService(msgName string) *table.RouteRule {

	ruleByMsgNameGuard.RLock()
	defer ruleByMsgNameGuard.RUnlock()

	if rule, ok := ruleByMsgName[msgName]; ok {
		return rule
	}

	return nil
}

func GetRuleByMsgID(msgid int) *table.RouteRule {
	ruleByMsgNameGuard.RLock()
	defer ruleByMsgNameGuard.RUnlock()

	if rule, ok := ruleByMsgID[msgid]; ok {
		return rule
	}

	return nil
}

// 清除所有规则
func ClearRule() {

	ruleByMsgNameGuard.Lock()
	ruleByMsgName = map[string]*table.RouteRule{}
	ruleByMsgID = map[int]*table.RouteRule{}
	ruleByMsgNameGuard.Unlock()
}

// 添加路由规则
func AddRouteRule(rule *table.RouteRule) {

	ruleByMsgNameGuard.Lock()
	ruleByMsgName[rule.MsgName] = rule
	if rule.MsgID == 0 {
		panic("RouteRule msgid = 0, run MakeProto.sh please!")
	}
	ruleByMsgID[rule.MsgID] = rule
	ruleByMsgNameGuard.Unlock()
}