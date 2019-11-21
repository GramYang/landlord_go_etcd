package discovery

var (
	Default Discovery
)

type Discovery interface {
	//注册服务
	Register(*ServiceDesc) error

	//解注册服务
	Deregister(svcid string) error

	//查询注册服务
	Query(name string) (ret []*ServiceDesc)

	//注册服务变化通知
	RegisterNotify(mode string) (ret chan struct{})

	//解除服务变化通知
	DeregisterNotify(mode string, c chan struct{})

	// 设置值
	SetValue(key string, value interface{}) error

	// 取值返回
	GetValue(key string) (string, error)

	// 删除值
	DeleteValue(key string) error

	//主动关闭
	Close() error
}