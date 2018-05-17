package sp2p

import (
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/kooksee/uspnet/common"
)

type Table struct {
	ITable

	mutex sync.Mutex // protects buckets, their content, and nursery

	buckets  [cfg.NBuckets]*bucket
	selfNode *Node //info of local node
}

func newTable(id NodeID, addr *net.UDPAddr) *Table {

	table := &Table{
		selfNode: NewNode(id, addr.IP, uint16(addr.Port)),
	}

	for i := 0; i < cfg.NBuckets; i++ {
		table.buckets[i] = newBuckets()
	}

	return table
}

func (t *Table) GetNode() *Node {
	return t.selfNode
}

func (t *Table) GetAllNodes() []*Node {
	nodes := make([]*Node, 0)
	for _, b := range t.buckets {
		b.peers.Each(func(index int, value interface{}) {
			nodes = append(nodes, value.(*Node))
		})
	}
	return nodes
}

func (t *Table) GetRawNodes() []string {
	nodes := make([]string, 0)
	for _, n := range t.GetAllNodes() {
		nodes = append(nodes, n.String())
	}
	return nodes
}

func (t *Table) AddNode(node *Node) {
	t.buckets[logdist(t.selfNode.sha, node.sha)].addNodes(node)
}

func (t *Table) UpdateNode(node *Node) {
	t.buckets[logdist(t.selfNode.sha, node.sha)].updateNodes(node)
}

func (t *Table) Size() int {
	n := 0
	for _, b := range t.buckets {
		n += b.peers.Size()
	}
	return n
}

// ReadRandomNodes fills the given slice with random nodes from the
// table. It will not write the same node more than once. The nodes in
// the slice are copies and can be modified by the caller.
func (t *Table) FindRandomNodes(n int) []*Node {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if n > cfg.NBuckets {
		n = cfg.NBuckets
	} else if n < 1 {
		n = 5
	}

	nodes := make([]*Node, 0)
	for _, b := range t.buckets {
		b.peers.Each(func(_ int, value interface{}) {
			nodes = append(nodes, value.(*Node))
		})
	}

	if len(nodes) < n {
		return nodes
	}

	nodeSet := hashset.New()
	rand.Seed(time.Now().Unix())
	k := int32(len(nodes))
	for nodeSet.Size() < n {
		nodeSet.Add(nodes[rand.Int31n(k)])
	}

	rnodes := make([]*Node, 0)
	for _, n := range nodeSet.Values() {
		rnodes = append(rnodes, n.(*Node))
	}
	return rnodes
}

// findNodeWithTarget find nodes that distance of target is less than measure with target
func (t *Table) FindNodeWithTarget(target common.Hash, measure common.Hash) []*Node {
	minDis := make([]*Node, 0)
	for _, n := range t.FindMinDisNodes(target, cfg.NodeResponseNumber) {
		if distCmp(target, n.sha, measure) < 0 {
			minDis = append(minDis, n)
		}
	}

	return minDis
}

func (t *Table) FindNodeWithTargetBySelf(target common.Hash) []*Node {
	return t.FindNodeWithTarget(target, t.selfNode.sha)
}

func (t *Table) DeleteNode(target common.Hash) {
	t.buckets[logdist(t.selfNode.sha, target)].deleteNodes(target)
}

func (t *Table) FindMinDisNodes(target common.Hash, number int) []*Node {
	if number > cfg.NBuckets {
		number = cfg.NBuckets
	} else if number < 1 {
		number = 5
	}

	result := &nodesByDistance{
		target:   target,
		maxElems: number,
		entries:  make([]*Node, 0),
	}

	for _, b := range t.buckets {
		b.peers.Each(func(_ int, value interface{}) {
			result.push(value.(*Node))
		})
	}

	return result.entries
}

// nodesByDistance is a list of nodes, ordered by
// distance to to.
type nodesByDistance struct {
	entries  []*Node
	target   common.Hash
	maxElems int
}

// push adds the given node to the list, keeping the total size below maxElems.
func (h *nodesByDistance) push(n *Node) {
	ix := sort.Search(len(h.entries), func(i int) bool {
		return distCmp(h.target, h.entries[i].sha, n.sha) > 0
	})
	if len(h.entries) < h.maxElems {
		h.entries = append(h.entries, n)
	}
	if ix == len(h.entries) {
		// farther away than all nodes we already have.
		// if there was room for it, the node is now the last element.
	} else {
		// slide existing entries down to make room
		// this will overwrite the entry we just appended.
		copy(h.entries[ix+1:], h.entries[ix:])
		h.entries[ix] = n
	}
}
