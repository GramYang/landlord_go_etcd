package service

var (
	svcName  string
	//etcdv3可以用多个地址
	discoveryAddr []string
)

// 获取当前服务进程名称
func GetSvcName() string {
	return svcName
}

func SetDiscoveryAddrs(addrs ...string) {
	discoveryAddr = append(discoveryAddr, addrs...)
}