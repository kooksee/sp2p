package sp2p

type IHandler func(*sp2p, *KMsg)

type IMessage interface {
	// 获取类型
	T() byte
	// 描述信息
	String() string
	// 业务处理
	OnHandle(ISP2P, *KMsg)
}

type ITable interface {
	// 路由表大小
	size() int
	// 获得节点列表,把节点列表转换为[sp2p://<hex node id>@10.3.58.6:30303?discport=30301]的方式
	getRawNodes() []string
	// 添加节点
	addNode(*node)
	// 更新节点
	updateNode(*node)
	// 删除节点
	deleteNode(Hash)
	// 随机得到路由表中的n个节点
	findRandomNodes(int) []*node
	// 查找距离最近的n个节点
	findMinDisNodes(Hash, int) []*node
	// 查找目标相比本节点更近的节点
	findNodeWithTargetBySelf(Hash) []*node
	// 查找目标相比另一个节点的更近的节点
	findNodeWithTarget(Hash, Hash) []*node
}

type ISP2P interface {
	GetAddr() string
	Write(msg *KMsg)
	GetSelfNode() string
	GetNodes() []string
	TableSize() int
	UpdateNode(rawUrl string) error
	AddNode(rawUrl string) error
	FindMinDisNodes(targetID string, n int) (nodes []string, err error)
	FindRandomNodes(n int) (nodes []string)
	FindNodeWithTargetBySelf(d string) (nodes []string)
	FindNodeWithTarget(targetId string, measure string) (nodes []string)
	Broadcast(msg *KMsg)
	PingN()
	FindN()
}
