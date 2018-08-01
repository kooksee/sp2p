package sp2p

type IMessage interface {
	// 获取类型
	T() byte
	// 描述信息
	String() string
	// 业务处理
	OnHandle(ISP2P, *KMsg) error
}

type ITable interface {
	// 路由表大小
	size() int
	// 获得节点列表,把节点列表转换为[sp2p://<hex node id>@10.3.58.6:30303?discport=30301]的方式
	getRawNodes() []string
	// 添加节点
	addNode(*node)error
	// 更新节点
	updateNode(*node)error
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
	Write(msg *KMsg)
	Broadcast(msg *KMsg)
	RandomCast(msg *KMsg)
	InitSeeds(seeds []string)

	SelfNode() string
	NodeDumps() []string
	NodeUpdate(rawUrl ... string) error
	NodeDel(nodeID ... string) error

	FindMinDisNodes(targetID string, n int) (nodes []string, err error)
	FindRandomNodes(n int) (nodes []string)
	FindNodeWithTarget(targetID string, measure string) (nodes []string)
	FindNodeWithTargetBySelf(targetID string) (nodes []string)

	PingRandom()
	FindRandom()
}
