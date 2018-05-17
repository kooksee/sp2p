package sp2p

import (
	"github.com/kooksee/uspnet/common"
)

type IHandler func(*SP2p, *KMsg)

type IP2p interface {
	LoadSeeds([]string) error
	Start()
	Write(*KMsg) error
}

type ITable interface {
	// 路由表大小
	Size() int
	// 获得节点列表,把节点列表转换为[enode://<hex node id>@10.3.58.6:30303?discport=30301]的方式
	GetRawNodes() []string
	// 添加节点
	AddNode(*Node)
	// 更新节点
	UpdateNode(*Node)
	// 删除节点
	DeleteNode(common.Hash)
	// 随机得到路由表中的n个节点
	FindRandomNodes(int) []*Node
	// 查找距离最近的n个节点
	FindMinDisNodes(common.Hash, int) []*Node
	// 查找目标相比本节点更近的节点
	FindNodeWithTargetBySelf(common.Hash) []*Node
	// 查找目标相比另一个节点的更近的节点
	FindNodeWithTarget(common.Hash, common.Hash) []*Node
}